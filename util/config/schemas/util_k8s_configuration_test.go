package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReconcileConfiguration_UtilK8sConfiguration(t *testing.T) {
	var tests = []struct {
		desc string
		arg  *UtilK8sConfiguration
	}{
		{
			"reconcile configuration, no values",
			&UtilK8sConfiguration{},
		},
		{
			"reconcile configuration, kubeconfig set",
			&UtilK8sConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			kubeconfigPath := "test"
			instance := viper.New()
			instance.Set("kubeconfig", kubeconfigPath)
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.Nil(t, err)
			assert.Equal(t, kubeconfigPath, tt.arg.Kubeconfig)
		})
	}
}

func TestGetSubConfigurations_UtilK8sConfiguration(t *testing.T) {
	// Arrange
	arg := &UtilK8sConfiguration{}
	// Act
	result := arg.getSubConfigurations()
	// Assert
	assert.Equal(t, []BaseConfiguration{}, result)
}
