package cmd

import (
	"strings"
	"testing"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	bbtestutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/test"
)

func TestBashCompletion(t *testing.T) {

	factory := bbtestutil.GetFakeFactory(nil, nil, nil, nil)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"bash"})

	if !strings.Contains(buf.String(), "bash completion") {
		t.Errorf("unexpected output")
	}
}

func TestZshCompletion(t *testing.T) {

	factory := bbtestutil.GetFakeFactory(nil, nil, nil, nil)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"zsh"})

	if !strings.Contains(buf.String(), "zsh completion") {
		t.Errorf("unexpected output")
	}
}

func TestFishCompletion(t *testing.T) {

	factory := bbtestutil.GetFakeFactory(nil, nil, nil, nil)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"fish"})

	if !strings.Contains(buf.String(), "fish completion") {
		t.Errorf("unexpected output")
	}
}

func TestFooCompletion(t *testing.T) {

	factory := bbtestutil.GetFakeFactory(nil, nil, nil, nil)

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewCompletionCmd(factory, streams)
	cmd.Run(cmd, []string{"foo"})

	if buf.String() != "" {
		t.Errorf("unexpected output")
	}
}
