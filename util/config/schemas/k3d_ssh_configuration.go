package schemas

import "github.com/spf13/viper"

type K3dSshConfiguration struct {
	// ssh user: valid ssh username
	User string `mapstructure:"ssh-username" yaml:"ssh-username"`
	// private ip: if true, use private ip
	PrivateIp bool `mapstructure:"private-ip" yaml:"private-ip"`
}

// ReconcileConfiguration reconciles the configuration.
func (k *K3dSshConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	k.PrivateIp = instance.GetBool("private-ip")
	if instance.IsSet("ssh-username") {
		k.User = instance.GetString("ssh-username")
	} else {
		if k.User == "" {
			k.User = "ubuntu"
		}
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (k *K3dSshConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
