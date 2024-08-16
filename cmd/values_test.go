package cmd

import (
	"bytes"
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
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")

	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)

	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	cmd := NewValuesCmd(factory)
	err := cmd.RunE(cmd, []string{"foo"})
	assert.NoError(t, err)

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
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")

	cmd := NewValuesCmd(factory)

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
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")

	cmd := NewValuesCmd(factory)
	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{"too", "many", "args"}, "test")
	assert.Empty(t, suggestions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

// TestNewValuesCmdHelperSucces tests that the default values helper
// constructor returns the helper successfully
func TestNewValuesCmdHelperSucces(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")
	cmd := NewValuesCmd(factory)

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
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")
	cmd := NewValuesCmd(factory)

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
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")
	cmd := NewValuesCmd(factory)

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
	factory.ResetIOStream()
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")
	cmd := NewValuesCmd(factory)

	expectedError := fmt.Errorf("error getting helm release values in namespace bigbang: release test not found")

	v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)
	assert.NoError(t, err)
	assert.NotNil(t, v)

	// Will fail to get helm values as we've set releases to be a nil value
	streams, _ := factory.GetIOStream()
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
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")
	cmd := NewValuesCmd(factory)

	v, err := newValuesCmdHelper(cmd, factory, static.DefaultClient)
	assert.NoError(t, err)
	assert.NotNil(t, v)

	matches, directive := v.matchingReleaseNames("test")
	assert.Empty(t, matches)
	assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
}

func TestValuesOutputClientError(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")
	factory.SetFail.GetConfigClient = true
	cmd := NewValuesCmd(factory)
	assert.NotNil(t, cmd)

	suggestions, directive := cmd.ValidArgsFunction(cmd, []string{}, "")
	err := cmd.RunE(cmd, []string{})
	assert.Nil(t, suggestions)
	assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting output client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestValuesGetIOStreamsError(t *testing.T) {
	factory := bbUtil.GetFakeFactory()
	viper, _ := factory.GetViper()
	// Required value or the execution will fail
	viper.Set("big-bang-repo", "/path/to/repo")
	viper.Set("output-config.format", "yaml")
	factory.SetFail.GetIOStreams = true
	cmd := NewValuesCmd(factory)
	assert.NotNil(t, cmd)

	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "failed to get streams") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
