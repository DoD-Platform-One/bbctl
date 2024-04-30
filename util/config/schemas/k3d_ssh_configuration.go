package schemas

import "github.com/spf13/viper"

type K3dSshConfiguration struct {
	// ssh user: valid ssh username
	User string `mapstructure:"ssh-username" yaml:"ssh-username"`
}

// ReconcileConfiguration reconciles the configuration.
func (k *K3dSshConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("ssh-username") {
		k.User = instance.GetString("ssh-username")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (k *K3dSshConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
