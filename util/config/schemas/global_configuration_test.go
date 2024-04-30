package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestReconcileConfiguration_GlobalConfiguration(t *testing.T) {
	var tests = []struct {
		desc         string
		arg          *GlobalConfiguration
		willError    bool
		errorMessage string
	}{
		{
			"reconcile configuration, pass",
			&GlobalConfiguration{
				DeployBigBangConfiguration:        DeployBigBangConfiguration{},
				ExampleConfiguration:              ExampleConfiguration{},
				K3dSshConfiguration:               K3dSshConfiguration{},
				PreflightCheckConfiguration:       PreflightCheckConfiguration{},
				UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
				UtilK8sConfiguration:              UtilK8sConfiguration{},
			},
			false,
			"",
		},
		{
			"reconcile configuration, fail",
			&GlobalConfiguration{
				DeployBigBangConfiguration:        DeployBigBangConfiguration{},
				ExampleConfiguration:              ExampleConfiguration{ShouldError: true},
				K3dSshConfiguration:               K3dSshConfiguration{},
				PreflightCheckConfiguration:       PreflightCheckConfiguration{},
				UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
				UtilK8sConfiguration:              UtilK8sConfiguration{},
			},
			true,
			"should error was set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			instance := viper.New()
			instance.Set("big-bang-repo", "test1")                                    // root
			instance.Set("k3d", true)                                                 // DeployBigBangConfiguration
			instance.Set("ssh-username", "test2")                                     // K3dSshConfiguration
			instance.Set("registryserver", "test3")                                   // PreflightCheckConfiguration
			instance.Set("big-bang-credential-helper-credentials-file-path", "test4") // UtilCredentialHelperConfiguration
			instance.Set("kubeconfig", "test")                                        // UtilK8sConfiguration
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			if tt.willError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				// we can't check the values because we don't know what they are because we don't know where it errored
			} else {
				assert.Nil(t, err)
				assert.Equal(t, "", tt.arg.BigBangRepo) // this would normally be loaded by viper, but we aren't loading a config file
				assert.Equal(t, true, tt.arg.DeployBigBangConfiguration.K3d)
				assert.Equal(t, "test2", tt.arg.K3dSshConfiguration.User)
				assert.Equal(t, "test3", tt.arg.PreflightCheckConfiguration.RegistryServer)
				assert.Equal(t, "test4", tt.arg.UtilCredentialHelperConfiguration.FilePath)
				assert.Equal(t, "test", tt.arg.UtilK8sConfiguration.Kubeconfig)
			}
		})
	}
}

func TestGetSubConfigurations_GlobalConfiguration(t *testing.T) {
	// Arrange
	arg := &GlobalConfiguration{
		DeployBigBangConfiguration:        DeployBigBangConfiguration{},
		ExampleConfiguration:              ExampleConfiguration{},
		K3dSshConfiguration:               K3dSshConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
	}
	// Act
	result := arg.getSubConfigurations()
	// Assert
	assert.Equal(t, 6, len(result))
	assert.Equal(t, &arg.DeployBigBangConfiguration, result[0])
	assert.Equal(t, &arg.ExampleConfiguration, result[1])
	assert.Equal(t, &arg.K3dSshConfiguration, result[2])
	assert.Equal(t, &arg.PreflightCheckConfiguration, result[3])
	assert.Equal(t, &arg.UtilCredentialHelperConfiguration, result[4])
	assert.Equal(t, &arg.UtilK8sConfiguration, result[5])
}
