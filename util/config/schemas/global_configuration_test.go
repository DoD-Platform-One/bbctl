package schemas

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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
				K3dSshConfiguration:               K3dSshConfiguration{},
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
				K3dSshConfiguration:               K3dSshConfiguration{},
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
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				// we can't check the values because we don't know what they are because we don't know where it errored
			} else {
				assert.Nil(t, err)
				assert.Equal(t, "test1", tt.arg.BigBangRepo)
				assert.Equal(t, true, tt.arg.DeployBigBangConfiguration.K3d)
				assert.Equal(t, "test2", tt.arg.K3dSshConfiguration.User)
				assert.Equal(t, "token", tt.arg.GitLabConfiguration.Token)
				assert.Equal(t, true, tt.arg.PolicyConfiguration.Gatekeeper)
				assert.Equal(t, "test3", tt.arg.PreflightCheckConfiguration.RegistryServer)
				assert.Equal(t, "test4", tt.arg.UtilCredentialHelperConfiguration.FilePath)
				assert.Equal(t, "test", tt.arg.UtilK8sConfiguration.Kubeconfig)
				assert.Equal(t, true, tt.arg.VersionConfiguration.Client)
				assert.Equal(t, true, tt.arg.ViolationsConfiguration.Audit)
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
		K3dSshConfiguration:               K3dSshConfiguration{},
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
	assert.Equal(t, 10, len(result))
	assert.Equal(t, &arg.DeployBigBangConfiguration, result[0])
	assert.Equal(t, &arg.ExampleConfiguration, result[1])
	assert.Equal(t, &arg.K3dSshConfiguration, result[2])
	assert.Equal(t, &arg.GitLabConfiguration, result[3])
	assert.Equal(t, &arg.PolicyConfiguration, result[4])
	assert.Equal(t, &arg.PreflightCheckConfiguration, result[5])
	assert.Equal(t, &arg.UtilCredentialHelperConfiguration, result[6])
	assert.Equal(t, &arg.UtilK8sConfiguration, result[7])
	assert.Equal(t, &arg.VersionConfiguration, result[8])
	assert.Equal(t, &arg.ViolationsConfiguration, result[9])
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
		K3dSshConfiguration:               K3dSshConfiguration{},
		PolicyConfiguration:               PolicyConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
		VersionConfiguration:              VersionConfiguration{},
		ViolationsConfiguration:           ViolationsConfiguration{},
	}
	// Act
	result, _ := arg.MarshalYaml()
	var unmarshalled GlobalConfiguration
	err := yaml.Unmarshal(result, &unmarshalled)

	// Assert
	assert.Nil(t, err)
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
		K3dSshConfiguration:               K3dSshConfiguration{},
		PolicyConfiguration:               PolicyConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
		VersionConfiguration:              VersionConfiguration{},
		ViolationsConfiguration:           ViolationsConfiguration{},
	}
	// Act
	result, _ := arg.MarshalJson()
	var unmarshalled GlobalConfiguration
	err := json.Unmarshal(result, &unmarshalled)

	// Assert
	assert.Nil(t, err)
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
		K3dSshConfiguration:               K3dSshConfiguration{},
		PolicyConfiguration:               PolicyConfiguration{},
		PreflightCheckConfiguration:       PreflightCheckConfiguration{},
		UtilCredentialHelperConfiguration: UtilCredentialHelperConfiguration{},
		UtilK8sConfiguration:              UtilK8sConfiguration{},
		VersionConfiguration:              VersionConfiguration{},
		ViolationsConfiguration:           ViolationsConfiguration{},
	}
	// Act
	result, err := arg.MarshalHumanReadable()

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("%#v", arg), string(result))
}
