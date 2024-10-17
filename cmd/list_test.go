package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "/path/to/repo")
	v.Set("output-config.format", "text")

	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	cmd := NewReleasesCmd(factory)
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)

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
		return nil, errors.New(errorMessage)
	})
	err := listHelmReleases(cmd, factory, static.DefaultClient)

	// then
	require.Error(t, err)
	assert.Equal(t, "error getting helm releases in namespace bigbang: "+errorMessage, err.Error())
}

func TestListHelmReleases_NoHelmClient(t *testing.T) {
	// given

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	cmd := NewReleasesCmd(factory)

	// when
	factory.SetFail.GetHelmClient = true
	err := listHelmReleases(cmd, factory, static.DefaultClient)

	// then
	require.Error(t, err)
	assert.Equal(t, "failed to get helm client", err.Error())
}

func TestListHelmReleases_NoConstants(t *testing.T) {
	// given
	expectedError := errors.New("failed to get constants")

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	cmd := NewReleasesCmd(factory)

	// when
	constantsClient := mock.MockConstantsClient{}
	constantsClient.On("GetConstants").Return(static.Constants{BigBangNamespace: "bigbang"}, expectedError)
	err := listHelmReleases(cmd, factory, &constantsClient)

	// then
	require.Error(t, err)
}

func TestListHelmReleases_OutputClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetIOStreams = 1
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "")

	// Act
	cmd := NewReleasesCmd(factory)
	err := listHelmReleases(cmd, factory, static.DefaultClient)

	// Assert
	expectedError := "error getting output client: failed to get streams"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}

func TestListHelmReleases_MarshalError(t *testing.T) {
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
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "test")

	// Act
	cmd := NewReleasesCmd(factory)
	err := cmd.RunE(cmd, []string{})

	// Assert
	expectedError := "error marshaling Helm release output: unsupported format: test"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %v", expectedError, err)
	}
}
