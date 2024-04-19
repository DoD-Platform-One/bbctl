package cmd

import (
	"strings"
	"testing"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestBashCompletion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"bash"})

	if !strings.Contains(buf.String(), "bash completion") {
		t.Errorf("unexpected output")
	}
}

func TestZshCompletion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"zsh"})

	if !strings.Contains(buf.String(), "zsh completion") {
		t.Errorf("unexpected output")
	}
}

func TestFishCompletion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()

	streams, _, buf, _ := genericIOOptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"fish"})

	if !strings.Contains(buf.String(), "fish completion") {
		t.Errorf("unexpected output")
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
