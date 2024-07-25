package schemas

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetGitLabConfigurations(t *testing.T) {
	var tests = []struct {
		desc       string
		arg        *GitLabConfiguration
		token      string
		setToken   bool
		baseURL    string
		setBaseURL bool
	}{
		{
			"no configs and no args",
			&GitLabConfiguration{},
			"",
			false,
			"",
			false,
		},
		{
			"token and url config with no args",
			&GitLabConfiguration{Token: "qnxuwoei", BaseURL: "https://localhost"},
			"qnxuwoei",
			false,
			"https://localhost",
			false,
		},
		{
			"empty config with both token and url args",
			&GitLabConfiguration{},
			"qnxuwoei",
			true,
			"https://localhost",
			true,
		},
		{
			"split config token and url arg",
			&GitLabConfiguration{Token: "qnxuwoei"},
			"qnxuwoei",
			false,
			"https://localhost",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// Arrange
			instance := viper.New()
			if tt.setToken {
				instance.Set("gitlab-access-token", tt.token)
			}
			if tt.setBaseURL {
				instance.Set("gitlab-base-url", tt.baseURL)
			}
			// Act
			err := tt.arg.ReconcileConfiguration(instance)
			// Assert
			assert.NoError(t, err)
			if tt.token != "" {
				assert.Equal(t, tt.token, tt.arg.Token)
			}
			if tt.baseURL != "" {
				assert.Equal(t, tt.baseURL, tt.arg.BaseURL)
			}
		})
	}
}
