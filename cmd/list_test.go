package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	mock "repo1.dso.mil/big-bang/product/packages/bbctl/mocks/repo1.dso.mil/big-bang/product/packages/bbctl/static"
	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestListHelmReleases_HappyPath(t *testing.T) {
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

func TestListHelmReleases_NoList(t *testing.T) {
	// given
	errorMessage := "error retrieving list"
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	factory := bbTestUtil.GetFakeFactory()

	cmd := NewReleasesCmd(factory, streams)

	// when
	factory.SetHelmGetListFunc(func() ([]*release.Release, error) {
		return nil, fmt.Errorf(errorMessage)
	})
	error := listHelmReleases(cmd, factory, streams, static.DefaultClient)

	// then
	assert.NotNil(t, error)
	assert.Equal(t, "error getting helm releases in namespace bigbang: "+errorMessage, error.Error())
}

func TestListHelmReleases_NoHelmClient(t *testing.T) {
	// given
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	factory := bbTestUtil.GetFakeFactory()

	cmd := NewReleasesCmd(factory, streams)

	// when
	factory.SetFail.GetHelmClient = true
	error := listHelmReleases(cmd, factory, streams, static.DefaultClient)

	// then
	assert.NotNil(t, error)
	assert.Equal(t, "failed to get helm client", error.Error())
}

func TestListHelmReleases_NoConstants(t *testing.T) {
	// given
	expectedError := fmt.Errorf("failed to get constants")
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	factory := bbTestUtil.GetFakeFactory()

	cmd := NewReleasesCmd(factory, streams)

	// when
	constantsClient := mock.MockConstantsClient{}
	constantsClient.On("GetConstants").Return(static.Constants{BigBangNamespace: "bigbang"}, expectedError)
	error := listHelmReleases(cmd, factory, streams, &constantsClient)

	// then
	assert.NotNil(t, error)
}
