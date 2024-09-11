package cmd

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	helmRelease "helm.sh/helm/v3/pkg/release"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sClient "k8s.io/client-go/dynamic"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	versionUse = `version`

	versionShort = i18n.T(`Print the current bbctl client version and the version of the Big Bang currently deployed.`)

	versionLong = templates.LongDesc(i18n.T(`Print the version of the bbctl client and the version of Big Bang currently deployed.
	 The Big Bang deployment version is pulled from the cluster currently referenced by your KUBECONFIG setting if no cluster parameters are provided.
	 Using the --client flag will only return the bbctl client version.`))

	versionExample = templates.Examples(i18n.T(`
		# Print version
		bbctl version
		
		# Print the bbctl client version only
		bbctl version --client

		# Get the version of a specific chart
		bbctl version CHART_NAME

		# Get the version of all current installed chartes managed by Big Bang
		bbctl version --all-charts

		# Get the latest version of a given chart
		bbctl version CHART_NAME --check-for-updates

		# Get the latest version of all current installed chartes managed by Big Bang
		bbctl version --all-charts --check-for-updates
		`))
)

type versionCmdHelper struct {
	constants    static.Constants
	config       *schemas.GlobalConfiguration
	helmClient   helm.Client
	outputClient output.Client
	logger       bbLog.Client
	gitlabClient gitlab.Client
	kubeClient   k8sClient.Interface
}

func newVersionCmdHelper(cmd *cobra.Command, factory bbUtil.Factory, constantsClient static.ConstantsClient) (*versionCmdHelper, error) {
	constants, err := constantsClient.GetConstants()
	if err != nil {
		return nil, err
	}

	configClient, err := factory.GetConfigClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("unable to get config client: %w", err)
	}
	config, err := configClient.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting config: %w", err)
	}

	helmClient, err := factory.GetHelmClient(cmd, constants.BigBangNamespace)
	if err != nil {
		return nil, err
	}

	outputClient, err := factory.GetOutputClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting output client: %w", err)
	}

	logger, err := factory.GetLoggingClient()
	if err != nil {
		return nil, fmt.Errorf("error getting logging client: %w", err)
	}

	gitlabClient, err := factory.GetGitLabClient()
	if err != nil {
		return nil, fmt.Errorf("error getting gitlab client: %w", err)
	}

	kubeClient, err := factory.GetK8sDynamicClient(cmd)
	if err != nil {
		return nil, fmt.Errorf("error getting k8s client: %w", err)
	}

	return &versionCmdHelper{
		constants:    constants,
		config:       config,
		helmClient:   helmClient,
		outputClient: outputClient,
		logger:       logger,
		gitlabClient: gitlabClient,
		kubeClient:   kubeClient,
	}, nil
}

// NewVersionCmd - Creates a new Cobra command which implements the `bbctl version` functionality
func NewVersionCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     versionUse,
		Short:   versionShort,
		Long:    versionLong,
		Example: versionExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
			if err != nil {
				return fmt.Errorf("error creating version helper: %w", err)
			}

			return v.bbVersion(args)

		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("unable to get config client: %w", clientError)
	}

	flagError := configClient.SetAndBindFlag(
		"client",
		"",
		false,
		"Print the bbctl client version only",
	)
	if flagError != nil {
		return nil, fmt.Errorf("error setting and binding client flag: %w", flagError)
	}

	err := configClient.SetAndBindFlag(
		"all-charts",
		"A",
		false,
		"Print the version of all Big Bang charts",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding all-charts flag: %w", err)
	}

	err = configClient.SetAndBindFlag(
		"check-for-updates",
		"U",
		false,
		"Check the upstream repo for a newer version of a chart",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding check-for-updates flag: %w", err)
	}

	return cmd, nil
}

// bbVersion is a helper function to separate a lot of the args and config logic
// from the command function creation for easier unit testing
func (v *versionCmdHelper) bbVersion(args []string) error {
	var err error

	outputMap := map[string]any{}
	// Short circuit if the user only wants the bbctl client version
	if v.config.VersionConfiguration.Client {
		return v.outputClient.Output(&output.BasicOutput{
			Vals: map[string]any{
				"bbctl": v.constants.BigBangCliVersion,
			},
		})
	}

	if v.config.VersionConfiguration.AllCharts {
		outputMap, err = v.getAllChartVersions(v.config.VersionConfiguration.CheckForUpdates)
		if err != nil {
			return fmt.Errorf("error getting all chart versions: %w", err)
		}
		return v.outputClient.Output(&output.BasicOutput{
			Vals: outputMap,
		})
	}

	switch len(args) {
	// If no arguments are provided, print the version of the Big Bang release
	// and the bbctl client
	case 0:
		if v.config.VersionConfiguration.CheckForUpdates {
			outputMap, err = v.checkForUpdates("bigbang")
			if err != nil {
				return fmt.Errorf("error checking for updates: %w", err)
			}
		} else {

			outputMap, err = v.getBigBangVersion()
			if err != nil {
				return fmt.Errorf("error getting Big Bang version: %w", err)
			}
		}
	// If an argument is provided, print the version of the specific release named
	case 1:
		chartName := args[0]
		chartVersion, err := v.getChartVersion(chartName)
		if err != nil {
			return fmt.Errorf("error getting chart version: %w", err)
		}

		if v.config.VersionConfiguration.CheckForUpdates {
			outputMap, err = v.checkForUpdates(chartName)
			if err != nil {
				return fmt.Errorf("error checking for updates: %w", err)
			}
		} else {
			outputMap = map[string]any{
				chartName: chartVersion,
			}
		}
	default:
		return fmt.Errorf("invalid number of arguments")
	}

	return v.outputClient.Output(&output.BasicOutput{
		Vals: outputMap,
	})
}

// getAllChartVersions gets the version of all the charts in the cluster owned by Big Bang
func (v *versionCmdHelper) getAllChartVersions(checkForUpdates bool) (map[string]any, error) {
	outputMap := map[string]any{}

	v.logger.Debug("getting all charts' versions")
	releases, err := v.helmClient.GetList()
	if err != nil {
		return outputMap, fmt.Errorf("error getting helm information for all charts: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(releases))

	for _, release := range releases {
		key := release.Chart.Metadata.Name
		wg.Add(1)
		go func(release *helmRelease.Release, key string) {
			defer wg.Done()
			if checkForUpdates {
				latestVersion, err := v.getLatestChartVersion(key)
				if err != nil {
					errChan <- fmt.Errorf("error getting latest chart version: %w", err)
					return
				}
				update, err := updateIsNewer(release.Chart.Metadata.Version, latestVersion)
				if err != nil {
					errChan <- fmt.Errorf("error checking for update: %w", err)
					return
				}
				mu.Lock()
				outputMap[key] = map[string]any{
					"version":         release.Chart.Metadata.Version,
					"latest":          latestVersion,
					"updateAvailable": update,
				}
				mu.Unlock()
			} else {
				mu.Lock()
				outputMap[key] = release.Chart.Metadata.Version
				mu.Unlock()
			}
		}(release, key)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return outputMap, err
		}
	}

	return outputMap, nil
}

// getBigBangVersion gets the version of the Big Bang release as well as the version of the bbctl client
func (v *versionCmdHelper) getBigBangVersion() (map[string]any, error) {
	outputMap := map[string]any{}

	bigbangVersion, err := v.getReleaseVersion(v.constants.BigBangHelmReleaseName)
	if err != nil {
		return outputMap, fmt.Errorf("error fetching Big Bang release version: %w", err)
	}

	bbctlKey := "bbctl"
	outputMap = map[string]any{
		bbctlKey:                           v.constants.BigBangCliVersion,
		v.constants.BigBangHelmReleaseName: bigbangVersion,
	}

	return outputMap, nil
}

// getReleaseVersion gets the version of a release by the release name
func (v *versionCmdHelper) getReleaseVersion(releaseName string) (string, error) {
	release, err := v.helmClient.GetRelease(releaseName)
	if err != nil {
		return "", fmt.Errorf("error getting helm information for release %s: %w",
			releaseName, err)
	}
	version := release.Chart.Metadata.Version

	if version == "" {
		return "", fmt.Errorf(`error getting version for release "%s": no version specified for the chart`, releaseName)
	}

	return version, nil
}

// getChartVersion gets the version of a chart by the chart name
func (v *versionCmdHelper) getChartVersion(chartName string) (string, error) {
	// We want to find the release name from the chart name, since the release name
	// is less obvious to the end user. To to this, we'll fetch all the releases and iterate over them
	// looking at the associated chart name until we find the one we're looking for.
	releases, err := v.helmClient.GetList()
	if err != nil {
		return "", fmt.Errorf("error getting helm information for all releases: %w", err)
	}

	// Filter releases by chart name
	for _, release := range releases {
		// Chart names have a version in them (e.g. chart-1.2.3-bb.0), so we'll split on the first instance of a number
		releaseChart := splitChartName(release.Chart.Metadata.Name)
		if releaseChart == chartName {
			return release.Chart.Metadata.Version, nil
		}
	}

	return "", fmt.Errorf("no matching releases found for %s", chartName)
}

// checkForChartUpdate checks the current chart version against the latest version available on repo1
func (v *versionCmdHelper) getLatestChartVersion(chartName string) (string, error) {
	v.logger.Debug(fmt.Sprintf("checking for update to %s", chartName))

	var packageUri, branch string

	// Special consideraiotns for the bigbang chart
	if chartName == "bigbang" {
		packageUri = "big-bang/bigbang"
		branch = "master"
	} else {
		// For all other charts, we'll use the main branch
		branch = "main"

		// We need to find the GitRepo CRD that matches the chart name to determine the upstream URl for the chart
		// because not all charts' names are the same as the repo (e.g. "gatekeeper" is "policy")
		customResource := schema.GroupVersionResource{Group: "source.toolkit.fluxcd.io", Version: "v1", Resource: "gitrepositories"}
		opts := metaV1.GetOptions{}

		resource, err := v.kubeClient.Resource(customResource).Namespace(v.constants.BigBangNamespace).Get(context.TODO(), chartName, opts)
		if err != nil {
			return "", fmt.Errorf("error getting gitrepositories: %w", err)
		}

		// Parse out the package URL from the GitRepo CRD
		packageUri = strings.TrimSuffix(resource.Object["spec"].(map[string]any)["url"].(string), ".git")
		packageUri = strings.TrimPrefix(packageUri, "https://repo1.dso.mil/")
	}

	// Fetch the latest Chart.yalm from the upstream repo
	body, err := v.gitlabClient.GetFile(packageUri, "chart/Chart.yaml", branch)
	if err != nil {
		return "", fmt.Errorf("error getting Chart.yaml: %w", err)
	}

	helmChart, err := decodeChartYAML(body)
	if err != nil {
		return "", fmt.Errorf("failed to decode Chart.yaml: %w", err)
	}

	return helmChart.Version, nil
}

func (v *versionCmdHelper) checkForUpdates(chartName string) (map[string]any, error) {
	outputMap := map[string]any{}

	latestVersion, err := v.getLatestChartVersion(chartName)
	if err != nil {
		return outputMap, fmt.Errorf("error checking for latest chart version: %w", err)
	}

	currentVersion, err := v.getChartVersion(chartName)
	if err != nil {
		return outputMap, fmt.Errorf("error getting current chart version: %w", err)
	}

	update, err := updateIsNewer(currentVersion, latestVersion)
	if err != nil {
		return outputMap, fmt.Errorf("error checking for update: %w", err)
	}

	outputMap = map[string]any{
		"updateAvailable": update,
		"version":         currentVersion,
		"latest":          latestVersion,
	}

	return outputMap, nil
}

// splitChartName splits a chart name at the first instance of a number
// e.g. chart-1.2.3-bb.0 -> chart
func splitChartName(fullName string) string {
	re := regexp.MustCompile(`^(.*?)-(\d.*)$`)
	matches := re.FindStringSubmatch(fullName)

	if len(matches) == 3 {
		return matches[1]
	}

	// If no match found, return the full name as the chart name
	return fullName
}

// helmChartManifest is a struct that that partially represents the Chart.yaml file
type helmChartManifest struct {
	Version string `yaml:"version"`
}

// decodeChartYAML decodes the Chart.yaml file body from GitLab into a type
func decodeChartYAML(fileBody []byte) (*helmChartManifest, error) {
	var chart helmChartManifest
	err := yaml.Unmarshal(fileBody, &chart)
	if err != nil {
		return nil, err
	}
	return &chart, nil
}

// updateIsNewer parses the current and latest versions and
// returns true if the current version is greater than the latest version
func updateIsNewer(current, latest string) (bool, error) {
	currentVersion, err := semver.NewVersion(current)
	if err != nil {
		return false, fmt.Errorf(`invalid version "%s": %w`, current, err)
	}

	latestVersion, err := semver.NewVersion(latest)
	if err != nil {
		return false, fmt.Errorf(`invalid version "%s": %w`, latest, err)
	}

	return latestVersion.GreaterThan(currentVersion), nil
}
