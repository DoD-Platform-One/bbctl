package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	bbutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util"
)

func TestGetValues(t *testing.T) {

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

	factory := bbutil.FakeFactory(releaseFixture)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewGetValuesCmd(factory, streams)
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
		command *cobra.Command
		input   string
		output  []string
	}

	factory := bbutil.FakeFactory(releaseFixture)

	streams, _, _, _ := genericclioptions.NewTestIOStreams()

	cmd := NewGetValuesCmd(factory, streams)

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
