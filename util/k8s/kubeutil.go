package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// BuildKubeConfigFromFlags - is a helper function that builds config object used to
// interact with the k8s cluster. The configuration is sourced in the following order:
//
// Read config from file specified using --kubeconfig flag
// Read config from file(s) specified using KUBECONFIG env variable
// Read config from default location at $HOME/.kube/config
//
// If all these steps fail, fallback to default kubernetes config mechanism.
func BuildKubeConfigFromFlags(flags *pflag.FlagSet) (*restclient.Config, error) {
	kubeconfig, _ := flags.GetString("kubeconfig")

	if kubeconfig != "" {
		_, err := os.Stat(kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
	}

	if kubeconfig == "" {
		kubeconfig = viper.GetString("kubeconfig")
		if kubeconfig != "" {
			return GetKubeConfigFromPathList(kubeconfig)
		}
	}

	if kubeconfig == "" {

		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// BuildDynamicClientFromFlags is a helper function that builds a dynamic client
// used to interact with the k8s cluster.
func BuildDynamicClientFromFlags(flags *pflag.FlagSet) (dynamic.Interface, error) {
	restConfig, err := BuildKubeConfigFromFlags(flags)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(restConfig)
}

// GetKubeConfigFromPathList is a helper function that builds config object used to
// interact with the k8s cluster using a list of kubeconfig file(s)
func GetKubeConfigFromPathList(configPaths string) (*restclient.Config, error) {
	configPathList := filepath.SplitList(configPaths)
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{}
	if len(configPathList) <= 1 {
		configLoadingRules.ExplicitPath = configPaths
	} else {
		configLoadingRules.Precedence = configPathList
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, nil)
	return clientConfig.ClientConfig()
}
