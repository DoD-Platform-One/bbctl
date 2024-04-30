package schemas

import "github.com/spf13/viper"

type UtilK8sConfiguration struct {
	// kubeconfig file path: file path
	Kubeconfig string `mapstructure:"kubeconfig" yaml:"kubeconfig"`
}

// ReconcileConfiguration reconciles the configuration.
func (u *UtilK8sConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("kubeconfig") {
		u.Kubeconfig = instance.GetString("kubeconfig")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (u *UtilK8sConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
