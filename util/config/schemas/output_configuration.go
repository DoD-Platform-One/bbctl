package schemas

import (
	"github.com/spf13/viper"
	bbOutput "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

type OutputConfiguration struct {
	// Format for output: json, text, yaml
	Format bbOutput.Format `mapstructure:"format" yaml:"format"`
}

func (o *OutputConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("format") {
		o.Format = bbOutput.Format(instance.GetString("format"))
	} else {
		if o.Format == "" {
			o.Format = bbOutput.TEXT
		}
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (o *OutputConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
