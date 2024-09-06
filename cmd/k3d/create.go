package k3d

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
)

var (
	createUse   = `create`
	createShort = i18n.T(`Creates a k3d cluster`)
	createLong  = templates.LongDesc(
		i18n.T(`Creates a minimal k3d cluster in AWS for development or testing.
	This is a wrapper around the k3d-dev.sh script. It must be checked out at --big-bang-repo location.
	Any command line arguments following -- are passed to k3d-dev.sh (including --help).`),
	)
	createExample = templates.Examples(i18n.T(`
	    # Create a default k3d cluster in AWS
		bbctl k3d create

		# Get the full help message from k3d-dev.sh
		bbctl k3d create -- --help
		
		# Create a k3d cluster in AWS on a BIG M5 with a private IP and metalLB installed
		bbctl k3d create -- -b -p -m`))
)

// NewCreateClusterCmd - Returns a command to create the k3d cluster using createCluster
func NewCreateClusterCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     createUse,
		Short:   createShort,
		Long:    createLong,
		Example: createExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createCluster(factory, cmd, args)
		},
	}

	return cmd
}

// createCluster - Passes through the global configurations, the path to the script, and command line arguments to the k3d-dev script to create the k3d dev cluster
func createCluster(factory bbUtil.Factory, cobraCmd *cobra.Command, args []string) (err error) {
	streams, err := factory.GetIOStream()
	if err != nil {
		return err
	}
	configClient, err := factory.GetConfigClient(cobraCmd)
	if err != nil {
		return err
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}
	outputClient, err := factory.GetOutputClient(cobraCmd)
	if err != nil {
		return fmt.Errorf("Unable to create output client: %w", err)
	}

	command := path.Join(config.BigBangRepo,
		"docs",
		"assets",
		"scripts",
		"developer",
		"k3d-dev.sh",
	)
	cmd, err := factory.GetCommandWrapper(command, args...)
	if err != nil {
		return err
	}

	// Use the factory to get the pipe
	r, w, err := factory.GetPipe()
	if err != nil {
		return fmt.Errorf("unable to get pipe: %w", err)
	}
	// Redirect command's output
	cmd.SetStdout(w)
	cmd.SetStderr(streams.ErrOut) // Set stderr to original
	cmd.SetStdin(streams.In)

	// Use a buffer to capture the output
	var buf bytes.Buffer
	var wg sync.WaitGroup

	// Add one to the WaitGroup counter
	wg.Add(1)

	go func() {
		defer wg.Done() // Decrement the WaitGroup counter when the goroutine completes
		if _, newErr := io.Copy(&buf, r); newErr != nil {
			if err == nil {
				err = fmt.Errorf("(sole deferred error: %w)", newErr)
			} else {
				err = fmt.Errorf("%w (additional deferred error: %v)", err, newErr)
			}
		}
	}()

	// Run the command
	err = cmd.Run()
	if err != nil {
		w.Close()
		wg.Wait() // Wait for the goroutine to finish before returning
		return err
	}

	// Close the writer to finish the reading process
	w.Close()

	// Wait for the goroutine to finish
	wg.Wait()
	if err != nil {
		err = fmt.Errorf("error waiting for goroutine: %w", err)
		return err
	}

	// Process the captured output
	data := &outputSchema.K3dOutput{
		Data: parseOutput(buf.String()),
	}
	err = outputClient.Output(data)
	return err
}

func parseOutput(data string) outputSchema.Output {
	lines := strings.Split(data, "\n")
	parsedOutput := outputSchema.Output{
		Actions:  []string{},
		Warnings: []string{},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "Warning:") {
			// If the line starts with "Warning:", treat it as a warning
			parsedOutput.Warnings = append(parsedOutput.Warnings, line)
		} else {
			// Otherwise, treat it as an action
			parsedOutput.Actions = append(parsedOutput.Actions, line)
		}
	}

	return parsedOutput
}
