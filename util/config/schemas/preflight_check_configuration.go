package schemas

import "github.com/spf13/viper"

type PreflightCheckConfiguration struct {
	// registry server: valid registry server
	RegistryServer string `mapstructure:"registryserver" yaml:"registryserver"`
	// registry username: valid registry username
	RegistryUsername string `mapstructure:"registryusername" yaml:"registryusername"`
	// registry password: valid registry password
	RegistryPassword string `mapstructure:"registrypassword" yaml:"registrypassword"`
	// retry count: number of retries
	RetryCount int `mapstructure:"retrycount" yaml:"retrycount"`
	// retry delay: delay between retries
	RetryDelay int `mapstructure:"retrydelay" yaml:"retrydelay"`
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
	if instance.IsSet("retrycount") {
		p.RetryCount = instance.GetInt("retrycount")
	} else {
		if p.RetryCount == 0 {
			p.RetryCount = 5
		}
	}
	if instance.IsSet("retrydelay") {
		p.RetryDelay = instance.GetInt("retrydelay")
	} else {
		if p.RetryDelay == 0 {
			p.RetryDelay = 5
		}
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (p *PreflightCheckConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
