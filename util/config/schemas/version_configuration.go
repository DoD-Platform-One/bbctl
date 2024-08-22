package schemas

import "github.com/spf13/viper"

type VersionConfiguration struct {
	// Gatekeeper: if true, check gatekeeper
	Client bool `mapstructure:"client" yaml:"client"`

	// AllCharts tells the version command to print the currently deployed release version for Big Bang and all it's components
	AllCharts bool `mapstructure:"all-charts" yaml:"all-charts"`
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
