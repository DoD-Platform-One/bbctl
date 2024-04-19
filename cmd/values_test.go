package cmd

import (
	"reflect"
	"strings"
	"testing"

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
	cmd.Run(cmd, []string{"foo"})

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
