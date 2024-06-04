package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestGetVersionUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	// Act
	cmd := NewVersionCmd(factory, streams)

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
	factory.SetHelmReleases(releaseFixture)
	factory.GetViper().Set("big-bang-repo", "test")

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	// Act
	cmd := NewVersionCmd(factory, streams)
	res := cmd.RunE(cmd, []string{})

	// Assert
	assert.NoError(t, res)
	if !assert.Contains(t, buf.String(), "release version 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
	if !assert.Contains(t, buf.String(), "bigbang cli version") {
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
	factory.SetHelmReleases(releaseFixture)
	factory.GetViper().Set("big-bang-repo", "test")
	factory.GetViper().Set("client", true)

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	// Act
	cmd := NewVersionCmd(factory, streams)
	err := cmd.RunE(cmd, []string{})

	// Assert
	assert.NoError(t, err)
	if !assert.NotContains(t, buf.String(), "release version 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
	if !assert.Contains(t, buf.String(), "bigbang cli version") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestGetVersionInvalidClientFlag(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	// Act
	cmd := NewVersionCmd(factory, streams)
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

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	// Act
	cmd := NewVersionCmd(factory, streams)
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

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	// Act
	cmd := NewVersionCmd(factory, streams)
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

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	// Act
	factory.SetFail.GetConfigClient = true

	// Assert
	assert.Panics(t, func() {
		NewVersionCmd(factory, streams)
	})
}

func TestGetVersionWithHelmError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	factory.GetViper().Set("big-bang-repo", "test")

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	// Act
	factory.SetFail.GetHelmClient = true
	cmd := NewVersionCmd(factory, streams)
	err := cmd.RunE(cmd, []string{})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "failed to get helm client")
}
