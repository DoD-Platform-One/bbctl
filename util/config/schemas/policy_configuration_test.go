package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReconcileConfiguration_PolicyConfiguration(t *testing.T) {
	var tests = []struct {
		desc          string
		arg           *PolicyConfiguration
		useGatekeeper bool
		useKyverno    bool
	}{
		{
			"reconcile configuration, no values",
			&PolicyConfiguration{},
			false,
			false,
		},
		{
			"reconcile configuration, gatekeeper set",
			&PolicyConfiguration{},
			true,
			false,
		},
		{
			"reconcile configuration, kyverno set",
			&PolicyConfiguration{},
			false,
			true,
		},
		{
			"reconcile configuration, gatekeeper and kyverno set",
			&PolicyConfiguration{},
			true,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			instance := viper.New()
			if tt.useGatekeeper {
				instance.Set("gatekeeper", true)
			}
			if tt.useKyverno {
				instance.Set("kyverno", true)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.Nil(t, err)
			assert.Equal(t, tt.useGatekeeper, tt.arg.Gatekeeper)
			assert.Equal(t, tt.useKyverno, tt.arg.Kyverno)
		})
	}
}

func TestReconcileConfigurationDefaults_PolicyConfiguration(t *testing.T) {
	// Arrange
	policyConfiguration := &PolicyConfiguration{}
	v := viper.New()
	// Act
	assert.Nil(t, policyConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.Equal(t, false, policyConfiguration.Gatekeeper)
	assert.Equal(t, false, policyConfiguration.Kyverno)
}

func TestGetSubConfigurations_PolicyConfiguration(t *testing.T) {
	// Arrange
	policyConfiguration := &PolicyConfiguration{}
	// Act
	subConfigurations := policyConfiguration.getSubConfigurations()
	// Assert
	assert.Equal(t, 0, len(subConfigurations))
}
