package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcileConfiguration_VersionConfiguration(t *testing.T) {
	var tests = []struct {
		desc      string
		arg       *VersionConfiguration
		setClient bool
	}{
		{
			"reconcile configuration, no values",
			&VersionConfiguration{},
			false,
		},
		{
			"reconcile configuration, client set",
			&VersionConfiguration{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			instance := viper.New()
			if tt.setClient {
				instance.Set("client", true)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.setClient, tt.arg.Client)
		})
	}
}

func TestReconcileConfigurationDefaults_VersionConfiguration(t *testing.T) {
	// Arrange
	policyConfiguration := &VersionConfiguration{}
	v := viper.New()
	// Act
	require.NoError(t, policyConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.False(t, policyConfiguration.Client)
}

func TestGetSubConfigurations_VersionConfiguration(t *testing.T) {
	// Arrange
	versionConfiguration := &VersionConfiguration{}
	// Act
	subConfigurations := versionConfiguration.getSubConfigurations()
	// Assert
	assert.Empty(t, subConfigurations)
}
