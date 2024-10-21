package schemas

import "github.com/spf13/viper"

type VersionConfiguration struct {
	// Gatekeeper: if true, check gatekeeper
	Client bool `mapstructure:"client" yaml:"client"`

	// AllCharts enables fetching information on Big Bang and it's subcharts
	AllCharts bool `mapstructure:"all-charts" yaml:"all-charts"`

	// CheckForUpdates configures bbctl to check for updates for a given chart
	CheckForUpdates bool `mapstructure:"check-for-updates" yaml:"check-for-updates"`

	// NoSHAs disables checking the deployed image SHAs against the upstream repo
	NoSHAs bool `mapstructure:"no-shas" yaml:"no-shas"`
}

// ReconcileConfiguration reconciles the configuration.
func (v *VersionConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	v.Client = instance.GetBool("client")
	v.AllCharts = instance.GetBool("all-charts")
	v.CheckForUpdates = instance.GetBool("check-for-updates")
	v.NoSHAs = instance.GetBool("no-shas")
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (v *VersionConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
