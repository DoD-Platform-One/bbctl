package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReconcileConfiguration_K3dSshConfiguration(t *testing.T) {
	var tests = []struct {
		desc string
		arg  *K3dSshConfiguration
	}{
		{
			"reconcile configuration, no values",
			&K3dSshConfiguration{},
		},
		{
			"reconcile configuration, user set",
			&K3dSshConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			username := "test"
			instance := viper.New()
			instance.Set("ssh-username", username)
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.Nil(t, err)
			assert.Equal(t, username, instance.GetString("ssh-username"))
		})
	}
}

func TestGetSubConfigurations_K3dSshConfiguration(t *testing.T) {
	// Arrange
	k3dSshConfiguration := &K3dSshConfiguration{}
	// Act
	subConfigurations := k3dSshConfiguration.getSubConfigurations()
	// Assert
	assert.Equal(t, 0, len(subConfigurations))
}
