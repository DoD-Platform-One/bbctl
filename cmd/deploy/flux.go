package deploy

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
)

var (
	fluxUse   = `flux`
	fluxShort = i18n.T(`Deploy flux to your kubernetes cluster`)
	fluxLong  = templates.LongDesc(
		i18n.T(
			`Deploy flux to your kubernetes cluster in a way specifically designed to support the deployment of Big Bang`,
		),
	)
	fluxExample = templates.Examples(i18n.T(`# Deploy flux to your cluster
		bbctl deploy flux`))
)

// NewDeployFluxCmd - parent for deploy commands
func NewDeployFluxCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fluxUse,
		Short:   fluxShort,
		Long:    fluxLong,
		Example: fluxExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployFluxToCluster(factory, cmd, args)
		},
	}

	return cmd
}

func deployFluxToCluster(factory bbUtil.Factory, command *cobra.Command, args []string) error {
	configClient, err := factory.GetConfigClient(command)
	if err != nil {
		return err
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}
	streams, err := factory.GetIOStream()
	if err != nil {
		return fmt.Errorf("unable to create IO streams: %w", err)
	}
	outputClient, err := factory.GetOutputClient(command)
	if err != nil {
		return fmt.Errorf("unable to create output client: %w", err)
	}
	credentialHelper, err := factory.GetCredentialHelper()
	if err != nil {
		return fmt.Errorf("unable to get credential helper: %w", err)
	}
	username, err := credentialHelper("username", "registry1.dso.mil")
	if err != nil {
		return fmt.Errorf("unable to get username: %w", err)
	}
	password, err := credentialHelper("password", "registry1.dso.mil")
	if err != nil {
		return fmt.Errorf("unable to get password: %w", err)
	}

	installFluxPath := path.Join(config.BigBangRepo, "scripts", "install_flux.sh")
	fluxArgs := append(slices.Clone(args), "-u", username, "-p", password)

	cmd, err := factory.GetCommandWrapper(installFluxPath, fluxArgs...)
	if err != nil {
		return fmt.Errorf("unable to get command wrapper: %w", err)
	}

	// Use the factory to create the pipe
	err = factory.CreatePipe()
	if err != nil {
		return fmt.Errorf("unable to create pipe: %w", err)
	}

	r, w := factory.GetPipe()

	streams.In = r
	streams.Out = w

	// Redirect command's stdout to the pipe's writer
	cmd.SetStdout(streams.Out)
	cmd.SetStderr(streams.ErrOut) // Set stderr to original

	// Use a buffer to capture the output
	var buf bytes.Buffer
	var wg sync.WaitGroup

	// Add one to the WaitGroup counter
	wg.Add(1)

	go func() {
		defer wg.Done() // Decrement the WaitGroup counter when the goroutine completes
		io.Copy(&buf, streams.In)
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

	// Process captured output
	data := &outputSchema.FluxOutput{
		Data: parseOutput(buf.String()),
	}
	err = outputClient.Output(data)
	if err != nil {
		return err
	}

	return nil
}

func parseOutput(data string) outputSchema.Output {
	lines := strings.Split(data, "\n")
	parsedOutput := outputSchema.Output{
		GeneralInfo: make(map[string]string),
		Actions:     []string{},
		Warnings:    []string{},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if !strings.Contains(line, ": ") {
			parsedOutput.Actions = append(parsedOutput.Actions, line)
			continue
		}

		parts := strings.Split(line, ": ")
		if len(parts) != 2 {
			parsedOutput.Actions = append(parsedOutput.Actions, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "REGISTRY_URL", "REGISTRY_USERNAME":
			parsedOutput.GeneralInfo[key] = value
		case "Warning":
			parsedOutput.Warnings = append(parsedOutput.Warnings, value)
		default:
			parsedOutput.Actions = append(parsedOutput.Actions, line)
		}
	}

	return parsedOutput
}
