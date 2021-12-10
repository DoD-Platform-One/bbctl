package cmd

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/test"
)

func TestGetList(t *testing.T) {

	chartFoo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "foo",
			Version: "1.0.2",
		},
	}

	chartBar := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bar",
			Version: "1.0.4",
		},
	}

	releaseFixture := []*release.Release{
		{
			Name:      "foo",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: chartFoo,
		},
		{
			Name:      "bar",
			Version:   2,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusFailed,
			},
			Chart: chartBar,
		},
	}

	factory := &bbutil.FakeFactory{HelmReleases: releaseFixture}

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewGetReleasesCmd(factory, streams)
	cmd.Run(cmd, []string{})

	response := strings.Split(buf.String(), "\n")

	releaseFoo := strings.ReplaceAll(response[1], "\t", "")
	releaseFoo = strings.ReplaceAll(releaseFoo, " ", "")

	if !strings.Contains(releaseFoo, "foobigbang1deployedfoo-1.0.2") {
		t.Errorf("unexpected output: %s", releaseFoo)
	}

	releaseBar := strings.ReplaceAll(response[2], "\t", "")
	releaseBar = strings.ReplaceAll(releaseBar, " ", "")

	if !strings.Contains(releaseBar, "barbigbang2failedbar-1.0.4") {
		t.Errorf("unexpected output: %s", releaseBar)
	}
}
