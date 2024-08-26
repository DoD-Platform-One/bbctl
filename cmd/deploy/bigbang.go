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
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	outputSchema "repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"
)

var (
	bigBangUse = `bigbang`

	bigBangShort = i18n.T(`Deploy Big Bang components to your cluster`)

	bigBangLong = templates.LongDesc(
		i18n.T(`Deploy Big Bang and optional Big Bang addons to your cluster.
		This command invokes the helm command, so arguments after -- are passed to the underlying helm command.

		Note: deployment of Big Bang requires Flux to have been deployed to your cluster. See "bbctl deploy flux" for more information.
	`),
	)

	bigBangExample = templates.Examples(i18n.T(`
	    # Deploy Big Bang to your cluster
		bbctl deploy bigbang

		# Deploy Big Bang with additional configurations for a k3d development cluster
		bbctl deploy bigbang --k3d

		# Deploy Big Bang with a helm overrides file. All arguments after -- are passed to the underlying helm command
		bbctl deploy bigbang -- -f ../path/to/overrides/values.yaml
		`))
)

// NewDeployBigBangCmd - deploy Big Bang to your cluster
func NewDeployBigBangCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     bigBangUse,
		Short:   bigBangShort,
		Long:    bigBangLong,
		Example: bigBangExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployBigBangToCluster(cmd, factory, args)
		},
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("unable to get config client: %w", clientError)
	}

	k3dFlagError := configClient.SetAndBindFlag(
		"k3d",
		"",
		false,
		"Include some boilerplate suitable for deploying into k3d",
	)
	if k3dFlagError != nil {
		return nil, fmt.Errorf("Error setting k3d flag: %w", k3dFlagError)
	}

	addOnFlagError := configClient.SetAndBindFlag(
		"addon",
		"",
		[]string(nil),
		"Enable this Big Bang addon in the deployment",
	)
	if addOnFlagError != nil {
		return nil, fmt.Errorf("error setting addon flag: %w", addOnFlagError)
	}

	return cmd, nil
}

func getChartRelativePath(configClient *schemas.GlobalConfiguration, pathCmp ...string) string {
	repoPath := configClient.BigBangRepo
	return path.Join(slices.Insert(pathCmp, 0, repoPath)...)
}

func insertHelmOptForExampleConfig(
	config *schemas.GlobalConfiguration,
	helmOpts []string,
	chartName string,
) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(
			config,
			"docs",
			"assets",
			"configs",
			"example",
			chartName,
		),
	)
}

func insertHelmOptForRelativeChart(
	config *schemas.GlobalConfiguration,
	helmOpts []string,
	chartName string,
) []string {
	return slices.Insert(helmOpts,
		0,
		"-f",
		getChartRelativePath(
			config,
			"chart",
			chartName,
		),
	)
}

func deployBigBangToCluster(command *cobra.Command, factory bbUtil.Factory, args []string) error {
	loggingClient, err := factory.GetLoggingClient()
	if err != nil {
		return err
	}
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

	chartPath := getChartRelativePath(config, "chart")
	helmOpts := slices.Clone(args)
	loggingClient.Info(
		fmt.Sprintf(
			"preparing to deploy Big Bang to cluster, k3d=%v",
			config.DeployBigBangConfiguration.K3d,
		),
	)
	if config.DeployBigBangConfiguration.K3d {
		loggingClient.Info("Using k3d configuration")
		helmOpts = insertHelmOptForExampleConfig(config, helmOpts, "policy-overrides-k3d.yaml")
		helmOpts = insertHelmOptForRelativeChart(config, helmOpts, "ingress-certs.yaml")
	}
	for _, x := range config.DeployBigBangConfiguration.Addon {
		helmOpts = slices.Insert(helmOpts,
			0,
			"--set",
			fmt.Sprintf("addons.%s.enabled=true", x),
		)
	}
	helmOpts = slices.Insert(helmOpts,
		0,
		"upgrade",
		"-i",
		"bigbang",
		chartPath,
		"-n",
		"bigbang",
		"--create-namespace",
		"--set",
		fmt.Sprintf("registryCredentials.username=%v", username),
		"--set",
		fmt.Sprintf("registryCredentials.password=%v", password),
	)

	cmd, err := factory.GetCommandWrapper("helm", helmOpts...)
	if err != nil {
		return fmt.Errorf("unable to get command wrapper: %w", err)
	}

	// Use the factory to create the pipe
	err = factory.CreatePipe()
	if err != nil {
		return fmt.Errorf("unable to create pipe: %w", err)
	}

	r, w, err := factory.GetPipe()
	if err != nil {
		return fmt.Errorf("Unable to get pipe: %w", err)
	}

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
	data := &outputSchema.BigbangOutput{
		Data: encodeHelmOpts(buf.String()),
	}
	err = outputClient.Output(data)
	if err != nil {
		return err
	}

	return nil
}

func encodeHelmOpts(data string) outputSchema.HelmOutput {
	// Read the buffered output
	lines := strings.Split(data, "\n")

	// Initialize the outputSchemas.HelmOutput struct
	helmOutput := outputSchema.HelmOutput{}

	// Iterate over the lines to populate the struct
	notes := false
	for i, line := range lines {
		if i == 0 {
			helmOutput.Message = line // Store the first line as Message
			continue
		}
		// Once we read the line starting with `NOTES:`, all the remaining input should be considered part of a multi-line note string
		if notes {
			if strings.TrimSpace(line) != "" {
				helmOutput.Notes += "\n" + line
			}
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			switch key {
			case "NAME":
				helmOutput.Name = value
			case "LAST DEPLOYED":
				helmOutput.LastDeployed = value
			case "NAMESPACE":
				helmOutput.Namespace = value
			case "STATUS":
				helmOutput.Status = value
			case "REVISION":
				helmOutput.Revision = value
			case "TEST SUITE":
				helmOutput.TestSuite = value
			case "NOTES":
				helmOutput.Notes = value
				notes = true
			}
		}
	}

	return helmOutput
}
