package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	pFlag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/dynamic"
	restClient "k8s.io/client-go/rest"
	clientCmd "k8s.io/client-go/tools/clientcmd"
	homeDir "k8s.io/client-go/util/homedir"
)

// BuildKubeConfigFromFlags - is a helper function that builds config object used to
// interact with the k8s cluster. The configuration is sourced in the following order:
//
// Read config from file specified using --kubeconfig flag
// Read config from file(s) specified using KUBECONFIG env variable
// Read config from default location at $HOME/.kube/config
//
// If all these steps fail, fallback to default kubernetes config mechanism.
func BuildKubeConfigFromFlags(flags *pFlag.FlagSet) (*restClient.Config, error) {
	kubeConfig, _ := flags.GetString("kubeconfig")

	if kubeConfig != "" {
		_, err := os.Stat(kubeConfig)
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
	}

	if kubeConfig == "" {
		kubeConfig = viper.GetString("kubeconfig")
		if kubeConfig != "" {
			return GetKubeConfigFromPathList(kubeConfig)
		}
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
func BuildDynamicClientFromFlags(flags *pFlag.FlagSet) (dynamic.Interface, error) {
	restConfig, err := BuildKubeConfigFromFlags(flags)
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
