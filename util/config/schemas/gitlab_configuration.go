package schemas

import "github.com/spf13/viper"

type GitLabConfiguration struct {
	// access-token used for GitLab auth
	Token   string `mapstructure:"access-token" yaml:"access-token"`
	BaseURL string `mapstructure:"base-url"     yaml:"base-url"`
}

func (g *GitLabConfiguration) ReconcileConfiguration(instance *viper.Viper) error {
	if instance.IsSet("gitlab-access-token") {
		g.Token = instance.GetString("gitlab-access-token")
	}
	g.BaseURL = "https://repo1.dso.mil/api/v4"
	if instance.IsSet("gitlab-base-url") {
		g.BaseURL = instance.GetString("gitlab-base-url")
	}
	return nil
}

// getSubConfigurations returns the sub-configurations.
func (g *GitLabConfiguration) getSubConfigurations() []BaseConfiguration {
	return []BaseConfiguration{}
}
