package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReconcileConfiguration_PreflightCheckConfiguration(t *testing.T) {
	var tests = []struct {
		desc                string
		arg                 *PreflightCheckConfiguration
		setRegistryServer   bool
		setRegistryUsername bool
		setRegistryPassword bool
	}{
		{
			"reconcile configuration, no values",
			&PreflightCheckConfiguration{},
			false,
			false,
			false,
		},
		{
			"reconcile configuration, registry server set",
			&PreflightCheckConfiguration{},
			true,
			false,
			false,
		},
		{
			"reconcile configuration, registry username set",
			&PreflightCheckConfiguration{},
			false,
			true,
			false,
		},
		{
			"reconcile configuration, registry password set",
			&PreflightCheckConfiguration{},
			false,
			false,
			true,
		},
		{
			"reconcile configuration, registry server, username, and password set",
			&PreflightCheckConfiguration{},
			true,
			true,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			registryServer := "test"
			registryUsername := "test"
			registryPassword := "test"
			instance := viper.New()
			if tt.setRegistryServer {
				instance.Set("registryserver", registryServer)
			}
			if tt.setRegistryUsername {
				instance.Set("registryusername", registryUsername)
			}
			if tt.setRegistryPassword {
				instance.Set("registrypassword", registryPassword)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.Nil(t, err)
			if tt.setRegistryServer {
				assert.Equal(t, registryServer, instance.GetString("registryserver"))
				assert.Equal(t, registryServer, tt.arg.RegistryServer)
			} else {
				assert.Equal(t, "", instance.GetString("registryserver"))
				assert.Equal(t, "", tt.arg.RegistryServer)
			}
			if tt.setRegistryUsername {
				assert.Equal(t, registryUsername, instance.GetString("registryusername"))
				assert.Equal(t, registryUsername, tt.arg.RegistryUsername)
			} else {
				assert.Equal(t, "", instance.GetString("registryusername"))
				assert.Equal(t, "", tt.arg.RegistryUsername)
			}
			if tt.setRegistryPassword {
				assert.Equal(t, registryPassword, instance.GetString("registrypassword"))
				assert.Equal(t, registryPassword, tt.arg.RegistryPassword)
			} else {
				assert.Equal(t, "", instance.GetString("registrypassword"))
				assert.Equal(t, "", tt.arg.RegistryPassword)
			}
		})
	}
}

func TestGetSubConfigurations_PreflightCheckConfiguration(t *testing.T) {
	// Arrange
	preflightCheckConfiguration := &PreflightCheckConfiguration{}
	// Act
	subConfigurations := preflightCheckConfiguration.getSubConfigurations()
	// Assert
	assert.Equal(t, 0, len(subConfigurations))
}
