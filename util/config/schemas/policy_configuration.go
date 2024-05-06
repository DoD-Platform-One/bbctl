package schemas

import "github.com/spf13/viper"

type PolicyConfiguration struct {
	// Gatekeeper: if true, check gatekeeper
	Gatekeeper bool `mapstructure:"gatekeeper" yaml:"gatekeeper"`
	// Kyverno: if true, check kyverno
	Kyverno bool `mapstructure:"kyverno" yaml:"kyverno"`
}

// ReconcileConfiguration reconciles the configuration.
func (p *PolicyConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("gatekeeper") {
		p.Gatekeeper = instance.GetBool("gatekeeper")
	}
	if instance.IsSet("kyverno") {
		p.Kyverno = instance.GetBool("kyverno")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (p *PolicyConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
