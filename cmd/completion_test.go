package cmd

import (
	"fmt"
	"strings"
	"testing"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestAllCompletions(t *testing.T) {
	var tests = []struct {
		desc  string
		shell string
	}{
		{
			desc:  "BashCompletion",
			shell: "bash",
		},
		{
			desc:  "ZshCompletion",
			shell: "zsh",
		},
		{
			desc:  "FishCompletion",
			shell: "fish",
		},
		{
			desc:  "PowershellCompletion",
			shell: "powershell",
		},
	}

	for _, test := range tests {
		factory := bbTestUtil.GetFakeFactory()
		streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

		t.Run(test.desc, func(t *testing.T) {
			cmd := NewCompletionCmd(factory, streams)
			cmd.Run(cmd, []string{test.shell})

			if !strings.Contains(buf.String(), fmt.Sprintf("%v completion", test.shell)) {
				t.Errorf("unexpected output")
			}
		})
	}
}

func TestFooCompletion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"foo"})

	if buf.String() != "" {
		t.Errorf("unexpected output")
	}
}
