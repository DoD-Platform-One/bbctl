package config

import (
	"errors"
	"fmt"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
)

// SetAndBindFlag sets and binds a flag to a command and viper.
//
// Note these are some unsupported types: map[string]interface{}, map[string][]string, time.Time
func SetAndBindFlag(client *ConfigClient, name string, shorthand string, value any, description string) error {
	command := client.command
	if command == nil {
		return errors.New("command is required to set and bind flag")
	}
	switch typedValue := value.(type) {
	case bool:
		command.PersistentFlags().BoolP(name, shorthand, typedValue, description)
	case time.Duration:
		command.PersistentFlags().DurationP(name, shorthand, typedValue, description)
	case float64:
		command.PersistentFlags().Float64P(name, shorthand, typedValue, description)
	case int:
		command.PersistentFlags().IntP(name, shorthand, typedValue, description)
	case int32:
		command.PersistentFlags().Int32P(name, shorthand, typedValue, description)
	case int64:
		command.PersistentFlags().Int64P(name, shorthand, typedValue, description)
	case []int:
		command.PersistentFlags().IntSliceP(name, shorthand, typedValue, description)
	case string:
		command.PersistentFlags().StringP(name, shorthand, typedValue, description)
	// this is not supported by pFlag without a custom implementation
	// case map[string]interface{}:
	case map[string]string:
		command.PersistentFlags().StringToStringP(name, shorthand, typedValue, description)
	// this is not supported by pFlag without a custom implementation
	// case map[string][]string:
	case []string:
		command.PersistentFlags().StringSliceP(name, shorthand, typedValue, description)
	// this is not supported by pFlag without a custom implementation
	// case time.Time:
	case uint:
		command.PersistentFlags().UintP(name, shorthand, typedValue, description)
	case uint32:
		command.PersistentFlags().Uint32P(name, shorthand, typedValue, description)
	case uint64:
		command.PersistentFlags().Uint64P(name, shorthand, typedValue, description)
	default:
		return errors.New("unsupported type")
	}

	return client.viperInstance.BindPFlag(name, command.PersistentFlags().Lookup(name))
}

// getConfig returns the global configuration.
func getConfig(client *ConfigClient) (*schemas.GlobalConfiguration, error) {
	return getConfigWithFunc(client, client.viperInstance.Unmarshal, client.viperInstance.BindPFlags)
}

// see getConfig
func getConfigWithFunc(client *ConfigClient, u func(rawVal any, opts ...viper.DecoderConfigOption) error, b func(flags *pflag.FlagSet) error) (*schemas.GlobalConfiguration, error) {
	var config schemas.GlobalConfiguration
	unmarshalError := u(&config)
	if unmarshalError != nil {
		return nil, fmt.Errorf("error unmarshalling configuration: %w", unmarshalError)
	}

	if client.command != nil {
		bindingError := b(client.command.PersistentFlags())
		if bindingError != nil {
			return nil, fmt.Errorf("error binding flags: %w", bindingError)
		}
	}

	reconcileError := config.ReconcileConfiguration(client.viperInstance)
	if reconcileError != nil {
		return nil, fmt.Errorf("error reconciling configuration: %w", reconcileError)
	}
	validator := validator.New()
	validatorError := validator.Struct(config)
	if validatorError != nil {
		return nil, fmt.Errorf("error during validation for configuration: %w", validatorError)
	}
	return &config, nil
}
