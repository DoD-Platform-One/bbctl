package util

import (
	"github.com/spf13/pflag"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/helm"
	bbk8sutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/k8s"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Factory interface
type Factory interface {
	GetHelmClient(namespace string) (helm.Client, error)
	GetK8sClientset() (kubernetes.Interface, error)
	GetRuntimeClient(*runtime.Scheme) (runtimeclient.Client, error)
	GetK8sDynamicClient() (dynamic.Interface, error)
}

// NewFactory - new factory method
func NewFactory(flags *pflag.FlagSet) *UtilityFactory {
	return &UtilityFactory{flags: flags}
}

// UtilityFactory - util factory
type UtilityFactory struct {
	flags *pflag.FlagSet
}

// GetHelmClient - get helm client
func (f *UtilityFactory) GetHelmClient(namespace string) (helm.Client, error) {
	config, err := bbk8sutil.BuildKubeConfigFromFlags(f.flags)
	if err != nil {
		return nil, err
	}

	opt := &helm.Options{
		Namespace:        namespace,
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
		Linting:          true,
		RestConfig:       config,
	}

	return helm.New(opt)
}

// GetK8sClientset - get k8s clientset
func (f *UtilityFactory) GetK8sClientset() (kubernetes.Interface, error) {

	config, err := bbk8sutil.BuildKubeConfigFromFlags(f.flags)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// GetK8sDynamicClient - get k8s dynamic client
func (f *UtilityFactory) GetK8sDynamicClient() (dynamic.Interface, error) {
	return bbk8sutil.BuildDynamicClientFromFlags(f.flags)
}

// GetRuntimeClient - get runtime client
func (f *UtilityFactory) GetRuntimeClient(scheme *runtime.Scheme) (runtimeclient.Client, error) {

	// init runtime cotroller client
	runtimeClient, err := runtimeclient.New(ctrl.GetConfigOrDie(), runtimeclient.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return runtimeClient, err
}
