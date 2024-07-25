package schemas

import (
	"errors"

	"github.com/spf13/viper"
)

type GlobalConfiguration struct {
	// Big Bang repository location: file path
	BigBangRepo string `mapstructure:"big-bang-repo" yaml:"big-bang-repo" validate:"required"`
	// Deploy Big Bang configuration: object
	DeployBigBangConfiguration DeployBigBangConfiguration `mapstructure:"deploy-big-bang" yaml:"deploy-big-bang"`
	// Example configuration: object
	ExampleConfiguration ExampleConfiguration `mapstructure:"example-config" yaml:"example-config"`
	// K3d SSH configuration: object
	K3dSshConfiguration K3dSshConfiguration `mapstructure:"k3d-ssh" yaml:"k3d-ssh"`
	// Add source to log: boolean
	LogAddSource bool `mapstructure:"bbctl-log-add-source" yaml:"bbctl-log-add-source"`
	// GitLab configuration: object
	GitLabConfiguration GitLabConfiguration `mapstructure:"gitlab" yaml:"gitlab"`
	// Log file location: file path
	LogFile string `mapstructure:"bbctl-log-file" yaml:"bbctl-log-file"`
	// Log format: json, text
	LogFormat string `mapstructure:"bbctl-log-format" yaml:"bbctl-log-format"`
	// Log level: debug, info, warn, error
	LogLevel string `mapstructure:"bbctl-log-level" yaml:"bbctl-log-level"`
	// Log output: stdout, stderr
	LogOutput string `mapstructure:"bbctl-log-output" yaml:"bbctl-log-output"`
	// Policy configuration: object
	PolicyConfiguration PolicyConfiguration `mapstructure:"policy" yaml:"policy"`
	// Preflight check configuration: object
	PreflightCheckConfiguration PreflightCheckConfiguration `mapstructure:"preflight-check" yaml:"preflight-check"`
	// Util credential helper configuration: object
	UtilCredentialHelperConfiguration UtilCredentialHelperConfiguration `mapstructure:"util-credential-helper" yaml:"util-credential-helper"`
	// Util k8s configuration: object
	UtilK8sConfiguration UtilK8sConfiguration `mapstructure:"util-k8s" yaml:"util-k8s"`
	// Version configuration: object
	VersionConfiguration VersionConfiguration `mapstructure:"version" yaml:"version"`
	// Violations configuration: object
	ViolationsConfiguration ViolationsConfiguration `mapstructure:"violation" yaml:"violation"`
}

// ReconcileConfiguration recursively reconciles the configurations.
func (g *GlobalConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	g.BigBangRepo = instance.GetString("big-bang-repo")
	g.LogAddSource = instance.GetBool("bbctl-log-add-source")
	g.LogFile = instance.GetString("bbctl-log-file")
	g.LogFormat = instance.GetString("bbctl-log-format")
	g.LogLevel = instance.GetString("bbctl-log-level")
	g.LogOutput = instance.GetString("bbctl-log-output")

	allErrors := []error{}
	for _, subConfig := range g.getSubConfigurations() {
		err := subConfig.ReconcileConfiguration(instance)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}
	return errors.Join(allErrors...)
}

// getSubConfigurations returns the sub-configurations.
func (g *GlobalConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{
		&g.DeployBigBangConfiguration,
		&g.ExampleConfiguration,
		&g.K3dSshConfiguration,
		&g.GitLabConfiguration,
		&g.PolicyConfiguration,
		&g.PreflightCheckConfiguration,
		&g.UtilCredentialHelperConfiguration,
		&g.UtilK8sConfiguration,
		&g.VersionConfiguration,
		&g.ViolationsConfiguration,
	}
}
