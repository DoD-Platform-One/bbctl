package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		streams, _ := factory.GetIOStream()
		buf := streams.Out.(*bytes.Buffer)

		t.Run(test.desc, func(t *testing.T) {
			cmd, err := NewCompletionCmd(factory)
			require.NoError(t, err)
			err = cmd.RunE(cmd, []string{test.shell})
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

	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	cmd, err := NewCompletionCmd(factory)
	require.NoError(t, err)
	err = cmd.RunE(cmd, []string{"foo"})

	assert.Empty(t, buf.String())
	assert.Equal(t, "unknown shell: foo", err.Error())
}

func TestNewCompletionCmdErrorOnIOStreams(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetFail.GetIOStreams = 1
	// Act
	cmd, err := NewCompletionCmd(factory)
	// Assert
	assert.Nil(t, cmd)
	require.Error(t, err)
	assert.Equal(t, "unable to get IO streams: failed to get streams", err.Error())
}
