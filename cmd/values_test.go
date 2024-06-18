package cmd

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	mock "repo1.dso.mil/big-bang/product/packages/bbctl/mocks/repo1.dso.mil/big-bang/product/packages/bbctl/static"
	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestValues(t *testing.T) {
	chartFoo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "foo",
			Version: "1.0.2",
		},
		Values: map[string]interface{}{
			"domain":   "test",
			"hostname": "test.com",
			"foo": map[string]interface{}{
				"enabled": 2,
				"count":   1,
			},
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
	}

	factory := bbUtil.GetFakeFactory()
	factory.SetHelmReleases(releaseFixture)

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewValuesCmd(factory, streams)
	cmd.RunE(cmd, []string{"foo"})

	if !strings.Contains(buf.String(), "enabled: 2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestGetValuesCompletion(t *testing.T) {
	chartFoo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "foo",
			Version: "1.0.2",
		},
		Values: map[string]interface{}{
			"domain":   "test",
			"hostname": "test.com",
			"foo": map[string]interface{}{
				"enabled": 2,
				"count":   1,
			},
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

	type test struct {
		input  string
		output []string
	}

	factory := bbUtil.GetFakeFactory()
	factory.SetHelmReleases(releaseFixture)

	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewValuesCmd(factory, streams)

	tests := []test{
		{input: "", output: []string{"foo", "bar"}},
		{input: "f", output: []string{"foo"}},
		{input: "ba", output: []string{"bar"}},
		{input: "z", output: []string{}},
	}

	for _, tc := range tests {
		suggestions, _ := cmd.ValidArgsFunction(cmd, []string{}, tc.input)
		if !reflect.DeepEqual(tc.output, suggestions) {
			t.Fatalf("expected: %v, got: %v", tc.output, suggestions)
		}
	}
}

func TestGetValuesCompletionTooManyArgs(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewValuesCmd(factory, streams)
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{"too", "many", "args"}, "test")
	assert.Empty(t, suggestions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

// TestNewValuesCmdHelperSucces tests that the default values helper
// constructor returns the helper successfully
func TestNewValuesCmdHelperSucces(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewValuesCmd(factory, streams)

	v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)

	// Assert that we don't return an error
	assert.NoError(t, err)

	// Assert that the client returned is not nil (nil is returned in all
	// error cases)
	assert.NotNil(t, v)
}

// TestNewValuesCmdHelperFailHelmClient tests that if we fail to get the default helper client,
// we correctly return an error and stop execution
func TestNewValuesCmdHelperFailHelmClient(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	factory.SetFail.GetHelmClient = true
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewValuesCmd(factory, streams)

	v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)

	// Assert that we failed to get a helm client
	assert.Equal(t, "failed to get helm client", err.Error())

	// Assert that the helper client is nil
	assert.Nil(t, v)
}

// TestNewValuesCmdHelperFailConstantsClient tests that if we fail to get the default helper client,
// we correctly return an error and stop execution
func TestNewValuesCmdHelperFailConstantsClient(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewValuesCmd(factory, streams)

	expectedError := fmt.Errorf("failed to get constants")

	constantsClient := mock.MockConstantsClient{}
	constantsClient.On("GetConstants").Return(static.Constants{BigBangNamespace: "bigbang"}, expectedError)

	v, err := newValuesCmdHelper(cmd, factory, &constantsClient)

	// Assert that we failed to get the constants
	assert.Equal(t, expectedError.Error(), err.Error())

	// Assert that the helper client is nil
	assert.Nil(t, v)

	// Check that expected methods were called
	constantsClient.AssertExpectations(t)
}

// TestGetHelmValuesFailGettingValues tests that when failing to get helm values
// we capture the error and exit appropriately
func TestGetHelmValuesFailGettingValues(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewValuesCmd(factory, streams)

	expectedError := fmt.Errorf("error getting helm release values in namespace bigbang: release test not found")

	v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)
	assert.NoError(t, err)
	assert.NotNil(t, v)

	// Will fail to get helm values as we've set releases to be a nil value
	err = v.getHelmValues(streams, "test")

	assert.Equal(t, expectedError.Error(), err.Error())
}

// TestMatchingReleaseNamesFailGettingList tests that if the helm client fails to get the list of
// release names, we return the default shell completion directive
func TestMatchingReleaseNamesFailGettingList(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	factory.SetHelmGetListFunc(func() ([]*release.Release, error) {
		return nil, fmt.Errorf("error getting list")
	})
	streams, _, _, _ := genericIOOptions.NewTestIOStreams()
	cmd := NewValuesCmd(factory, streams)

	v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)
	assert.NoError(t, err)
	assert.NotNil(t, v)

	matches, directive := v.matchingReleaseNames("test")
	assert.Empty(t, matches)
	assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)

}
