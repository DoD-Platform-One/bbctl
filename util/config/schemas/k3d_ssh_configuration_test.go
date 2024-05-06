package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReconcileConfiguration_K3dSshConfiguration(t *testing.T) {
	var tests = []struct {
		desc    string
		arg     *K3dSshConfiguration
		setUser bool
	}{
		{
			"reconcile configuration, no values",
			&K3dSshConfiguration{},
			false,
		},
		{
			"reconcile configuration, user set",
			&K3dSshConfiguration{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			username := "test"
			instance := viper.New()
			if tt.setUser {
				instance.Set("ssh-username", username)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.Nil(t, err)
			if tt.setUser {
				assert.Equal(t, username, tt.arg.User)
			} else {
				assert.Equal(t, "ubuntu", tt.arg.User)
			}
		})
	}
}

func TestReconcileConfigurationDefaults_K3dSshConfiguration(t *testing.T) {
	// Arrange
	k3dSshConfiguration := &K3dSshConfiguration{}
	v := viper.New()
	// Act
	assert.Nil(t, k3dSshConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.Equal(t, "ubuntu", k3dSshConfiguration.User)
}

func TestReconcileConfigurationSetOutsideOfViper_K3dSshConfiguration(t *testing.T) {
	// Arrange
	k3dSshConfiguration := &K3dSshConfiguration{
		User: "test",
	}
	v := viper.New()
	// Act
	assert.Nil(t, k3dSshConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.Equal(t, "test", k3dSshConfiguration.User)
}

func TestGetSubConfigurations_K3dSshConfiguration(t *testing.T) {
	// Arrange
	k3dSshConfiguration := &K3dSshConfiguration{}
	// Act
	subConfigurations := k3dSshConfiguration.getSubConfigurations()
	// Assert
	assert.Equal(t, 0, len(subConfigurations))
}
