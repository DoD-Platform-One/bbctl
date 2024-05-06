package schemas

import "github.com/spf13/viper"

type VersionConfiguration struct {
	// Gatekeeper: if true, check gatekeeper
	Client bool `mapstructure:"client" yaml:"client"`
}

// ReconcileConfiguration reconciles the configuration.
func (v *VersionConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	v.Client = instance.GetBool("client")
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (p *VersionConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
