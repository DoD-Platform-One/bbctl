package schemas

import "github.com/spf13/viper"

type UtilCredentialHelperConfiguration struct {
	// credentials file path: file path
	FilePath string `mapstructure:"big-bang-credential-helper-credentials-file-path" yaml:"big-bang-credential-helper-credentials-file-path"`
	// credential helper: file path OR credentials-file
	CredentialHelper string `mapstructure:"big-bang-credential-helper" yaml:"big-bang-credential-helper"`
}

// ReconcileConfiguration reconciles the configuration.
func (u *UtilCredentialHelperConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("big-bang-credential-helper-credentials-file-path") {
		u.FilePath = instance.GetString("big-bang-credential-helper-credentials-file-path")
	}
	if instance.IsSet("big-bang-credential-helper") {
		u.CredentialHelper = instance.GetString("big-bang-credential-helper")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (u *UtilCredentialHelperConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
