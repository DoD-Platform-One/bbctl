package schemas

import "github.com/spf13/viper"

type DeployBigBangConfiguration struct {
	// Include some boilerplate suitable for deploying into k3d: true, false
	K3d bool `mapstructure:"k3d" yaml:"k3d"`
	// Enable this bigbang addon in the deployment: any key in the addons map in the bb values.yaml
	Addon []string `mapstructure:"addon" yaml:"addon"`
}

// ReconcileConfiguration reconciles the configuration.
func (d *DeployBigBangConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("k3d") {
		d.K3d = instance.GetBool("k3d")
	}
	if instance.IsSet("addon") {
		d.Addon = instance.GetStringSlice("addon")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (d *DeployBigBangConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
