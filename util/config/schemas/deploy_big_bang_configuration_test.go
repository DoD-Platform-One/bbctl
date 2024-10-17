package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcileConfiguration_DeployBigBangConfiguration(t *testing.T) {
	var tests = []struct {
		desc     string
		arg      *DeployBigBangConfiguration
		k3d      bool
		setK3d   bool
		addon    []string
		setAddon bool
	}{
		{
			"reconcile configuration, no values",
			&DeployBigBangConfiguration{},
			false,
			false,
			[]string{},
			false,
		},
		{
			"reconcile configuration, k3d set",
			&DeployBigBangConfiguration{},
			true,
			true,
			[]string{},
			false,
		},
		{
			"reconcile configuration, addon set",
			&DeployBigBangConfiguration{},
			false,
			false,
			[]string{"test"},
			true,
		},
		{
			"reconcile configuration, k3d and addon set",
			&DeployBigBangConfiguration{},
			true,
			true,
			[]string{"test"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			instance := viper.New()
			if tt.setK3d {
				instance.Set("k3d", tt.k3d)
			}
			if tt.setAddon {
				instance.Set("addon", tt.addon)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			require.NoError(t, err)
			if tt.setK3d {
				assert.Equal(t, tt.k3d, tt.arg.K3d)
			} else {
				assert.False(t, tt.arg.K3d)
			}
			if tt.setAddon {
				assert.Equal(t, tt.addon, tt.arg.Addon)
			} else {
				assert.Equal(t, []string(nil), tt.arg.Addon)
			}
		})
	}
}

func TestGetSubConfigurations_DeployBigBangConfiguration(t *testing.T) {
	// Arrange
	configuration := &DeployBigBangConfiguration{}
	// Act
	result := configuration.getSubConfigurations()
	// Assert
	assert.Equal(t, []BaseConfiguration{}, result)
}
