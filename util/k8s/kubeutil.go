package k8s

import (
	"path/filepath"

	"k8s.io/client-go/dynamic"
	restClient "k8s.io/client-go/rest"
	clientCmd "k8s.io/client-go/tools/clientcmd"
	homeDir "k8s.io/client-go/util/homedir"

	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
)

// BuildKubeConfigFromFlags is a helper function that builds a config object used to
// interact with the k8s cluster. The configuration is sourced in the following order:
//
// Read config from file specified using --kubeconfig flag
// Read config from file(s) specified using KUBECONFIG env variable
// Read config from default location at $HOME/.kube/config
//
// If all these steps fail, fallback to default kubernetes config mechanism.
func BuildKubeConfig(bbConfig *bbConfig.GlobalConfiguration) (*restClient.Config, error) {
	kubeConfig := bbConfig.UtilK8sConfiguration.Kubeconfig
	if kubeConfig != "" {
		return GetKubeConfigFromPathList(kubeConfig)
	}

	if kubeConfig == "" {
		if home := homeDir.HomeDir(); home != "" {
			kubeConfig = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientCmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// BuildDynamicClientFromFlags is a helper function that builds a dynamic client
// used to interact with the k8s cluster.
func BuildDynamicClient(bbConfig *bbConfig.GlobalConfiguration) (dynamic.Interface, error) {
	restConfig, err := BuildKubeConfig(bbConfig)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(restConfig)
}

// GetKubeConfigFromPathList is a helper function that builds config object used to
// interact with the k8s cluster using a list of kubeconfig file(s)
func GetKubeConfigFromPathList(configPaths string) (*restClient.Config, error) {
	configPathList := filepath.SplitList(configPaths)
	configLoadingRules := &clientCmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = configPaths
	} else {
		configLoadingRules.Precedence = configPathList
	}
	clientConfig := clientCmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, nil)
	return clientConfig.ClientConfig()
}
