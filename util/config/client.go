package config

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
)

// Client interface
type Client[C schemas.BaseConfiguration] interface {
	GetConfig(*viper.Viper) C
	SetAndBindFlag(string, interface{}, string) error
}

// GetConfigFunc type
type GetConfigFunc func(*ConfigClient, *viper.Viper) *schemas.GlobalConfiguration

// SetAndBindFlagFunc type
type SetAndBindFlagFunc func(*ConfigClient, string, interface{}, string) error

// ConfigClient is composed of functions to interact with configuration.
type ConfigClient struct {
	getConfig      GetConfigFunc
	setAndBindFlag SetAndBindFlagFunc
	loggingClient  *bbLog.Client
	command        *cobra.Command
}

// NewClient returns a new config client with the provided configuration
func NewClient(
	getConfig GetConfigFunc,
	setAndBindFlag SetAndBindFlagFunc,
	loggingClient *bbLog.Client,
	command *cobra.Command,
) (*ConfigClient, error) {
	if loggingClient == nil {
		return nil, errors.New("logging client is required")
	}
	if command == nil {
		return nil, errors.New("command is required")
	}
	return &ConfigClient{
		getConfig:      getConfig,
		setAndBindFlag: setAndBindFlag,
		loggingClient:  loggingClient,
		command:        command,
	}, nil
}

// GetConfig returns the global configuration.
func (client *ConfigClient) GetConfig(viper *viper.Viper) *schemas.GlobalConfiguration {
	return client.getConfig(client, viper)
}

// SetAndBindFlag sets and binds a flag to a command and viper.
func (client *ConfigClient) SetAndBindFlag(name string, value interface{}, description string) error {
	return client.setAndBindFlag(client, name, value, description)
}
