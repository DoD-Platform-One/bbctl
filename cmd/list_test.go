package cmd

import (
	"strings"
	"testing"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
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

	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(releaseFixture)

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewReleasesCmd(factory, streams)
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
