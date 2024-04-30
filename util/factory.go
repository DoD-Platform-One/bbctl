package util

import (
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
	pFlag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"istio.io/client-go/pkg/clientset/versioned"

	bbUtilApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/apiwrappers"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbK8sUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	remoteCommand "k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Factory interface
type Factory interface {
	GetAWSClient() bbAws.Client
	GetHelmClient(namespace string) (helm.Client, error)
	GetK8sClientset() (kubernetes.Interface, error)
	GetLoggingClient() bbLog.Client                              // this can't bubble up an error, if it fails it will panic
	GetLoggingClientWithLogger(logger *slog.Logger) bbLog.Client // this can't bubble up an error, if it fails it will panic
	GetRuntimeClient(*runtime.Scheme) (runtimeClient.Client, error)
	GetK8sDynamicClient() (dynamic.Interface, error)
	GetRestConfig() (*rest.Config, error)
	GetCommandExecutor(pod *coreV1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remoteCommand.Executor, error)
	GetCredentialHelper() func(string, string) string
	GetCommandFlags() *pFlag.FlagSet
	GetCommandWrapper(name string, args ...string) *bbUtilApiWrappers.Command
	GetIstioClientSet(cfg *rest.Config) (bbUtilApiWrappers.IstioClientset, error)
	GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error)
	GetViper() *viper.Viper
}

// NewFactory - new factory method
func NewFactory(flags *pFlag.FlagSet) *UtilityFactory {
	return &UtilityFactory{flags: flags}
}

// UtilityFactory - util factory
type UtilityFactory struct {
	flags *pFlag.FlagSet
}

// GetCommandFlags - get the @*#%&# command line flags
func (f *UtilityFactory) GetCommandFlags() *pFlag.FlagSet {
	return f.flags
}

// CredentialsFile struct
type CredentialsFile struct {
	Credentials []Credentials `yaml:"credentials"`
}

// Credentials Struct
type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	URI      string `yaml:"uri"`
}

// ReadCredentialsFile - read credentials file
func (f *UtilityFactory) ReadCredentialsFile(component string, uri string) string {
	// Get credentials path
	credentialsPath := viper.GetString("big-bang-credential-helper-credentials-file-path")
	if credentialsPath == "" {
		// Get the home directory
		homeDir, err := os.UserHomeDir()
		f.GetLoggingClient().HandleError("Unable to get home directory: %v", err)
		credentialsPath = path.Join(homeDir, ".bbctl", "credentials.yaml")
	}

	// Read the credentials file
	credentialsYaml, err := os.ReadFile(credentialsPath)
	loggingClient := f.GetLoggingClient()
	loggingClient.HandleError("Unable to read credentials file %v: %v", err, credentialsPath)

	// Unmarshal the credentials file
	var credentialsFile CredentialsFile
	err = yaml.Unmarshal(credentialsYaml, &credentialsFile)
	loggingClient.HandleError("Unable to unmarshal credentials file %v: %v", err, credentialsPath)

	// Find the credentials for the uri
	credentials := Credentials{}
	for _, c := range credentialsFile.Credentials {
		if c.URI == uri {
			credentials = c
			break
		}
	}
	if credentials.URI == "" {
		loggingClient.Error("No credentials found for %v in %v", uri, credentialsPath)
	}

	// Return the requested component
	if component == "username" {
		return credentials.Username
	} else if component == "password" {
		return credentials.Password
	} else {
		// this will panic
		loggingClient.Error("Invalid component %v", component)
		return ""
	}
}

// GetCredentialHelper - get the credential helper
func (f *UtilityFactory) GetCredentialHelper() func(string, string) string {
	return func(component string, uri string) string {
		loggingClient := f.GetLoggingClient()
		helper := viper.GetString("big-bang-credential-helper")
		if helper == "" {
			loggingClient.Error("No credential helper defined (\"big-bang-credential-helper\")")
		}
		output := ""
		if helper == "credentials-file" {
			output = f.ReadCredentialsFile(component, uri)
		} else {
			cmd := exec.Command(helper, component, uri)
			rawOutput, err := cmd.Output()
			loggingClient.HandleError("Unable to get %v for %v from %v: %v", err, component, uri, helper)
			output = string(rawOutput[:])
		}
		if output == "" {
			loggingClient.Error("No %v found for %v in %v", component, uri, helper)
		}
		return output
	}
}

// GetAWSClient - get aws client
func (f *UtilityFactory) GetAWSClient() bbAws.Client {
	loggingClient := f.GetLoggingClient()
	clientGetter := bbAws.ClientGetter{}
	client, err := clientGetter.GetClient(loggingClient)
	loggingClient.HandleError("Unable to get AWS client: %v", err)
	return client
}

// GetHelmClient - get helm client
func (f *UtilityFactory) GetHelmClient(namespace string) (helm.Client, error) {
	actionConfig, err := f.getHelmConfig(namespace)
	if err != nil {
		return nil, err
	}

	getReleaseClient := action.NewGet(actionConfig)

	getListClient := action.NewList(actionConfig)

	getValuesClient := action.NewGetValues(actionConfig)
	getValuesClient.AllValues = true

	return helm.NewClient(getReleaseClient.Run, getListClient.Run, getValuesClient.Run)
}

// GetK8sClientset - get k8s clientset
func (f *UtilityFactory) GetK8sClientset() (kubernetes.Interface, error) {
	config, err := f.GetRestConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// GetK8sDynamicClient - get k8s dynamic client
func (f *UtilityFactory) GetK8sDynamicClient() (dynamic.Interface, error) {
	return bbK8sUtil.BuildDynamicClientFromFlags(f.flags)
}

// GetLoggingClient - get logging client
func (f *UtilityFactory) GetLoggingClient() bbLog.Client {
	return f.GetLoggingClientWithLogger(nil)
}

// GetLoggingClientWithLogger - get logging client providing logger
func (f *UtilityFactory) GetLoggingClientWithLogger(logger *slog.Logger) bbLog.Client {
	clientGetter := bbLog.ClientGetter{}
	client := clientGetter.GetClient(logger)
	return client
}

// GetRuntimeClient - get runtime client
func (f *UtilityFactory) GetRuntimeClient(scheme *runtime.Scheme) (runtimeClient.Client, error) {
	// init runtime controller client
	runtimeClient, err := runtimeClient.New(ctrl.GetConfigOrDie(), runtimeClient.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return runtimeClient, err
}

// GetRestConfig - get rest config
func (f *UtilityFactory) GetRestConfig() (*rest.Config, error) {
	return bbK8sUtil.BuildKubeConfigFromFlags(f.flags)
}

// GetCommandExecutor - get executor to run command in a Pod
func (f *UtilityFactory) GetCommandExecutor(pod *coreV1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remoteCommand.Executor, error) {
	client, err := f.GetK8sClientset()
	if err != nil {
		return nil, err
	}

	req := client.Discovery().RESTClient().Post().
		Prefix("/api/v1").
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")

	req.SpecificallyVersionedParams(&coreV1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     false,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       false,
	}, scheme.ParameterCodec, schema.GroupVersion{Version: "v1"})

	config, err := f.GetRestConfig()
	if err != nil {
		return nil, err
	}

	return remoteCommand.NewSPDYExecutor(config, "POST", req.URL())
}

func (f *UtilityFactory) getHelmConfig(namespace string) (*action.Configuration, error) {
	loggingClient := f.GetLoggingClient()
	config, err := bbK8sUtil.BuildKubeConfigFromFlags(f.flags)
	if err != nil {
		return nil, err
	}

	// TODO: add support for an alternate warning handler and then just default nil
	clientGetter := helm.NewRESTClientGetter(config, namespace, nil)

	debugLog := func(format string, v ...interface{}) {
		loggingClient.Debug(format, v...)
	}

	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		clientGetter,
		namespace,
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)
	if err != nil {
		return nil, err
	}

	return actionConfig, nil
}

// GetCommandWrapper - get command wrapper
func (f *UtilityFactory) GetCommandWrapper(name string, args ...string) *bbUtilApiWrappers.Command {
	return bbUtilApiWrappers.NewExecRunner(name, args...)
}

// GetIstioClientSet - get istio client set
func (f *UtilityFactory) GetIstioClientSet(cfg *rest.Config) (bbUtilApiWrappers.IstioClientset, error) {
	clientSet, err := versioned.NewForConfig(cfg)
	return clientSet, err
}

// GetConfigClient - get config client
func (f *UtilityFactory) GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error) {
	clientGetter := bbConfig.ClientGetter{}
	loggingClient := f.GetLoggingClient()
	client, err := clientGetter.GetClient(command, &loggingClient)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// GetViper returns the viper instance.
func (f *UtilityFactory) GetViper() *viper.Viper {
	return viper.GetViper()
}
