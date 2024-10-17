package schemas

import (
	"github.com/spf13/viper"
)

type BaseConfiguration interface {
	// reconcileConfiguration reconciles the configuration.
	ReconcileConfiguration(instance *viper.Viper) error
	getSubConfigurations() []BaseConfiguration
}
