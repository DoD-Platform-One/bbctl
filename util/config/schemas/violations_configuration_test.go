package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
			assert.Nil(t, err)
			assert.Equal(t, tt.AuditMode, tt.arg.Audit)
		})
	}
}

func TestReconcileConfigurationDefaults_ViolationsConfiguration(t *testing.T) {
	// Arrange
	policyConfiguration := &ViolationsConfiguration{}
	v := viper.New()
	// Act
	assert.Nil(t, policyConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.Equal(t, false, policyConfiguration.Audit)
}

func TestGetSubConfigurations_ViolationsConfiguration(t *testing.T) {
	// Arrange
	policyConfiguration := &PolicyConfiguration{}
	// Act
	subConfigurations := policyConfiguration.getSubConfigurations()
	// Assert
	assert.Equal(t, 0, len(subConfigurations))
}
