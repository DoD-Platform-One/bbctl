package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcileConfiguration_K3dSshConfiguration(t *testing.T) {
	var tests = []struct {
		desc    string
		arg     *K3dSSHConfiguration
		setUser bool
	}{
		{
			"reconcile configuration, no values",
			&K3dSSHConfiguration{},
			false,
		},
		{
			"reconcile configuration, user set",
			&K3dSSHConfiguration{},
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
			require.NoError(t, err)
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
	k3dSSHConfiguration := &K3dSSHConfiguration{}
	v := viper.New()
	// Act
	require.NoError(t, k3dSSHConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.Equal(t, "ubuntu", k3dSSHConfiguration.User)
}

func TestReconcileConfigurationSetOutsideOfViper_K3dSshConfiguration(t *testing.T) {
	// Arrange
	k3dSSHConfiguration := &K3dSSHConfiguration{
		User: "test",
	}
	v := viper.New()
	// Act
	require.NoError(t, k3dSSHConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.Equal(t, "test", k3dSSHConfiguration.User)
}

func TestGetSubConfigurations_K3dSshConfiguration(t *testing.T) {
	// Arrange
	k3dSSHConfiguration := &K3dSSHConfiguration{}
	// Act
	subConfigurations := k3dSSHConfiguration.getSubConfigurations()
	// Assert
	assert.Empty(t, subConfigurations)
}
