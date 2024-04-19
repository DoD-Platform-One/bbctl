package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestGetVersion(t *testing.T) {
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

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewVersionCmd(factory, streams)
	cmd.Run(cmd, []string{})

	if !strings.Contains(buf.String(), "release version 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestGetClientVersionOnly(t *testing.T) {
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

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewVersionCmd(factory, streams)
	cmd.SetArgs([]string{"-c"})
	err := cmd.Execute()
	assert.NoError(t, err)

	if strings.Contains(buf.String(), "release version 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}

	if !strings.Contains(buf.String(), "bigbang cli version") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}
