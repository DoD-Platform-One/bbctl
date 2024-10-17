package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReconcileConfiguration_UtilCredentialHelperConfiguration(t *testing.T) {
	var tests = []struct {
		desc                string
		arg                 *UtilCredentialHelperConfiguration
		setFilePath         bool
		setCredentialHelper bool
	}{
		{
			"reconcile configuration, no values",
			&UtilCredentialHelperConfiguration{},
			false,
			false,
		},
		{
			"reconcile configuration, file path set",
			&UtilCredentialHelperConfiguration{},
			true,
			false,
		},
		{
			"reconcile configuration, credential helper set",
			&UtilCredentialHelperConfiguration{},
			false,
			true,
		},
		{
			"reconcile configuration, file path and credential helper set",
			&UtilCredentialHelperConfiguration{},
			true,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			filePath := "test"
			credentialHelper := "test"
			instance := viper.New()
			if tt.setFilePath {
				instance.Set("big-bang-credential-helper-credentials-file-path", filePath)
			}
			if tt.setCredentialHelper {
				instance.Set("big-bang-credential-helper", credentialHelper)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			require.NoError(t, err)
			if tt.setFilePath {
				assert.Equal(t, filePath, tt.arg.FilePath)
			} else {
				assert.Equal(t, "", tt.arg.FilePath)
			}
			if tt.setCredentialHelper {
				assert.Equal(t, credentialHelper, tt.arg.CredentialHelper)
			} else {
				assert.Equal(t, "", tt.arg.CredentialHelper)
			}
		})
	}
}

func TestGetSubConfigurations_UtilCredentialHelperConfiguration(t *testing.T) {
	// Arrange
	arg := &UtilCredentialHelperConfiguration{}
	// Act
	result := arg.getSubConfigurations()
	// Assert
	assert.Equal(t, []BaseConfiguration{}, result)
}
