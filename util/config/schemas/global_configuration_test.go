package schemas

import (
	"encoding/json"
	"fmt"
	"testing"

	"repo1.dso.mil/big-bang/product/packages/bbctl/util/yamler"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				GitLabConfiguration:               GitLabConfiguration{},
				K3dSSHConfiguration:               K3dSSHConfiguration{},
				OutputConfiguration:               OutputConfiguration{},
				PolicyConfiguration:               PolicyConfiguration{},
				PreflightCheckConfiguration:       PreflightCheckConfiguration{},
				UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
				UtilK8sConfiguration:              UtilK8sConfiguration{},
				VersionConfiguration:              VersionConfiguration{},
				ViolationsConfiguration:           ViolationsConfiguration{},
			},
			false,
			"",
		},
		{
			"reconcile configuration, fail",
			&GlobalConfiguration{
				DeployBigBangConfiguration:        DeployBigBangConfiguration{},
				ExampleConfiguration:              ExampleConfiguration{},
				GitLabConfiguration:               GitLabConfiguration{},
				K3dSSHConfiguration:               K3dSSHConfiguration{},
				OutputConfiguration:               OutputConfiguration{},
				PolicyConfiguration:               PolicyConfiguration{},
				PreflightCheckConfiguration:       PreflightCheckConfiguration{},
				UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
				UtilK8sConfiguration:              UtilK8sConfiguration{},
				VersionConfiguration:              VersionConfiguration{},
				ViolationsConfiguration:           ViolationsConfiguration{},
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
			instance.Set("gitlab-access-token", "token")                              // GitLabConfiguration
			instance.Set("ssh-username", "test2")                                     // K3dSshConfiguration
			instance.Set("format", "json")                                            // OutputConfiguration
			instance.Set("gatekeeper", true)                                          // PolicyConfiguration
			instance.Set("registryserver", "test3")                                   // PreflightCheckConfiguration
			instance.Set("big-bang-credential-helper-credentials-file-path", "test4") // UtilCredentialHelperConfiguration
			instance.Set("kubeconfig", "test")                                        // UtilK8sConfiguration
			instance.Set("client", true)                                              // VersionConfiguration
			instance.Set("audit", true)                                               // ViolationsConfiguration
			if tt.willError {
				instance.Set("example-config-should-error", true)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			if tt.willError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				// we can't check the values because we don't know what they are because we don't know where it errored
			} else {
				require.NoError(t, err)
				assert.Equal(t, "test1", tt.arg.BigBangRepo)
				assert.True(t, tt.arg.DeployBigBangConfiguration.K3d)
				assert.Equal(t, "test2", tt.arg.K3dSSHConfiguration.User)
				assert.Equal(t, "token", tt.arg.GitLabConfiguration.Token)
				assert.Equal(t, "json", string(tt.arg.OutputConfiguration.Format))
				assert.True(t, tt.arg.PolicyConfiguration.Gatekeeper)
				assert.Equal(t, "test3", tt.arg.PreflightCheckConfiguration.RegistryServer)
				assert.Equal(t, "test4", tt.arg.UtilCredentialHelperConfiguration.FilePath)
				assert.Equal(t, "test", tt.arg.UtilK8sConfiguration.Kubeconfig)
				assert.True(t, tt.arg.VersionConfiguration.Client)
				assert.True(t, tt.arg.ViolationsConfiguration.Audit)
			}
		})
	}
}

func TestGetSubConfigurations_GlobalConfiguration(t *testing.T) {
	// Arrange
	arg := &GlobalConfiguration{
		DeployBigBangConfiguration:        DeployBigBangConfiguration{},
		ExampleConfiguration:              ExampleConfiguration{},
		GitLabConfiguration:               GitLabConfiguration{},
		K3dSSHConfiguration:               K3dSSHConfiguration{},
		OutputConfiguration:               OutputConfiguration{},
		PolicyConfiguration:               PolicyConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
		VersionConfiguration:              VersionConfiguration{},
		ViolationsConfiguration:           ViolationsConfiguration{},
	}
	// Act
	result := arg.getSubConfigurations()
	// Assert
	assert.Len(t, result, 11)
	assert.Equal(t, &arg.DeployBigBangConfiguration, result[0])
	assert.Equal(t, &arg.ExampleConfiguration, result[1])
	assert.Equal(t, &arg.GitLabConfiguration, result[2])
	assert.Equal(t, &arg.K3dSSHConfiguration, result[3])
	assert.Equal(t, &arg.OutputConfiguration, result[4])
	assert.Equal(t, &arg.PolicyConfiguration, result[5])
	assert.Equal(t, &arg.PreflightCheckConfiguration, result[6])
	assert.Equal(t, &arg.UtilCredentialHelperConfiguration, result[7])
	assert.Equal(t, &arg.UtilK8sConfiguration, result[8])
	assert.Equal(t, &arg.VersionConfiguration, result[9])
	assert.Equal(t, &arg.ViolationsConfiguration, result[10])
}

func TestGetYamlMarshalling(t *testing.T) {
	// Arrange
	bbConfig := DeployBigBangConfiguration{
		K3d:   false,
		Addon: []string{},
	}

	arg := &GlobalConfiguration{
		DeployBigBangConfiguration:        bbConfig,
		ExampleConfiguration:              ExampleConfiguration{},
		GitLabConfiguration:               GitLabConfiguration{},
		K3dSSHConfiguration:               K3dSSHConfiguration{},
		OutputConfiguration:               OutputConfiguration{},
		PolicyConfiguration:               PolicyConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
		VersionConfiguration:              VersionConfiguration{},
		ViolationsConfiguration:           ViolationsConfiguration{},
	}
	// Act
	result, _ := arg.EncodeYAML()
	var unmarshalled GlobalConfiguration
	err := yamler.Unmarshal(result, &unmarshalled)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, arg.BigBangRepo, unmarshalled.BigBangRepo)
	assert.Equal(t, bbConfig, unmarshalled.DeployBigBangConfiguration)
}

func TestGetJsonMarshalling(t *testing.T) {
	// Arrange
	bbConfig := DeployBigBangConfiguration{
		K3d:   false,
		Addon: []string{},
	}

	arg := &GlobalConfiguration{
		DeployBigBangConfiguration:        bbConfig,
		ExampleConfiguration:              ExampleConfiguration{},
		GitLabConfiguration:               GitLabConfiguration{},
		K3dSSHConfiguration:               K3dSSHConfiguration{},
		OutputConfiguration:               OutputConfiguration{},
		PolicyConfiguration:               PolicyConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
		VersionConfiguration:              VersionConfiguration{},
		ViolationsConfiguration:           ViolationsConfiguration{},
	}
	// Act
	result, _ := arg.EncodeJSON()
	var unmarshalled GlobalConfiguration
	err := json.Unmarshal(result, &unmarshalled) //nolint:musttag

	// Assert
	require.NoError(t, err)
	assert.Equal(t, arg.BigBangRepo, unmarshalled.BigBangRepo)
	assert.Equal(t, bbConfig, unmarshalled.DeployBigBangConfiguration)
}

func TestGetTextMarshalling(t *testing.T) {
	// Arrange
	bbConfig := DeployBigBangConfiguration{
		K3d:   false,
		Addon: []string{},
	}

	arg := &GlobalConfiguration{
		DeployBigBangConfiguration:        bbConfig,
		ExampleConfiguration:              ExampleConfiguration{},
		GitLabConfiguration:               GitLabConfiguration{},
		K3dSSHConfiguration:               K3dSSHConfiguration{},
		OutputConfiguration:               OutputConfiguration{},
		PolicyConfiguration:               PolicyConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
		VersionConfiguration:              VersionConfiguration{},
		ViolationsConfiguration:           ViolationsConfiguration{},
	}
	// Act
	result, err := arg.EncodeText()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%#v", arg), string(result))
}
