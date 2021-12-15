package cmd

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	bbtestutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/test"
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

	factory := bbtestutil.GetFakeFactory(releaseFixture, nil, nil)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

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

	factory := bbtestutil.GetFakeFactory(releaseFixture, nil, nil)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewVersionCmd(factory, streams)
	cmd.SetArgs([]string{"-c"})
	cmd.Execute()

	if strings.Contains(buf.String(), "release version 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}

	if !strings.Contains(buf.String(), "bigbang cli version") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}
