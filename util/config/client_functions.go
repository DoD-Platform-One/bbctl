package config

import (
	"errors"
	"fmt"
	"time"

	validator "github.com/go-playground/validator/v10"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
)

// SetAndBindFlag sets and binds a flag to a command and viper.
//
// Note these are some unsupported types: map[string]interface{}, map[string][]string, time.Time
func SetAndBindFlag(client *ConfigClient, name string, value interface{}, description string) error {
	command := client.command
	if command == nil {
		return errors.New("command is required to set and bind flag")
	}
	switch typedValue := value.(type) {
	case bool:
		command.PersistentFlags().Bool(name, typedValue, description)
	case time.Duration:
		command.PersistentFlags().Duration(name, typedValue, description)
	case float64:
		command.PersistentFlags().Float64(name, typedValue, description)
	case int:
		command.PersistentFlags().Int(name, typedValue, description)
	case int32:
		command.PersistentFlags().Int32(name, typedValue, description)
	case int64:
		command.PersistentFlags().Int64(name, typedValue, description)
	case []int:
		command.PersistentFlags().IntSlice(name, typedValue, description)
	case string:
		command.PersistentFlags().String(name, typedValue, description)
	// this is not supported by pFlag without a custom implementation
	// case map[string]interface{}:
	case map[string]string:
		command.PersistentFlags().StringToString(name, typedValue, description)
	// this is not supported by pFlag without a custom implementation
	// case map[string][]string:
	case []string:
		command.PersistentFlags().StringSlice(name, typedValue, description)
	// this is not supported by pFlag without a custom implementation
	// case time.Time:
	case uint:
		command.PersistentFlags().Uint(name, typedValue, description)
	case uint32:
		command.PersistentFlags().Uint32(name, typedValue, description)
	case uint64:
		command.PersistentFlags().Uint64(name, typedValue, description)
	default:
		return errors.New("unsupported type")
	}

	return client.viperInstance.BindPFlag(name, command.PersistentFlags().Lookup(name))
}

// getConfig returns the global configuration.
func getConfig(client *ConfigClient) (*schemas.GlobalConfiguration, error) {
	var config schemas.GlobalConfiguration
	unmarshalError := client.viperInstance.Unmarshal(&config)
	if unmarshalError != nil {
		return nil, fmt.Errorf("Error unmarshalling configuration: %w", unmarshalError)
	}

	if client.command != nil {
		bindingError := client.viperInstance.BindPFlags(client.command.PersistentFlags())
		if bindingError != nil {
			return nil, fmt.Errorf("Error binding flags: %w", bindingError)
		}
	}

	reconcileError := config.ReconcileConfiguration(client.viperInstance)
	if reconcileError != nil {
		return nil, fmt.Errorf("Error reconciling configuration: %w", reconcileError)
	}
	validator := validator.New()
	validatorError := validator.Struct(config)
	if validatorError != nil {
		return nil, fmt.Errorf("Error during validation for configuration: %w", validatorError)
	}
	return &config, nil
}
