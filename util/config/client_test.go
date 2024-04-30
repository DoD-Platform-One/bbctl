package config

import (
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
)

// NewClient tested in client_getter_test.go

func TestClientGetConfig(t *testing.T) {
	// Arrange
	expected := "test"
	getConfigFunc := func(client *ConfigClient, viper *viper.Viper) *schemas.GlobalConfiguration {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}
	}
	client := &ConfigClient{
		getConfig: getConfigFunc,
	}

	// Act
	actual := client.GetConfig(nil)

	// Assert
	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual.BigBangRepo)
}

func TestClientSetAndBindFlag(t *testing.T) {
	// Arrange
	expected := "test"
	setAndBindFlagFunc := func(client *ConfigClient, name string, value interface{}, description string) error {
		return errors.New(expected)
	}
	client := &ConfigClient{
		setAndBindFlag: setAndBindFlagFunc,
	}

	// Act
	actual := client.SetAndBindFlag("", "", "")

	// Assert
	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual.Error())
}
