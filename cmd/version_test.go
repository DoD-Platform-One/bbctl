package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

func TestGetVersionUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.Contains(t, cmd.Example, "bbctl version --client")
}

func TestGetVersion(t *testing.T) {
	// Arrange
	chartInfo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bigbang",
			Version: "1.0.2",
		},
	}

	releaseFixture := []*release.Release{
		{
			Name:      "bigbang",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: chartInfo,
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)
	factory.GetViper().Set("big-bang-repo", "test")

	streams := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	// Act
	cmd, _ := NewVersionCmd(factory)
	res := cmd.RunE(cmd, []string{})

	// Assert
	assert.NoError(t, res)
	if !assert.Contains(t, buf.String(), "release version 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
	if !assert.Contains(t, buf.String(), "bbctl client version") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestGetVersionClientVersionOnly(t *testing.T) {
	// Arrange
	chartInfo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bigbang",
			Version: "1.0.2",
		},
	}

	releaseFixture := []*release.Release{
		{
			Name:      "bigbang",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: chartInfo,
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)
	factory.GetViper().Set("big-bang-repo", "test")
	factory.GetViper().Set("client", true)

	streams := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	// Act
	cmd, _ := NewVersionCmd(factory)
	err := cmd.RunE(cmd, []string{})

	// Assert
	assert.NoError(t, err)
	if !assert.NotContains(t, buf.String(), "release version 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
	if !assert.Contains(t, buf.String(), "bbctl client version") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestGetVersionInvalidClientFlag(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)
	cmd.SetArgs([]string{"--client=string-value"})
	err := cmd.Execute()

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "invalid argument \"string-value\" for \"--client\" flag") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestGetVersionWithError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)
	err := cmd.RunE(cmd, []string{})

	// Assert
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), "error getting helm information for release bigbang in namespace bigbang: release bigbang not found")
	}
}

func TestGetVersionWithBadParams(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)
	cmd.SetArgs([]string{"--invalid-parameter"})
	res := cmd.Execute()

	// Assert
	assert.Error(t, res)
	assert.Equal(t, res.Error(), "unknown flag: --invalid-parameter")
}

func TestGetVersionWithConfigError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	// Act
	factory.SetFail.GetConfigClient = true
	cmd, err := NewVersionCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "Unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestGetVersionWithHelmError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	// Act
	factory.SetFail.GetHelmClient = true
	cmd, _ := NewVersionCmd(factory)
	err := cmd.RunE(cmd, []string{})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "failed to get helm client")
}

func TestVersionFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient := factory.GetLoggingClient()
	cmd, _ := NewVersionCmd(factory)
	viper := factory.GetViper()
	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, fmt.Errorf("Dummy Error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	// Act
	err1 := bbVersion(cmd, factory)

	// Assert
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
}
