package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
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
		factory.ResetIOStream()

		streams := factory.GetIOStream()
		buf := streams.Out.(*bytes.Buffer)

		t.Run(test.desc, func(t *testing.T) {
			cmd := NewCompletionCmd(factory)
			err := cmd.RunE(cmd, []string{test.shell})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !strings.Contains(buf.String(), fmt.Sprintf("%v completion", test.shell)) {
				t.Errorf("unexpected output")
			}
		})
	}
}

func TestInvalidShellCompletion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	streams := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	cmd := NewCompletionCmd(factory)
	err := cmd.RunE(cmd, []string{"foo"})

	assert.Empty(t, buf.String())
	assert.Equal(t, err.Error(), "unknown shell: foo")
}
