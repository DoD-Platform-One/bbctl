package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	mock "repo1.dso.mil/big-bang/product/packages/bbctl/mocks/repo1.dso.mil/big-bang/product/packages/bbctl/static"
	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
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
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)

	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	cmd := NewReleasesCmd(factory)
	err := cmd.RunE(cmd, []string{})
	assert.Nil(t, err)

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

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	cmd := NewReleasesCmd(factory)

	// when
	factory.SetHelmGetListFunc(func() ([]*release.Release, error) {
		return nil, fmt.Errorf(errorMessage)
	})
	error := listHelmReleases(cmd, factory, static.DefaultClient)

	// then
	assert.NotNil(t, error)
	assert.Equal(t, "error getting helm releases in namespace bigbang: "+errorMessage, error.Error())
}

func TestListHelmReleases_NoHelmClient(t *testing.T) {
	// given

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	cmd := NewReleasesCmd(factory)

	// when
	factory.SetFail.GetHelmClient = true
	error := listHelmReleases(cmd, factory, static.DefaultClient)

	// then
	assert.NotNil(t, error)
	assert.Equal(t, "failed to get helm client", error.Error())
}

func TestListHelmReleases_NoConstants(t *testing.T) {
	// given
	expectedError := fmt.Errorf("failed to get constants")

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	cmd := NewReleasesCmd(factory)

	// when
	constantsClient := mock.MockConstantsClient{}
	constantsClient.On("GetConstants").Return(static.Constants{BigBangNamespace: "bigbang"}, expectedError)
	error := listHelmReleases(cmd, factory, &constantsClient)

	// then
	assert.NotNil(t, error)
}
