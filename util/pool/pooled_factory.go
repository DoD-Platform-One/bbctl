package pool

import (
	"io"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	remoteCommand "k8s.io/client-go/tools/remotecommand"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbUtilApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/apiwrappers"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	bbOutput "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ErrFactoryNotInitialized is returned when the factory is not initialized
type ErrFactoryNotInitialized struct{}

// Error returns the error message
func (e ErrFactoryNotInitialized) Error() string {
	return "factory not initialized"
}

// Factory is an interface for creating clients
type PooledFactory struct {
	underlyingFactory bbUtil.Factory
	awsClient         bbAws.Client
	gitLabClient      bbGitLab.Client
	helmClients       helmClientPool
	k8sClientset      kubernetes.Interface
	loggerClients     loggerClientPool
	runtimeClients    runtimeClientPool
	k8sDynamicClient  dynamic.Interface
	outputClients     outputClientPool
	restConfig        *rest.Config
	credentialHelper  bbUtil.CredentialHelper
	istioClientSets   istioClientsetPool
	configClient      *bbConfig.ConfigClient
	viper             *viper.Viper
	ioStream          *genericIOOptions.IOStreams
}

// NewPooledFactory returns a new pooled factory
func NewPooledFactory() *PooledFactory {
	return &PooledFactory{
		helmClients:     make([]*helmClientInstance, 0),
		loggerClients:   make([]*loggerClientInstance, 0),
		runtimeClients:  make([]*runtimeClientInstance, 0),
		outputClients:   make([]*outputClientInstance, 0),
		istioClientSets: make([]*istioClientsetInstance, 0),
	}
}

// SetUnderlyingFactory sets the underlying factory
func (pf *PooledFactory) SetUnderlyingFactory(factory bbUtil.Factory) {
	pf.underlyingFactory = factory
}

// GetAWSClient returns the AWS client
//
// Pooled by client (singleton)
func (pf *PooledFactory) GetAWSClient() (bbAws.Client, error) {
	if pf.awsClient != nil {
		return pf.awsClient, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetAWSClient()
	if err != nil {
		return client, err
	}
	pf.awsClient = client
	return client, nil
}

// GetGitLabClient returns the GitLab client
//
// Pooled by client (singleton)
func (pf *PooledFactory) GetGitLabClient() (bbGitLab.Client, error) {
	if pf.gitLabClient != nil {
		return pf.gitLabClient, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetGitLabClient()
	if err != nil {
		return client, err
	}
	pf.gitLabClient = client
	return client, nil
}

// GetHelmClient returns the Helm client
//
// Pooled by namespace, we assume cmd never changes
func (pf *PooledFactory) GetHelmClient(cmd *cobra.Command, namespace string) (helm.Client, error) {
	if contains, client := pf.helmClients.contains(namespace); contains {
		return client, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetHelmClient(cmd, namespace)
	if err != nil {
		return client, err
	}
	pf.helmClients.add(client, namespace)
	return client, nil
}

// GetK8sClientset returns the Kubernetes clientset
//
// Pooled by client (singleton), we assume cmd never changes
func (pf *PooledFactory) GetK8sClientset(cmd *cobra.Command) (kubernetes.Interface, error) {
	if pf.k8sClientset != nil {
		return pf.k8sClientset, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetK8sClientset(cmd)
	if err != nil {
		return client, err
	}
	pf.k8sClientset = client
	return client, nil
}

// GetLoggingClient returns the logging client
//
// Pooled by logger (see GetLoggingClientWithLogger)
func (pf *PooledFactory) GetLoggingClient() (bbLog.Client, error) {
	return pf.GetLoggingClientWithLogger(nil)
}

// GetLoggingClientWithLogger returns the logging client with a logger
//
// Pooled by logger
func (pf *PooledFactory) GetLoggingClientWithLogger(logger *slog.Logger) (bbLog.Client, error) {
	if contains, client := pf.loggerClients.contains(logger); contains {
		return client, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetLoggingClientWithLogger(logger)
	if err == nil {
		pf.loggerClients.add(client, logger)
	}
	return client, err
}

// GetRuntimeClient returns the runtime client
//
// Pooled by scheme
func (pf *PooledFactory) GetRuntimeClient(scheme *runtime.Scheme) (runtimeClient.Client, error) {
	if contains, client := pf.runtimeClients.contains(scheme); contains {
		return client, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetRuntimeClient(scheme)
	if err != nil {
		return client, err
	}
	pf.runtimeClients.add(client, scheme)
	return client, nil
}

// GetK8sDynamicClient returns the Kubernetes dynamic client
//
// Pooled by client (singleton), we assume cmd never changes
func (pf *PooledFactory) GetK8sDynamicClient(cmd *cobra.Command) (dynamic.Interface, error) {
	if pf.k8sDynamicClient != nil {
		return pf.k8sDynamicClient, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetK8sDynamicClient(cmd)
	if err != nil {
		return client, err
	}
	pf.k8sDynamicClient = client
	return client, nil
}

// GetOutputClient returns the output client
//
// Pooled by streams, we assume cmd never changes
func (pf *PooledFactory) GetOutputClient(cmd *cobra.Command, streams genericIOOptions.IOStreams) (bbOutput.Client, error) {
	if contains, client := pf.outputClients.contains(streams); contains {
		return client, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetOutputClient(cmd, streams)
	if err == nil {
		pf.outputClients.add(client, streams)
	}
	return client, err
}

// GetRestConfig returns the REST config
//
// Pooled by config (singleton), we assume cmd never changes
func (pf *PooledFactory) GetRestConfig(cmd *cobra.Command) (*rest.Config, error) {
	if pf.restConfig != nil {
		return pf.restConfig, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetRestConfig(cmd)
	if err != nil {
		return client, err
	}
	pf.restConfig = client
	return client, nil
}

// GetCommandExecutor returns the command executor
//
// Not pooled (pass-through)
func (pf *PooledFactory) GetCommandExecutor(cmd *cobra.Command, pod *coreV1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remoteCommand.Executor, error) {
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	return pf.underlyingFactory.GetCommandExecutor(cmd, pod, container, command, stdout, stderr)
}

// GetCredentialHelper returns the credential helper
//
// Pooled by helper (singleton)
func (pf *PooledFactory) GetCredentialHelper() (bbUtil.CredentialHelper, error) {
	if pf.credentialHelper != nil {
		return pf.credentialHelper, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetCredentialHelper()
	if err == nil {
		pf.credentialHelper = client
	}
	return client, err
}

// GetCommandWrapper returns the command wrapper
//
// Not pooled (pass-through)
func (pf *PooledFactory) GetCommandWrapper(name string, args ...string) (*bbUtilApiWrappers.Command, error) {
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	return pf.underlyingFactory.GetCommandWrapper(name, args...)
}

// GetIstioClientSet returns the Istio clientset
//
// Pooled by cfg
func (pf *PooledFactory) GetIstioClientSet(cfg *rest.Config) (bbUtilApiWrappers.IstioClientset, error) {
	if contains, client := pf.istioClientSets.contains(cfg); contains {
		return client, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetIstioClientSet(cfg)
	if err != nil {
		return client, err
	}
	pf.istioClientSets.add(client, cfg)
	return client, nil
}

// GetConfigClient returns the config client
//
// Pooled by client (singleton), we assume cmd never changes
func (pf *PooledFactory) GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error) {
	if pf.configClient != nil {
		return pf.configClient, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetConfigClient(command)
	if err != nil {
		return client, err
	}
	pf.configClient = client
	return client, nil
}

// GetViper returns the Viper
//
// Pooled by client (singleton)
func (pf *PooledFactory) GetViper() (*viper.Viper, error) {
	if pf.viper != nil {
		return pf.viper, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetViper()
	if err == nil {
		pf.viper = client
	}
	return client, err
}

// SetViper sets the Viper
func (pf *PooledFactory) SetViper(v *viper.Viper) error {
	if v == pf.viper {
		return nil
	}
	if pf.underlyingFactory == nil {
		return &ErrFactoryNotInitialized{}
	}
	pf.viper = v
	return pf.underlyingFactory.SetViper(v)
}

// GetIOStream returns the IO stream
//
// Pooled by instance (singleton)
func (pf *PooledFactory) GetIOStream() (*genericIOOptions.IOStreams, error) {
	if pf.ioStream != nil {
		return pf.ioStream, nil
	}
	if pf.underlyingFactory == nil {
		return nil, &ErrFactoryNotInitialized{}
	}
	client, err := pf.underlyingFactory.GetIOStream()
	if err == nil {
		pf.ioStream = client
	}
	return client, err
}
