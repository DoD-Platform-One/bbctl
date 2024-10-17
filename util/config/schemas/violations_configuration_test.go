package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcileConfiguration_ViolationsConfiguration(t *testing.T) {
	var tests = []struct {
		desc      string
		arg       *ViolationsConfiguration
		AuditMode bool
	}{
		{
			"reconcile configuration, no values",
			&ViolationsConfiguration{},
			false,
		},
		{
			"reconcile configuration, audit mode set",
			&ViolationsConfiguration{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			instance := viper.New()
			if tt.AuditMode {
				instance.Set("audit", true)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.AuditMode, tt.arg.Audit)
		})
	}
}

func TestReconcileConfigurationDefaults_ViolationsConfiguration(t *testing.T) {
	// Arrange
	violationsConfiguration := &ViolationsConfiguration{}
	v := viper.New()
	// Act
	require.NoError(t, violationsConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.False(t, violationsConfiguration.Audit)
}

func TestGetSubConfigurations_ViolationsConfiguration(t *testing.T) {
	// Arrange
	violationsConfiguration := &ViolationsConfiguration{}
	// Act
	subConfigurations := violationsConfiguration.getSubConfigurations()
	// Assert
	assert.Empty(t, subConfigurations)
}
