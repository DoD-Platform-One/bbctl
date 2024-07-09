package util

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
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

// Factory is an interface providing initialization methods for various external clients
type Factory interface {
	GetAWSClient() bbAws.Client
	GetHelmClient(cmd *cobra.Command, namespace string) (helm.Client, error)
	GetK8sClientset(cmd *cobra.Command) (kubernetes.Interface, error)
	GetLoggingClient() bbLog.Client                              // this can't bubble up an error, if it fails it will panic
	GetLoggingClientWithLogger(logger *slog.Logger) bbLog.Client // this can't bubble up an error, if it fails it will panic
	GetRuntimeClient(*runtime.Scheme) (runtimeClient.Client, error)
	GetK8sDynamicClient(cmd *cobra.Command) (dynamic.Interface, error)
	GetRestConfig(cmd *cobra.Command) (*rest.Config, error)
	GetCommandExecutor(cmd *cobra.Command, pod *coreV1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remoteCommand.Executor, error)
	GetCredentialHelper() func(string, string) string
	GetCommandWrapper(name string, args ...string) *bbUtilApiWrappers.Command
	GetIstioClientSet(cfg *rest.Config) (bbUtilApiWrappers.IstioClientset, error)
	GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error)
	GetViper() *viper.Viper
}

// NewFactory initializes and returns a new instance of UtilityFactory
func NewFactory() *UtilityFactory {
	return &UtilityFactory{
		viperInstance: viper.New(),
	}
}

// UtilityFactory is a concrete implementation of the Factory interface containing a pre-initialized Viper instance
type UtilityFactory struct {
	viperInstance *viper.Viper
}

// CredentialsFile struct represents credentials YAML files with a top level field called `credentials`
// which contains a list of Credentials struct values
type CredentialsFile struct {
	Credentials []Credentials `yaml:"credentials"`
}

// Credentials struct represents each individual entry in a valid credentials file.
// URI should be unique for every entry
type Credentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	URI      string `yaml:"uri"`
}

// ReadCredentialsFile reads the credentials file for the requested uri and attempts to return the string value
// for the requested component
// Multiple credentials can be present in the file and are identified by their `uri` field
//
// Credentials file path is pulled from the bbctl config and will default to ~/.bbctl/credentials.yaml when not set
//
// # Valid component values are `username` and `password`, any other value will panic
//
// Panics when the credentials file cannot be accessed or parsed, component value is invalid,
// and when uri is not found in credentials list
func (f *UtilityFactory) ReadCredentialsFile(component string, uri string) string {
	// Get credentials path
	loggingClient := f.GetLoggingClient()
	configClient, err := f.GetConfigClient(nil)
	loggingClient.HandleError("Unable to get config client: %v", err)
	config := configClient.GetConfig()
	credentialsPath := config.UtilCredentialHelperConfiguration.FilePath
	if credentialsPath == "" {
		// Get the home directory
		homeDir, err := os.UserHomeDir()
		loggingClient.HandleError("Unable to get home directory: %v", err)
		credentialsPath = path.Join(homeDir, ".bbctl", "credentials.yaml")
	}

	// Read the credentials file
	credentialsYaml, err := os.ReadFile(credentialsPath)
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
		loggingClient.Error(fmt.Sprintf("No credentials found for %v in %v", uri, credentialsPath))
	}

	// Return the requested component
	if component == "username" {
		return credentials.Username
	} else if component == "password" {
		return credentials.Password
	} else {
		// this will panic
		loggingClient.Error(fmt.Sprintf("Invalid component %v", component))
		return ""
	}
}

// GetCredentialHelper returns a function reference to the configured credential helper function that can
// be called to fetch credential values. A custom credential helper function is any CLI executable
// script which can be passed into this function via the bbctl config settings as a file path.
//
// Credential helper functions take 2 parameters. For the default `credentials-file` implementation, these parameters are:
//
// * component (string) - The Credentials struct field name, either `username` or `password`
//
// * uri (string) - The Credentials struct URI value which uniquely identifies the requested component
//
// These parameters are passed into custom credential helpers as CLI arguments in the same order.
//
// Panics when no credential helper is defined, there is an issue reading credentials from a file,
// there is an issue running a custom credential helper script, and when an empty value is returned
// for a requested credential component
func (f *UtilityFactory) GetCredentialHelper() func(string, string) string {
	return func(component string, uri string) string {
		loggingClient := f.GetLoggingClient()
		configClient, err := f.GetConfigClient(nil)
		loggingClient.HandleError("Unable to get config client: %v", err)
		config := configClient.GetConfig()
		helper := config.UtilCredentialHelperConfiguration.CredentialHelper
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
			loggingClient.Error(fmt.Sprintf("No %v found for %v in %v", component, uri, helper))
		}
		return output
	}
}

// GetAWSClient initializes and returns a new AWS API client
//
// Panics when there are issues with the bbctl configurations
func (f *UtilityFactory) GetAWSClient() bbAws.Client {
	loggingClient := f.GetLoggingClient()
	clientGetter := bbAws.ClientGetter{}
	client, err := clientGetter.GetClient(loggingClient)
	loggingClient.HandleError("Unable to get AWS client: %v", err)
	return client
}

// GetHelmClient initializes and returns a new Helm client that can perform operations in the given namespace
//
// # Returns a nil client and an error if there are any issues with the intialization
//
// Panics when there are issues with the bbctl configurations
func (f *UtilityFactory) GetHelmClient(cmd *cobra.Command, namespace string) (helm.Client, error) {
	actionConfig, err := f.getHelmConfig(cmd, namespace)
	if err != nil {
		return nil, err
	}

	getReleaseClient := action.NewGet(actionConfig)

	getListClient := action.NewList(actionConfig)

	getValuesClient := action.NewGetValues(actionConfig)
	getValuesClient.AllValues = true

	return helm.NewClient(getReleaseClient.Run, getListClient.Run, getValuesClient.Run)
}

// GetK8sClientset initializes and returns a new k8s client by calling the kubernetes.NewForConfig()
// function with the REST configuration defined in the bbctl config layered together with any existing
// KUBECONFIG settings and k8s config CLI parameters
//
// # Returns a nil client and an error if there are any issues with the intialization
//
// Panics when there are issues with the bbctl configurations
func (f *UtilityFactory) GetK8sClientset(cmd *cobra.Command) (kubernetes.Interface, error) {
	config, err := f.GetRestConfig(cmd)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// GetK8sDynamicClient initializes and returns a new dynamic k8s client by calling the dynamic.NewForConfig()
// function with the configuration defined in the bbctl config layered together with any existing
// KUBECONFIG settings and k8s config CLI parameters
//
// # Returns a nil client and an error if there are any issues with the intialization
//
// Panics when there are issues with the bbctl configurations
func (f *UtilityFactory) GetK8sDynamicClient(cmd *cobra.Command) (dynamic.Interface, error) {
	configClient, err := f.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	config := configClient.GetConfig()
	return bbK8sUtil.BuildDynamicClient(config)
}

// GetLoggingClient initializes and returns a new logging client using the default slog logger implementation
//
// Panics when there are issues initializing the logger
func (f *UtilityFactory) GetLoggingClient() bbLog.Client {
	return f.GetLoggingClientWithLogger(nil)
}

// GetLoggingClientWithLogger initializes and returns a new logging client using the given logger implementation
//
// Panics when there are issues initializing the logger
func (f *UtilityFactory) GetLoggingClientWithLogger(logger *slog.Logger) bbLog.Client {
	clientGetter := bbLog.ClientGetter{}
	client := clientGetter.GetClient(logger)
	return client
}

// GetRuntimeClient initializes and returns a new k8s runtime client by calling the client.New() function
//
// # Returns a nil client and an error if there are any issues with the intialization
//
// Panics when there are issues creating the k8s REST config
func (f *UtilityFactory) GetRuntimeClient(scheme *runtime.Scheme) (runtimeClient.Client, error) {
	// init runtime controller client
	runtimeClient, err := runtimeClient.New(ctrl.GetConfigOrDie(), runtimeClient.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return runtimeClient, err
}

// GetRestConfig returns the k8s REST configuration defined in the bbctl config layered together with any existing
// KUBECONFIG settings and k8s config CLI parameters
//
// # Returns a nil client and an error if there are any issues with the intialization
func (f *UtilityFactory) GetRestConfig(cmd *cobra.Command) (*rest.Config, error) {
	configClient, err := f.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	config := configClient.GetConfig()
	return bbK8sUtil.BuildKubeConfig(config)
}

// GetCommandExecutor initializes and returns a new SPDY executor that can run the given command in a Pod in the k8s cluster
//
// # Returns a nil executor and an error if there are any issues with the intialization
func (f *UtilityFactory) GetCommandExecutor(cmd *cobra.Command, pod *coreV1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remoteCommand.Executor, error) {
	client, err := f.GetK8sClientset(cmd)
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

	// REST config is already validated in the f.GetK8sClientset(cmd) call above
	config, _ := f.GetRestConfig(cmd)

	return remoteCommand.NewSPDYExecutor(config, "POST", req.URL())
}

// Internal helper function to create configs for GetHelmClient
//
// Panics if Helm action.Configuration.Init fails
func (f *UtilityFactory) getHelmConfig(cmd *cobra.Command, namespace string) (*action.Configuration, error) {
	configClient, err := f.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	bbctlConfig := configClient.GetConfig()

	loggingClient := f.GetLoggingClient()
	config, err := bbK8sUtil.BuildKubeConfig(bbctlConfig)
	if err != nil {
		return nil, err
	}

	// TODO: add support for an alternate warning handler and then just default nil
	clientGetter := helm.NewRESTClientGetter(config, namespace, nil)

	debugLog := func(format string, v ...interface{}) {
		loggingClient.Debug(format, v...)
	}

	// The actionConfig.Init method will either panic or return nil
	// It cannot return an error value like the return type says
	actionConfig := new(action.Configuration)
	actionConfig.Init(
		clientGetter,
		namespace,
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)

	return actionConfig, nil
}

// GetCommandWrapper initializes and returns a new Command instance which encapsulates the functionality needed to run a CLI command
// `name` is the command to execute i.e. kubectl
// `args` string values are all passed to the command as CLI arguments
func (f *UtilityFactory) GetCommandWrapper(name string, args ...string) *bbUtilApiWrappers.Command {
	return bbUtilApiWrappers.NewExecRunner(name, args...)
}

// GetIstioClientSet initializes and returns a new istio client set by calling versioned.NewForConfig() with the provided REST config settings
//
// # Returns a nil client and an error if there are any issues with the intialization
func (f *UtilityFactory) GetIstioClientSet(cfg *rest.Config) (bbUtilApiWrappers.IstioClientset, error) {
	clientSet, err := versioned.NewForConfig(cfg)
	return clientSet, err
}

// GetConfigClient initializes and returns a new bbctl config client
//
// # Returns a nil client and an error if there are any issues with the intialization
func (f *UtilityFactory) GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error) {
	clientGetter := bbConfig.ClientGetter{}
	loggingClient := f.GetLoggingClient()
	client, err := clientGetter.GetClient(command, &loggingClient, f.GetViper())
	if err != nil {
		return nil, err
	}
	return client, nil
}

// GetViper returns the viper instance
func (f *UtilityFactory) GetViper() *viper.Viper {
	return f.viperInstance
}
