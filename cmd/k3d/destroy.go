package k3d

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
)

var (
	destroyUse = `destroy`

	destroyShort = i18n.T(`Destroys a k3d cluster`)

	destroyLong = templates.LongDesc(i18n.T(`Destroys a previously created AWS k3d cluster.
	This is a wrapper around the k3d-dev.sh script. It must be checked out at --big-bang-repo location.
	Any command line arguments following -- are passed to k3d-dev.sh (including --help).`))

	destroyExample = templates.Examples(i18n.T(`
	    # Destroy your k3d cluster previously built with 'bbctl k3d create'
		bbctl k3d destroy
		
		# To get the full help message from k3d-dev.sh
		bbctl k3d destroy -- --help`))
)

// NewDestroyClusterCmd - Returns a command to destroy a k3d cluster using destroyCluster
func NewDestroyClusterCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     destroyUse,
		Short:   destroyShort,
		Long:    destroyLong,
		Example: destroyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return destroyCluster(factory, cmd, args)
		},
	}

	return cmd
}

// destroyCluster - Passes through the global configurations, the path to the script, and command line arguments to the k3d-dev script to destroy the k3d dev cluster

func destroyCluster(factory bbUtil.Factory, cobraCmd *cobra.Command, args []string) (err error) {
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

	// Define the command path
	command := path.Join(config.BigBangRepo, "docs", "assets", "scripts", "developer", "k3d-dev.sh")

	// Prepend the `-d` flag to the arguments for destroying the cluster
	args = append([]string{"-d"}, args...)

	// Wrap the command
	cmd, err := factory.GetCommandWrapper(command, args...)
	if err != nil {
		return fmt.Errorf("Unable to get command wrapper: %w", err)
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
		return fmt.Errorf("error waiting for goroutine: %w", err)
	}

	// Process the captured output
	data := &outputSchema.K3dOutput{
		Data: parseOutput(buf.String()),
	}
	err = outputClient.Output(data)
	return err
}
