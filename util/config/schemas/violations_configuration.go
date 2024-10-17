package schemas

import "github.com/spf13/viper"

type ViolationsConfiguration struct {
	// Audit: if true, list violations in audit mode
	Audit bool `mapstructure:"audit" yaml:"audit"`
}

// ReconcileConfiguration reconciles the configuration.
func (v *ViolationsConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("audit") {
		v.Audit = instance.GetBool("audit")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (v *ViolationsConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
