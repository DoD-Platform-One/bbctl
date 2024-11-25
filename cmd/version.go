package cmd

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"repo1.dso.mil/big-bang/apps/developer-tools/go-utils/yamler"

	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	config "repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/ironbank"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/output/schemas"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sClient "k8s.io/client-go/dynamic"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

const BigBangChartName = "bigbang"

type versionCmdHelper struct {
	constants      static.Constants
	config         *config.GlobalConfiguration
	helmClient     helm.Client
	outputClient   output.Client
	logger         bbLog.Client
	gitlabClient   gitlab.Client
	kubeClient     k8sClient.Interface
	ironbankClient ironbank.Client
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

	ironbankClient, errr := factory.GetIronBankClient()
	if errr != nil {
		return nil, fmt.Errorf("error getting ironbank client: %w", errr)
	}

	return &versionCmdHelper{
		constants:      constants,
		config:         config,
		helmClient:     helmClient,
		outputClient:   outputClient,
		logger:         logger,
		gitlabClient:   gitlabClient,
		kubeClient:     kubeClient,
		ironbankClient: ironbankClient,
	}, nil
}

// NewVersionCmd - Creates a new Cobra command which implements the `bbctl version` functionality
func NewVersionCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	var (
		versionUse   = `version`
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

	err = configClient.SetAndBindFlag(
		"no-shas",
		"",
		false,
		"Disable checking the deployed image SHAs against the upstream repo",
	)
	if err != nil {
		return nil, fmt.Errorf("error setting and binding no-shas flag: %w", err)
	}

	return cmd, nil
}

// bbVersion is a helper function to separate a lot of the args and config logic
// from the command function creation for easier unit testing
func (v *versionCmdHelper) bbVersion(args []string) error {
	var err error
	var outputMap schemas.VersionOutput

	// Short circuit if the user only wants the bbctl client version
	if v.config.VersionConfiguration.Client {
		outputMap = schemas.VersionOutput{}
		outputMap["bbctl"] = schemas.PackageVersion{
			Version: v.constants.BigBangCliVersion,
		}

		return v.outputClient.Output(&outputMap)
	}

	if v.config.VersionConfiguration.AllCharts {
		outputMap, err = v.getAllChartVersions(v.config.VersionConfiguration.CheckForUpdates, v.config.VersionConfiguration.NoSHAs)
		if err != nil {
			return fmt.Errorf("error getting all chart versions: %w", err)
		}
		return v.outputClient.Output(&outputMap)
	}

	switch len(args) {
	// If no arguments are provided, print the version of the Big Bang release
	// and the bbctl client
	case 0:
		if v.config.VersionConfiguration.CheckForUpdates {
			outputMap = schemas.VersionOutput{}
			packageInfo, err := v.checkForUpdates(v.constants.BigBangHelmReleaseName)
			if err != nil {
				return fmt.Errorf("error checking for updates: %w", err)
			}
			outputMap[v.constants.BigBangHelmReleaseName] = packageInfo
		} else {
			outputMap, err = v.outputBigBangVersion()
			if err != nil {
				return fmt.Errorf("error getting Big Bang version: %w", err)
			}
		}
	// If an argument is provided, print the version of the specific release named
	case 1:
		chartName := args[0]
		chartVersion, targetNamespace, err := v.getChartVersion(chartName)
		if err != nil {
			return fmt.Errorf("error getting chart version: %w", err)
		}

		outputMap = schemas.VersionOutput{}
		outputMap[chartName] = schemas.PackageVersion{
			Version: chartVersion,
		}

		if v.config.VersionConfiguration.CheckForUpdates {
			packageInfo, err := v.checkForUpdates(chartName)
			if err != nil {
				return fmt.Errorf("error checking for updates: %w", err)
			}
			outputMap[chartName] = packageInfo
		}

		// Ignore bigbang since bigbang itself doesn't deploy images
		if !v.config.VersionConfiguration.NoSHAs && chartName != BigBangChartName {
			message, err := v.matchSHAs(chartName, chartVersion, targetNamespace)
			if err != nil {
				return fmt.Errorf("error checking for SHAs: %w", err)
			}

			temp := outputMap[chartName]
			temp.SHAsMatch = message
			outputMap[chartName] = temp
		}

	default:
		return errors.New("invalid number of arguments")
	}

	return v.outputClient.Output(&outputMap)
}

// getAllChartVersions gets the version of all the charts in the cluster owned by Big Bang
func (v *versionCmdHelper) getAllChartVersions(checkForUpdates bool, noSHAs bool) (schemas.VersionOutput, error) {
	var output schemas.VersionOutput

	customResource := schema.GroupVersionResource{Group: "helm.toolkit.fluxcd.io", Version: "v2", Resource: "helmreleases"}
	opts := metaV1.ListOptions{}
	v.logger.Debug("getting all charts' versions")
	releases, err := v.kubeClient.Resource(customResource).Namespace(v.constants.BigBangNamespace).List(context.TODO(), opts)
	if err != nil {
		return output, fmt.Errorf("error getting helmreleases: %w", err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(releases.Items)+2)

	// bigbang isn't installed as a helm release in most distributions, so we need to handle it separately
	// but we want to check first that it's not in the list of releases
	bigbangFound := false
	for _, release := range releases.Items {
		// if a release is named bigbang (like in instances where kustomize is used),
		// we can assume it will be added to the map with the rest of the releases
		if release.Object["metadata"].(map[string]any)["name"].(string) == v.constants.BigBangHelmReleaseName {
			bigbangFound = true
			break
		}
	}

	outputMap := schemas.VersionOutput{}

	if !bigbangFound {
		wg.Add(1)
		go func() {
			defer wg.Done()

			version, err := v.getBigBangVersion()
			if err != nil {
				errChan <- fmt.Errorf("error getting Big Bang version: %w", err)
				return
			}
			p := schemas.PackageVersion{
				Version: version,
			}

			if checkForUpdates {
				latestVersion, err := v.getLatestChartVersion(v.constants.BigBangHelmReleaseName)
				if err != nil {
					errChan <- fmt.Errorf("error getting latest chart version: %w", err)
					return
				}
				update, err := updateIsNewer(version, latestVersion)
				if err != nil {
					errChan <- fmt.Errorf("error checking for update: %w", err)
					return
				}

				p.LatestVersion = latestVersion
				p.UpdateAvailable = update
			}

			mu.Lock()
			outputMap[v.constants.BigBangHelmReleaseName] = p
			mu.Unlock()
		}()
	}

	for _, release := range releases.Items {
		name := release.Object["metadata"].(map[string]any)["name"].(string)
		version := release.Object["status"].(map[string]any)["history"].([]any)[0].(map[string]any)["chartVersion"].(string)
		targetNamespace := release.Object["spec"].(map[string]any)["targetNamespace"].(string)

		wg.Add(1)
		go func(name, version string) {
			defer wg.Done()

			p := schemas.PackageVersion{
				Version: version,
			}

			if checkForUpdates {
				latestVersion, err := v.getLatestChartVersion(name)
				if err != nil {
					errChan <- fmt.Errorf("error getting latest chart version: %w", err)
					return
				}
				update, err := updateIsNewer(version, latestVersion)
				if err != nil {
					errChan <- fmt.Errorf("error checking for update: %w", err)
					return
				}
				p.LatestVersion = latestVersion
				p.UpdateAvailable = update
			}

			if !noSHAs {
				SHAsMatch, err := v.matchSHAs(name, version, targetNamespace)
				if err != nil {
					errChan <- fmt.Errorf("error checking for SHAs: %w", err)
					return
				}
				p.SHAsMatch = SHAsMatch
			}

			mu.Lock()
			outputMap[name] = p
			mu.Unlock()
		}(name, version)
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

// getBigBangVersion gets the version of the Big Bang release currently deployed
func (v *versionCmdHelper) getBigBangVersion() (string, error) {
	bigbangVersion, err := v.getReleaseVersion(v.constants.BigBangHelmReleaseName)
	if err != nil {
		return "", fmt.Errorf("error fetching Big Bang release version: %w", err)
	}
	return bigbangVersion, nil
}

// outputBigBangVersion outputs the current version of the Big Bang release and the bbctl client
func (v *versionCmdHelper) outputBigBangVersion() (schemas.VersionOutput, error) {
	outputMap := schemas.VersionOutput{}

	bigbangVersion, err := v.getBigBangVersion()
	if err != nil {
		return outputMap, fmt.Errorf("error getting Big Bang version: %w", err)
	}

	outputMap[v.constants.BigBangHelmReleaseName] = schemas.PackageVersion{
		Version:   bigbangVersion,
		SHAsMatch: "",
	}

	outputMap["bbctl"] = schemas.PackageVersion{
		Version:   v.constants.BigBangCliVersion,
		SHAsMatch: "",
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

// getChartVersion gets the version of a chart by the chart name and also returns the target namespace for the chart
func (v *versionCmdHelper) getChartVersion(chartName string) (string, string, error) {
	if chartName == BigBangChartName {
		version, err := v.getBigBangVersion()
		return version, v.constants.BigBangNamespace, err
	}
	// We want to find the release name from the chart name, since the release name
	// is less obvious to the end user. To to this, we'll fetch all the releases and iterate over them
	// looking at the associated chart name until we find the one we're looking for.

	customResource := schema.GroupVersionResource{Group: "helm.toolkit.fluxcd.io", Version: "v2", Resource: "helmreleases"}
	opts := metaV1.GetOptions{}

	resource, err := v.kubeClient.Resource(customResource).Namespace(v.constants.BigBangNamespace).Get(context.TODO(), chartName, opts)
	if err != nil {
		return "", "", fmt.Errorf("error getting helmreleases: %w", err)
	}

	version := resource.Object["status"].(map[string]any)["history"].([]any)[0].(map[string]any)["chartVersion"].(string)
	if version == "" {
		return "", "", fmt.Errorf(`error getting version for release "%s": no version specified for the chart`, chartName)
	}

	targetNamespace := resource.Object["spec"].(map[string]any)["targetNamespace"].(string)
	if targetNamespace == "" {
		return "", "", fmt.Errorf(`error getting target namespace for release "%s": no target namespace specified for the chart`, chartName)
	}

	return version, targetNamespace, nil
}

// checkForChartUpdate checks the current chart version against the latest version available on repo1
func (v *versionCmdHelper) getLatestChartVersion(chartName string) (string, error) {
	v.logger.Debug("checking for update to " + chartName)

	packageURI, branch, err := v.getChartURL(chartName)
	if err != nil {
		return "", fmt.Errorf("error getting chart URL: %w", err)
	}

	// Fetch the latest Chart.yalm from the upstream repo
	body, err := v.gitlabClient.GetFile(packageURI, "chart/Chart.yaml", branch)
	if err != nil {
		return "", fmt.Errorf("error getting Chart.yaml: %w", err)
	}

	helmChart, err := decodeChartYAML(body)
	if err != nil {
		return "", fmt.Errorf("failed to decode Chart.yaml: %w", err)
	}

	return helmChart.Version, nil
}

func (v *versionCmdHelper) checkForUpdates(chartName string) (schemas.PackageVersion, error) {
	var outputMap schemas.PackageVersion

	latestVersion, err := v.getLatestChartVersion(chartName)
	if err != nil {
		return outputMap, fmt.Errorf("error checking for latest chart version: %w", err)
	}

	currentVersion, _, err := v.getChartVersion(chartName)
	if err != nil {
		return outputMap, fmt.Errorf("error getting current chart version: %w", err)
	}

	update, err := updateIsNewer(currentVersion, latestVersion)
	if err != nil {
		return outputMap, fmt.Errorf("error checking for update: %w", err)
	}

	outputMap = schemas.PackageVersion{
		UpdateAvailable: update,
		Version:         currentVersion,
		LatestVersion:   latestVersion,
	}

	return outputMap, nil
}

// getChartImages gets the images.txt associated with a chart at a specific release tag
func (v *versionCmdHelper) getChartImages(chartName, releaseTag string) ([]string, error) {
	var images []string

	projectPath, _, err := v.getChartURL(chartName)
	if err != nil {
		return images, fmt.Errorf("error getting chart URL: %w", err)
	}

	project, err := v.gitlabClient.GetProject(projectPath)
	if err != nil {
		return images, fmt.Errorf("error getting project: %w", err)
	}

	body, err := v.gitlabClient.GetReleaseArtifact(project.ID, releaseTag, "images.txt")
	if err != nil {
		return images, fmt.Errorf("error getting images.txt: %w", err)
	}

	return strings.Split(strings.TrimSpace((string(body))), "\n"), nil
}

func (v *versionCmdHelper) getChartURL(chartName string) (string, string, error) {
	var packageURI, branch string

	// Special consideraiotns for the bigbang chart
	if chartName == BigBangChartName {
		packageURI = "big-bang/bigbang"
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
			return packageURI, branch, fmt.Errorf("error getting gitrepositories: %w", err)
		}

		// Parse out the package URL from the GitRepo CRD
		packageURI = strings.TrimSuffix(resource.Object["spec"].(map[string]any)["url"].(string), ".git")
		packageURI = strings.TrimPrefix(packageURI, "https://repo1.dso.mil/")
	}

	return packageURI, branch, nil
}

// getUpstreamSHAsForChartRelease gets the SHAs of the images of a chart release by downloading the published images.txt
// and then getting the SHAs of the images from Iron Bank
func (v *versionCmdHelper) getUpstreamSHAsForChartRelease(chartName, releaseTag string) (map[string]string, error) {
	imageMap := map[string]string{}

	images, err := v.getChartImages(chartName, releaseTag)
	if err != nil {
		return imageMap, fmt.Errorf("error getting chart images: %w", err)
	}
	for _, image := range images {
		sha, err := v.ironbankClient.GetImageSHA(image)
		if err != nil {
			return imageMap, fmt.Errorf("error getting image SHA: %w", err)
		}
		imageMap[image] = sha
	}

	return imageMap, nil
}

// getSHAsForCurrentPods gets the SHAs of the images of the pods currently running in the cluster
// for a given chart
//
// Filters for images that from IronBank only
func (v *versionCmdHelper) getSHAsForCurrentPods(targetNamespace string) (map[string]string, error) {
	imageMap := map[string]string{}

	pods, err := v.kubeClient.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}).Namespace(targetNamespace).List(context.TODO(), metaV1.ListOptions{})

	if err != nil {
		return imageMap, fmt.Errorf("error getting pods: %w", err)
	}

	for _, pod := range pods.Items {
		containerStatuses, found, err := unstructured.NestedSlice(pod.Object, "status", "containerStatuses")
		if err != nil || !found {
			continue
		}
		for _, cs := range containerStatuses {
			csMap, ok := cs.(map[string]interface{})
			if !ok {
				continue
			}
			image, found, _ := unstructured.NestedString(csMap, "image")
			if !found {
				continue
			}
			imageID, found, _ := unstructured.NestedString(csMap, "imageID")
			if !found {
				continue
			}
			// Extract SHA from imageID
			parts := strings.Split(imageID, "@sha256:")
			if len(parts) == 2 {
				imageMap[image] = parts[1]
			}
		}
	}

	return imageMap, nil
}

// matchSHAs queries the cluster for the SHAs of the images of the pods currently running in the cluster and then compares them to the published SHAs provided by Iron Bank for a given release.
//
// If the SHAs match, it returns a message indicating that the SHAs match and the SHAs are the same.
// If the SHAs do not match, it returns a message indicating that the SHAs do not match and the SHAs are different.
// If the SHAs cannot be matched, it notifies the user that the SHAs could not be matched
func (v *versionCmdHelper) matchSHAs(chartName, chartVersion, targetNamespace string) (string, error) {
	var messageBuilder strings.Builder
	var hasErrors bool

	current, err := v.getSHAsForCurrentPods(targetNamespace)
	if err != nil {
		return "", fmt.Errorf("error getting SHAs for current pods: %w", err)
	}
	upstream, err := v.getUpstreamSHAsForChartRelease(chartName, chartVersion)
	if err != nil {
		return "", fmt.Errorf("error getting SHAs for upstream pods: %w", err)
	}

	for image, sha := range current {
		upstreamSha, ok := upstream[image]
		if !ok {
			hasErrors = true
			messageBuilder.WriteString(fmt.Sprintf(`Error: SHA for running container "%s" not found in published release artifacts`, image))
			continue
		}

		if sha != upstreamSha {
			hasErrors = true
			messageBuilder.WriteString(fmt.Sprintf(`Error: SHA mismatch for image "%s". Local: "%s", upstream: "%s"`, image, upstreamSha, sha))
		}
	}

	if hasErrors {
		return messageBuilder.String(), nil
	}

	return "All SHAs match", nil
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

// helmChartManifest is a struct that partially represents the Chart.yaml file
type helmChartManifest struct {
	Version string `yaml:"version"`
}

// decodeChartYAML decodes the Chart.yaml file body from GitLab into a type
func decodeChartYAML(fileBody []byte) (*helmChartManifest, error) {
	var chart helmChartManifest
	err := yamler.Unmarshal(fileBody, &chart)
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
