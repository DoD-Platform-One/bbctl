package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
)

func TestReconcileOutputConfigurations(t *testing.T) {
	var tests = []struct {
		desc      string
		arg       *OutputConfiguration
		format    output.OutputFormat
		setFormat bool
	}{
		{
			"no configs and no args",
			&OutputConfiguration{},
			"",
			false,
		},
		{
			"format config with no args",
			&OutputConfiguration{Format: "JSON"},
			"JSON",
			false,
		},
		{
			"empty config with format args",
			&OutputConfiguration{},
			"JSON",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			instance := viper.New()
			if tt.setFormat {
				instance.Set("format", string(tt.format))
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.NoError(t, err)
			if tt.format != "" {
				assert.Equal(t, tt.format, tt.arg.Format)
			}
		})
	}
}

func TestReconcileOutputConfigurationDefaults(t *testing.T) {
	// Arrange
	outputConfiguration := &OutputConfiguration{}
	v := viper.New()
	// Act
	assert.Nil(t, outputConfiguration.ReconcileConfiguration(v))
	// Assert
	assert.Equal(t, output.TEXT, outputConfiguration.Format)
}

func TestGetSubConfigurations(t *testing.T) {
	// Arrange
	outputConfiguration := &OutputConfiguration{}
	// Act
	subConfigs := outputConfiguration.getSubConfigurations()
	// Assert
	assert.Empty(t, subConfigs)
}
