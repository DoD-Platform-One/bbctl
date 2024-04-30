package schemas

import "github.com/spf13/viper"

type PreflightCheckConfiguration struct {
	// registry server: valid registry server
	RegistryServer string `mapstructure:"registryserver" yaml:"registryserver"`
	// registry username: valid registry username
	RegistryUsername string `mapstructure:"registryusername" yaml:"registryusername"`
	// registry password: valid registry password
	RegistryPassword string `mapstructure:"registrypassword" yaml:"registrypassword"`
}

// ReconcileConfiguration reconciles the configuration.
func (p *PreflightCheckConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("registryserver") {
		p.RegistryServer = instance.GetString("registryserver")
	}
	if instance.IsSet("registryusername") {
		p.RegistryUsername = instance.GetString("registryusername")
	}
	if instance.IsSet("registrypassword") {
		p.RegistryPassword = instance.GetString("registrypassword")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (p *PreflightCheckConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
