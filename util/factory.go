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
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"istio.io/client-go/pkg/clientset/versioned"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	bbUtilApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/apiwrappers"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/common_interfaces"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbK8sUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	bbOutput "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"

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
	GetAWSClient() (bbAws.Client, error)
	GetGitLabClient() (bbGitLab.Client, error)
	GetHelmClient(cmd *cobra.Command, namespace string) (helm.Client, error)
	GetK8sClientset(cmd *cobra.Command) (kubernetes.Interface, error)
	GetLoggingClient() (bbLog.Client, error)
	GetLoggingClientWithLogger(logger *slog.Logger) (bbLog.Client, error)
	GetRuntimeClient(*runtime.Scheme) (runtimeClient.Client, error)
	GetK8sDynamicClient(cmd *cobra.Command) (dynamic.Interface, error)
	GetOutputClient(cmd *cobra.Command) (bbOutput.Client, error)
	GetRestConfig(cmd *cobra.Command) (*rest.Config, error)
	GetCommandExecutor(
		cmd *cobra.Command,
		pod *coreV1.Pod,
		container string,
		command []string,
		stdout io.Writer,
		stderr io.Writer,
	) (remoteCommand.Executor, error)
	GetCredentialHelper() (CredentialHelper, error)
	GetCommandWrapper(name string, args ...string) (*bbUtilApiWrappers.Command, error)
	GetIstioClientSet(cfg *rest.Config) (bbUtilApiWrappers.IstioClientset, error)
	GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error)
	GetViper() (*viper.Viper, error)
	GetIOStream() (*genericIOOptions.IOStreams, error)
	GetPipe() (commonInterfaces.FileLike, commonInterfaces.FileLike, error)
}

// NewFactory initializes and returns a new instance of UtilityFactory
func NewFactory(referenceFactory Factory) *UtilityFactory {
	factory := &UtilityFactory{
		referenceFactory: referenceFactory,
	}
	if factory.referenceFactory == nil {
		factory.referenceFactory = factory
	}
	factory.getViperFunction = getViper
	return factory
}

// UtilityFactory is a concrete implementation of the Factory interface containing a pre-initialized Viper instance
type UtilityFactory struct {
	referenceFactory Factory
	getViperFunction func() (*viper.Viper, error)
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
// # Valid component values are `username` and `password`, any other value will return an error
//
// Errors when the credentials file cannot be accessed or parsed, component value is invalid,
// and when uri is not found in credentials list
func (f *UtilityFactory) ReadCredentialsFile(component string, uri string) (string, error) {
	return f.readCredentialsFile(component, uri, yaml.Unmarshal)
}

// See ReadCredentialsFile
func (f *UtilityFactory) readCredentialsFile(component string, uri string, unmarshallFunc func(in []byte, out interface{}) (err error)) (string, error) {
	configClient, err := f.referenceFactory.GetConfigClient(nil)
	if err != nil {
		return "", fmt.Errorf("unable to get config client: %w", err)
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return "", fmt.Errorf("unable to get client: %w", configErr)
	}
	credentialsPath := config.UtilCredentialHelperConfiguration.FilePath
	if credentialsPath == "" {
		// Get the home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to get home directory: %w", err)
		}
		credentialsPath = path.Join(homeDir, ".bbctl", "credentials.yaml")
	}

	// Read the credentials file
	credentialsYaml, err := os.ReadFile(credentialsPath)
	if err != nil {
		return "", fmt.Errorf("unable to read credentials file %v: %w", credentialsPath, err)
	}

	// Unmarshal the credentials file
	var credentialsFile CredentialsFile
	err = unmarshallFunc(credentialsYaml, &credentialsFile)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal credentials file %v: %w", credentialsPath, err)
	}

	// Find the credentials for the uri
	credentials := Credentials{}
	for _, c := range credentialsFile.Credentials {
		if c.URI == uri {
			credentials = c
			break
		}
	}
	// If the credentials URI is empty, return an error
	if credentials.URI == "" {
		return "", fmt.Errorf("no credentials found for %v in %v", uri, credentialsPath)
	}

	// Return the requested component
	switch component {
	case "username":
		return credentials.Username, nil
	case "password":
		return credentials.Password, nil
	default:
		return "", fmt.Errorf("invalid component %v", component)
	}
}

// CredentialHelper is a function type that can be used to fetch credential values
// The function takes 2 parameters:
//
// * component (string) - The Credentials struct field name, either `username` or `password`
//
// * uri (string) - The Credentials struct URI value which uniquely identifies the requested component
//
// These parameters are passed into custom credential helpers as CLI arguments in the same order.
type CredentialHelper func(string, string) (string, error)

// GetCredentialHelper returns a function reference to the configured credential helper function that can
// be called to fetch credential values. A custom credential helper function is any CLI executable
// script which can be passed into this function via the bbctl config settings as a file path.
//
// Errors when no credential helper is defined, there is an issue reading credentials from a file,
// there is an issue running a custom credential helper script, and when an empty value is returned
// for a requested credential component
func (f *UtilityFactory) GetCredentialHelper() (CredentialHelper, error) {
	credentialHelper := func(component string, uri string) (string, error) {
		configClient, err := f.referenceFactory.GetConfigClient(nil)
		if err != nil {
			return "", fmt.Errorf("unable to get config client: %w", err)
		}
		config, configErr := configClient.GetConfig()
		if configErr != nil {
			return "", fmt.Errorf("unable to get client: %w", configErr)
		}
		helper := config.UtilCredentialHelperConfiguration.CredentialHelper
		if helper == "" {
			return "", fmt.Errorf("no credential helper defined (\"big-bang-credential-helper\")")
		}
		output := ""
		if helper == "credentials-file" {
			output, err = f.ReadCredentialsFile(component, uri)
			if err != nil {
				return "", fmt.Errorf("unable to read credentials file: %w", err)
			}
		} else {
			cmd := exec.Command(helper, component, uri)
			rawOutput, err := cmd.Output()
			if err != nil {
				return "", fmt.Errorf("unable to get %v from %v using %v: %w", component, uri, helper, err)
			}
			output = string(rawOutput[:])
		}
		if output == "" {
			return "", fmt.Errorf("no %v found for %v in %v", component, uri, helper)
		}
		return output, nil
	}
	return credentialHelper, nil
}

// GetAWSClient initializes and returns a new AWS API client
func (f *UtilityFactory) GetAWSClient() (bbAws.Client, error) {
	clientGetter := bbAws.ClientGetter{}
	client := clientGetter.GetClient()
	return client, nil
}

// GetGitLabClient initializes and returns a new GitLab API client
func (f *UtilityFactory) GetGitLabClient() (bbGitLab.Client, error) {
	return f.getGitLabClient()
}

// GetGitLabClient initializes and returns a new GitLab API client
func (f *UtilityFactory) getGitLabClient(options ...gitlab.ClientOptionFunc) (bbGitLab.Client, error) {
	configClient, err := f.referenceFactory.GetConfigClient(nil)
	if err != nil {
		return nil, err
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return nil, fmt.Errorf("unable to get client: %w", configErr)
	}
	clientGetter := bbGitLab.ClientGetter{}
	client, err := clientGetter.GetClient(
		config.GitLabConfiguration.BaseURL,
		config.GitLabConfiguration.Token,
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get GitLab client: %w", err)
	}
	return client, nil
}

// GetHelmClient initializes and returns a new Helm client that can perform operations in the given namespace
//
// # Returns a nil client and an error if there are any issues with the intialization
//
// Errors when there are issues with the bbctl configurations
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
// Errors when there are issues with the bbctl configurations
func (f *UtilityFactory) GetK8sClientset(cmd *cobra.Command) (kubernetes.Interface, error) {
	config, err := f.referenceFactory.GetRestConfig(cmd)
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
// Errors when there are issues with the bbctl configurations
func (f *UtilityFactory) GetK8sDynamicClient(cmd *cobra.Command) (dynamic.Interface, error) {
	configClient, err := f.referenceFactory.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return nil, fmt.Errorf("unable to get client: %w", configErr)
	}
	return bbK8sUtil.BuildDynamicClient(config)
}

// GetOutputClient initializes and returns an output client based on the specified format flag and I/O streams.
// It retrieves the "format" flag from the provided command, initializes an output client getter,
// and obtains the appropriate output client using the specified format and streams. Default format is "text"
//
// Errors when issues occur using the clients Output method.
func (f *UtilityFactory) GetOutputClient(cmd *cobra.Command) (bbOutput.Client, error) {
	streams, err := f.referenceFactory.GetIOStream()
	if err != nil {
		// NOTE: This branch is impossible to test because the GetIOStream method is hardcoded to return nil
		return nil, err
	}
	configClient, err := f.referenceFactory.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	config, err := configClient.GetConfig()
	if err != nil {
		return nil, err
	}
	outputCLientGetter := bbOutput.ClientGetter{}
	outputClient := outputCLientGetter.GetClient(config.OutputConfiguration.Format, *streams)

	return outputClient, nil
}

// GetLoggingClient initializes and returns a new logging client using the default slog logger implementation
//
// Errors when there are issues initializing the logger
func (f *UtilityFactory) GetLoggingClient() (bbLog.Client, error) {
	return f.referenceFactory.GetLoggingClientWithLogger(nil)
}

// GetLoggingClientWithLogger initializes and returns a new logging client using the given logger implementation
//
// Errors when there are issues initializing the logger
func (f *UtilityFactory) GetLoggingClientWithLogger(logger *slog.Logger) (bbLog.Client, error) {
	clientGetter := bbLog.ClientGetter{}
	return clientGetter.GetClient(logger), nil
}

// GetRuntimeClient initializes and returns a new k8s runtime client by calling the client.New() function
//
// # Returns a nil client and an error if there are any issues with the intialization
//
// Errors when there are issues creating the k8s REST config
func (f *UtilityFactory) GetRuntimeClient(scheme *runtime.Scheme) (runtimeClient.Client, error) {
	// init runtime controller client
	runtimeClient, err := runtimeClient.New(
		ctrl.GetConfigOrDie(),
		runtimeClient.Options{Scheme: scheme},
	)
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
	configClient, err := f.referenceFactory.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return nil, fmt.Errorf("unable to get client: %w", configErr)
	}
	return bbK8sUtil.BuildKubeConfig(config)
}

// GetCommandExecutor initializes and returns a new SPDY executor that can run the given command in a Pod in the k8s cluster
//
// # Returns a nil executor and an error if there are any issues with the intialization
func (f *UtilityFactory) GetCommandExecutor(
	cmd *cobra.Command,
	pod *coreV1.Pod,
	container string,
	command []string,
	stdout io.Writer,
	stderr io.Writer,
) (remoteCommand.Executor, error) {
	client, err := f.referenceFactory.GetK8sClientset(cmd)
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

	// REST config is already validated in the f.referenceFactory.GetK8sClientset(cmd) call above
	config, _ := f.referenceFactory.GetRestConfig(cmd)

	return remoteCommand.NewSPDYExecutor(config, "POST", req.URL())
}

// Internal helper function to create configs for GetHelmClient
//
// Errors if Helm action.Configuration.Init fails
func (f *UtilityFactory) getHelmConfig(
	cmd *cobra.Command,
	namespace string,
) (*action.Configuration, error) {
	configClient, err := f.referenceFactory.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	bbctlConfig, configErr := configClient.GetConfig()
	if configErr != nil {
		return nil, fmt.Errorf("unable to get client: %w", configErr)
	}

	loggingClient, err := f.referenceFactory.GetLoggingClient()
	if err != nil {
		// NOTE: this branch is impossible to test because a failure to get logger would have already errored at getconfigclient
		return nil, err
	}
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
	err = actionConfig.Init(
		clientGetter,
		namespace,
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)

	return actionConfig, err
}

// GetCommandWrapper initializes and returns a new Command instance which encapsulates the functionality needed to run a CLI command
// `name` is the command to execute i.e. kubectl
// `args` string values are all passed to the command as CLI arguments
func (f *UtilityFactory) GetCommandWrapper(
	name string,
	args ...string,
) (*bbUtilApiWrappers.Command, error) {
	return bbUtilApiWrappers.NewExecRunner(name, args...), nil
}

// GetIstioClientSet initializes and returns a new istio client set by calling versioned.NewForConfig() with the provided REST config settings
//
// # Returns a nil client and an error if there are any issues with the intialization
func (f *UtilityFactory) GetIstioClientSet(
	cfg *rest.Config,
) (bbUtilApiWrappers.IstioClientset, error) {
	clientSet, err := versioned.NewForConfig(cfg)
	return clientSet, err
}

// GetConfigClient initializes and returns a new bbctl config client
//
// # Returns a nil client and an error if there are any issues with the intialization
func (f *UtilityFactory) GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error) {
	loggingClient, err := f.referenceFactory.GetLoggingClient()
	if err != nil {
		// NOTE: This branch is impossible to test because the GetLoggingClient method is hardcoded to return nil
		return nil, err
	}
	v, err := f.referenceFactory.GetViper()
	if err != nil {
		// NOTE: This branch is impossible to test because the GetViper method is hardcoded to return nil
		return nil, err
	}
	clientGetter := bbConfig.ClientGetter{}
	client, err := clientGetter.GetClient(command, &loggingClient, v)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// GetViper returns the viper instance
func (f *UtilityFactory) GetViper() (*viper.Viper, error) {
	return f.getViperFunction()
}

// getViper initializes and returns a new viper instance
func getViper() (*viper.Viper, error) {
	return viper.New(), nil
}

// GetIOStream initializes and returns a new IOStreams object used to interact with console input, output, and error output
func (f *UtilityFactory) GetIOStream() (*genericIOOptions.IOStreams, error) {
	return &genericIOOptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}, nil
}

// GetPipe returns the currently set pipe reader and writer
func (f *UtilityFactory) GetPipe() (commonInterfaces.FileLike, commonInterfaces.FileLike, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create pipe: %w", err)
	}
	return r, w, nil
}
