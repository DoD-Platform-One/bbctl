package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	mock "repo1.dso.mil/big-bang/product/packages/bbctl/mocks/repo1.dso.mil/big-bang/product/packages/bbctl/static"
	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	bbTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8sTesting "k8s.io/client-go/testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

const bigBangRepo = "big-bang/bigbang"

func TestGetVersionUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
	assert.Contains(t, cmd.Example, "bbctl version --client")
}

// parseYAMLOutput takes a YAML string and returns a map[string]any{}
func parseYAMLOutput(t *testing.T, yamlString string) map[string]any {
	t.Helper()
	// Trim any leading or trailing whitespace
	yamlString = strings.TrimSpace(yamlString)

	// Parse the YAML string
	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlString), &result)
	if err != nil {
		t.Errorf("error parsing YAML: %s", err.Error())
	}

	return result
}

func TestGetVersion(t *testing.T) {
	// Arrange
	chartInfo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bigbang",
			Version: "1.0.2",
		},
	}

	releaseFixture := []*release.Release{
		{
			Name:      "bigbang",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: chartInfo,
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")

	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	// Act
	cmd, _ := NewVersionCmd(factory)
	res := cmd.RunE(cmd, []string{})
	require.NoError(t, res)
	outputMap := parseYAMLOutput(t, buf.String())

	constants, err := static.GetDefaultConstants()
	require.NoError(t, err)

	assert.Equal(t, map[string]any{
		"bigbang": map[any]any{"version": "1.0.2"},
		"bbctl":   map[any]any{"version": constants.BigBangCliVersion},
	}, outputMap)
}

func TestGetVersionClientVersionOnly(t *testing.T) {
	// Arrange
	chartInfo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bigbang",
			Version: "1.0.2",
		},
	}

	releaseFixture := []*release.Release{
		{
			Name:      "bigbang",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: chartInfo,
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("client", true)
	v.Set("output-config.format", "yaml")

	streams, _ := factory.GetIOStream()
	buf := streams.Out.(*bytes.Buffer)

	// Act
	cmd, _ := NewVersionCmd(factory)
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	outputMap := parseYAMLOutput(t, buf.String())

	constants, err := static.GetDefaultConstants()
	require.NoError(t, err)

	assert.Equal(t, map[string]any{
		"bbctl": map[any]any{"version": constants.BigBangCliVersion},
	}, outputMap)
}

func TestGetVersionInvalidClientFlag(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)
	cmd.SetArgs([]string{"--client=string-value"})
	err := cmd.Execute()

	// Assert
	require.Error(t, err)
	if !assert.Contains(t, err.Error(), "invalid argument \"string-value\" for \"--client\" flag") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestGetVersionWithError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)
	err := cmd.RunE(cmd, []string{})

	// Assert
	expected := "error getting Big Bang version: error getting Big Bang version: error fetching Big Bang release version: error getting helm information for release bigbang: release bigbang not found"
	require.Error(t, err)
	assert.Equal(t, expected, err.Error())
}

func TestGetVersionWithBadParams(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	cmd, _ := NewVersionCmd(factory)
	cmd.SetArgs([]string{"--invalid-parameter"})
	res := cmd.Execute()

	// Assert
	require.Error(t, res)
	assert.Equal(t, "unknown flag: --invalid-parameter", res.Error())
}

func TestGetVersionWithConfigError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	factory.SetFail.GetConfigClient = 1
	cmd, err := NewVersionCmd(factory)
	assert.Equal(t, "unable to get config client: failed to get config client", err.Error())

	helper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, helper)
	assert.Equal(t, "unable to get config client: failed to get config client", err.Error())
}

func TestGetVersionHelperWithLoggerError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	// Act
	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	// Set a config client to avoid this from failing on getting logging client
	getConfigFunc := func(_ *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: "test",
		}, nil
	}

	logger, err := factory.GetLoggingClient()
	require.NoError(t, err)

	client, err := bbConfig.NewClient(getConfigFunc, nil, &logger, cmd, v)
	require.NoError(t, err)

	factory.SetConfigClient(client)

	// Set constants client to fail now
	expectedError := errors.New("failed to get constants")
	constantsClient := mock.MockConstantsClient{}
	constantsClient.On("GetConstants").Return(static.Constants{BigBangNamespace: "bigbang"}, expectedError)

	versionHelper, err := newVersionCmdHelper(cmd, factory, &constantsClient)

	// Assert
	assert.Equal(t, expectedError.Error(), err.Error())
	assert.Nil(t, versionHelper)
	// Check that expected methods were called
	constantsClient.AssertExpectations(t)
}

func TestGetVersionHelperWithConstantsClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	// Act
	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	// Set a config client to avoid this from failing on getting logging client
	getConfigFunc := func(_ *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: "test",
		}, nil
	}

	logger, err := factory.GetLoggingClient()
	require.NoError(t, err)

	client, err := bbConfig.NewClient(getConfigFunc, nil, &logger, cmd, v)
	require.NoError(t, err)

	factory.SetConfigClient(client)

	// Set logging client to fail now
	factory.SetFail.GetLoggingClient = true
	versionHelper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)

	// Assert
	assert.Nil(t, versionHelper)
	assert.Equal(t, "error getting logging client: failed to get logging client", err.Error())
}

func TestGetVersionHelperWithGitlabError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	factory.SetFail.GetGitLabClient = true
	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	helper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, helper)
	assert.Equal(t, "error getting gitlab client: failed to get GitLab client", err.Error())
}

func TestGetVersionHelperWithKubeError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	factory.SetFail.GetK8sDynamicClient = true
	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	helper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, helper)
	assert.Equal(t, "error getting k8s client: failed to get K8sDynamicClient client", err.Error())
}

func TestGetVersionWithHelmError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	factory.SetFail.GetHelmClient = true
	cmd, _ := NewVersionCmd(factory)
	err := cmd.RunE(cmd, []string{})

	// Assert
	require.Error(t, err)
	assert.Equal(t, "error creating version helper: failed to get helm client", err.Error())
}

func TestGetVersionWithIronBankError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetHelmReleases(nil)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	// Act
	factory.SetFail.GetIronBankClient = true
	cmd, _ := NewVersionCmd(factory)
	err := cmd.RunE(cmd, []string{})

	// Assert
	require.Error(t, err)
	assert.Equal(t, "error creating version helper: error getting ironbank client: failed to get ironbank client", err.Error())
}

func TestVersionFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd, _ := NewVersionCmd(factory)
	viper, _ := factory.GetViper()
	expected := ""
	getConfigFunc := func(_ *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, errors.New("dummy error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, h)

	// Assert
	require.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestVersionErrorBindingFlags(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	expectedError := errors.New("failed to set and bind flag")
	logClient, _ := factory.GetLoggingClient()

	tests := []struct {
		flagName       string
		failOnCallNum  int
		expectedCmd    bool
		expectedErrMsg string
	}{
		{
			flagName:       "client",
			failOnCallNum:  1,
			expectedCmd:    false,
			expectedErrMsg: "error setting and binding client flag: failed to set and bind flag",
		},
		{
			flagName:       "all-charts",
			failOnCallNum:  2,
			expectedCmd:    false,
			expectedErrMsg: "error setting and binding all-charts flag: failed to set and bind flag",
		},
		{
			flagName:       "check-for-updates",
			failOnCallNum:  3,
			expectedCmd:    false,
			expectedErrMsg: "error setting and binding check-for-updates flag: failed to set and bind flag",
		},
		{
			flagName:       "no-shas",
			failOnCallNum:  4,
			expectedCmd:    false,
			expectedErrMsg: "error setting and binding no-shas flag: failed to set and bind flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			callCount := 0
			setAndBindFlagFunc := func(_ *bbConfig.ConfigClient, _ string, _ string, _ any, _ string) error {
				callCount++
				if callCount == tt.failOnCallNum {
					return expectedError
				}
				return nil
			}

			configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, v)
			require.NoError(t, err)
			factory.SetConfigClient(configClient)

			// Act
			cmd, err := NewVersionCmd(factory)

			// Assert
			if tt.expectedCmd {
				assert.NotNil(t, cmd)
			} else {
				assert.Nil(t, cmd)
			}

			if tt.expectedErrMsg != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVersionOutputClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetIOStreams = 1
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, h)
	assert.Equal(t, "error getting output client: failed to get streams", err.Error())
}

func TestClientVersionMarshalError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("client", true)
	v.Set("output-config.format", "")
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, true)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	// Act
	err = h.bbVersion([]string{})

	// Assert
	require.Error(t, err)
	expectedError := "unable to write human-readable output: FakeWriter intentionally errored"
	assert.Equal(t, expectedError, err.Error())
	assert.Empty(t, fakeWriter.ActualBuffer.(*bytes.Buffer).String())
}

func TestClientandBBVersionMarshalError(t *testing.T) {
	// Arrange
	chartInfo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bigbang",
			Version: "1.0.2",
		},
	}

	releaseFixture := []*release.Release{
		{
			Name:      "bigbang",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: chartInfo,
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "")
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, true)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	// Act
	err = h.bbVersion([]string{})

	// Assert
	expectedError := "unable to write human-readable output: FakeWriter intentionally errored"
	assert.Empty(t, fakeWriter.ActualBuffer.(*bytes.Buffer).String())
	assert.Equal(t, expectedError, err.Error())
}

func TestBBVersionAllCharts(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("output-config.format", "yaml")
	v.Set("big-bang-repo", "test")
	v.Set("all-charts", true)
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)
	require.NoError(t, h.bbVersion([]string{}))

	output := fakeWriter.ActualBuffer.(*bytes.Buffer).String()

	outputMap := parseYAMLOutput(t, output)

	assert.Equal(t, map[string]any{
		"bigbang": map[any]any{
			"version": "1.0.2",
		},
		"grafana": map[any]any{
			"version":   "1.0.3",
			"shasMatch": "All SHAs match",
		},
		"tempo": map[any]any{
			"version":   "3.2.1",
			"shasMatch": "All SHAs match",
		},
	}, outputMap)
}

func TestBBVersionErrorGettingAllCharts(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("all-charts", true)
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &errorListHelmReleasesFunc)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	err = h.bbVersion([]string{})
	require.Error(t, err)
	assert.Equal(t, "error getting all chart versions: error getting helmreleases: error in list crds", err.Error())
}

func TestBBVersionNoArgsErrorCheckingForUpdates(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("check-for-updates", true)
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGitLabGetFileFunc(func(_, _, _ string) ([]byte, error) {
		return nil, errors.New("dummy error")
	})

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	err = h.bbVersion([]string{})

	require.Error(t, err)
	assert.Equal(t, "error checking for updates: error checking for latest chart version: error getting Chart.yaml: dummy error", err.Error())
}

func TestBBVersionWithChartName(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	err = h.bbVersion([]string{"grafana"})
	require.NoError(t, err)

	output := fakeWriter.ActualBuffer.(*bytes.Buffer).String()
	outputMap := parseYAMLOutput(t, output)
	assert.Equal(t, map[string]any{
		"grafana": map[any]any{
			"version":   "1.0.3",
			"shasMatch": "All SHAs match",
		},
	}, outputMap)
}

func TestBBVersionWithChartNameErrorGettingChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGVRToListKind(versionTestGVR)

	// Create a release with no version set
	releases := &unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": "helm.toolkit.fluxcd.io/v2",
			"kind":       "HelmRelease",
		},
		Items: []unstructured.Unstructured{newHelmRelease("grafana", "bigbang", "", "monitoring")},
	}

	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), releases})
	factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &errorListHelmReleasesFunc)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	err = h.bbVersion([]string{"grafana"})
	assert.Equal(t, `error getting chart version: error getting version for release "grafana": no version specified for the chart`, err.Error())
}

func TestBBVersionWithChartNameErrorCheckingForUpdates(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("check-for-updates", true)
	streams, err := factory.GetIOStream()
	require.NoError(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetGitLabGetFileFunc(func(_, _, _ string) ([]byte, error) {
		return nil, errors.New("dummy error")
	})

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	err = h.bbVersion([]string{"grafana"})
	assert.Equal(t, `error checking for updates: error checking for latest chart version: error getting Chart.yaml: dummy error`, err.Error())
}

func TestBBVersionWithInvalidNumberOfArguments(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	streams, err := factory.GetIOStream()
	require.NoError(t, err)

	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	err = h.bbVersion([]string{"too", "many", "arguments"})
	assert.Equal(t, "invalid number of arguments", err.Error())
}

func TestGetReleaseVersionNoVersionSet(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	relases := []*release.Release{
		{
			Name:      "grafana",
			Namespace: "grafana",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Name:    "grafana",
					Version: "",
				},
			},
		},
	}
	factory.SetHelmReleases(relases)

	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	version, err := h.getReleaseVersion("grafana")
	assert.Empty(t, version)
	assert.Equal(t, `error getting version for release "grafana": no version specified for the chart`, err.Error())
}

func buildHelmReleasesFixture() []*release.Release {
	bigBangChartInfo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "bigbang",
			Version: "1.0.2",
		},
	}

	grafanaChartInfo := &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "grafana",
			Version: "1.0.3",
		},
	}

	return []*release.Release{
		{
			Name:      "bigbang",
			Version:   1,
			Namespace: "bigbang",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: bigBangChartInfo,
		},
		{
			Name:      "grafana",
			Version:   1,
			Namespace: "grafana",
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: grafanaChartInfo,
		},
	}
}

var versionTestGVR = map[runtimeSchema.GroupVersionResource]string{
	{
		Group:    "source.toolkit.fluxcd.io",
		Version:  "v1beta1",
		Resource: "gitrepositories",
	}: "GitRepositoryList",
	{
		Group:    "helm.toolkit.fluxcd.io",
		Version:  "v2",
		Resource: "helmreleases",
	}: "HelmReleaseList",
	{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}: "PodList",
}

var errorListHelmReleasesFunc = func(client *dynamicFake.FakeDynamicClient) {
	client.Fake.PrependReactor("list", "helmreleases", func(_ k8sTesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("error in list crds")
	})
}

func buildGitRepoFixture() *unstructured.UnstructuredList {
	return &unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": "source.toolkit.fluxcd.io/v1",
			"kind":       "GitRepository",
			"metadata": map[string]any{
				"name":      "bigbang",
				"namespace": "bigbang",
			},
		},
		Items: []unstructured.Unstructured{
			{
				Object: map[string]any{
					"apiVersion": "source.toolkit.fluxcd.io/v1",
					"kind":       "GitRepository",
					"metadata": map[string]any{
						"name":      "bigbang",
						"namespace": "bigbang",
					},
					"spec": map[string]any{
						"url": "https://github.com/repo1/bigbang",
					},
				},
			},
			{
				Object: map[string]any{
					"apiVersion": "source.toolkit.fluxcd.io/v1",
					"kind":       "GitRepository",
					"metadata": map[string]any{
						"name":      "grafana",
						"namespace": "bigbang",
					},
					"spec": map[string]any{
						"url": "https://repo1.dso.mil/big-bang/product/packages/grafana.git",
					},
				},
			},
			{
				Object: map[string]any{"apiVersion": "source.toolkit.fluxcd.io/v1",
					"kind": "GitRepository",
					"metadata": map[string]any{
						"name":      "tempo",
						"namespace": "bigbang",
					},
					"spec": map[string]any{
						"url": "https://repo1.dso.mil/big-bang/product/packages/tempo.git",
					},
				},
			},
		},
	}
}

//nolint:unparam
func newHelmRelease(name, namespace, chartVersion, targetNamespace string) unstructured.Unstructured {
	return unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "helm.toolkit.fluxcd.io/v2",
			"kind":       "HelmRelease",
			"metadata": map[string]any{
				"name":      name,
				"namespace": namespace,
			},
			"status": map[string]any{
				"history": []any{
					map[string]any{
						"chartVersion": chartVersion,
					},
				},
			},
			"spec": map[string]any{
				"targetNamespace": targetNamespace,
			},
		},
	}
}

func buildHelmReleaseFixture(includeBigBang bool) *unstructured.UnstructuredList {
	items := []unstructured.Unstructured{
		newHelmRelease("grafana", "bigbang", "1.0.3", "monitoring"),
		newHelmRelease("tempo", "bigbang", "3.2.1", "tempo"),
	}

	if includeBigBang {
		items = append(items, newHelmRelease("bigbang", "bigbang", "1.0.2", "bigbang"))
	}

	return &unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": "helm.toolkit.fluxcd.io/v2",
			"kind":       "HelmRelease",
		},
		Items: items,
	}
}

func newPod(name, namespace, image, sha string) corev1.Pod {
	return corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  name + "-1",
					Image: image,
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Image:   image,
					ImageID: "docker-pullable://image@sha256:" + sha,
				},
			},
		},
	}
}

func buildPodFixtures() *corev1.PodList {
	items := []corev1.Pod{
		newPod("monitoring-grafana-abc-123", "monitoring", "grafana:1.0.3", "1234567890"),
	}

	return &corev1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodList",
			APIVersion: "v1",
		},
		Items: items,
	}
}

func TestGetAllChartVersions(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	output, err := h.getAllChartVersions(false, false)
	require.NoError(t, err)

	outputMap := output.ToMap()

	assert.Equal(t, "1.0.2", outputMap["bigbang"].(map[string]any)["version"])
	assert.Equal(t, "1.0.3", outputMap["grafana"].(map[string]any)["version"])
	assert.Equal(t, "3.2.1", outputMap["tempo"].(map[string]any)["version"])
}

func TestGetAllChartVersionsBigBangManagedByFlux(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(true)})
	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")

	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	output, err := h.getAllChartVersions(false, false)
	require.NoError(t, err)

	outputMap := output.ToMap()

	assert.Equal(t, "1.0.2", outputMap["bigbang"].(map[string]any)["version"])
	assert.Equal(t, "1.0.3", outputMap["grafana"].(map[string]any)["version"])
	assert.Equal(t, "3.2.1", outputMap["tempo"].(map[string]any)["version"])
}

func TestGetAllChartVersionsErrorGettingBigBangVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	output, err := h.getAllChartVersions(false, false)
	assert.Equal(t, "error getting Big Bang version: error fetching Big Bang release version: error getting helm information for release bigbang: release bigbang not found", err.Error())

	outputMap := output.ToMap()
	assert.Equal(t, "1.0.3", outputMap["grafana"].(map[string]any)["version"])
	assert.Equal(t, "3.2.1", outputMap["tempo"].(map[string]any)["version"])
}

func TestGetAllChartVersionsErrorListingReleases(t *testing.T) {
	releaseFixture := buildHelmReleasesFixture()

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &errorListHelmReleasesFunc)

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	outputMap, err := h.getAllChartVersions(false, false)
	assert.Equal(t, "error getting helmreleases: error in list crds", err.Error())
	assert.Empty(t, outputMap)
}

func TestGetAllChartVersionsCheckForUpdates(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		switch repository {
		case bigBangRepo:
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "master", branch)
			return []byte("version: 1.0.2"), nil
		case "big-bang/product/packages/grafana":
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "main", branch)
			return []byte("version: 2.0.0"), nil
		case "big-bang/product/packages/tempo":
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "main", branch)
			return []byte("version: 3.2.2"), nil
		}

		return nil, errors.New("invalid repository")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")

	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	output, err := h.getAllChartVersions(true, false)
	require.NoError(t, err)

	outputMap := output.ToMap()

	assert.Equal(t, map[string]any{
		"grafana": map[string]any{
			"version":         "1.0.3",
			"latestVersion":   "2.0.0",
			"updateAvailable": true,
			"shasMatch":       "All SHAs match",
		},
		"bigbang": map[string]any{
			"version":         "1.0.2",
			"latestVersion":   "1.0.2",
			"updateAvailable": false,
		},
		"tempo": map[string]any{
			"version":         "3.2.1",
			"latestVersion":   "3.2.2",
			"updateAvailable": true,
			"shasMatch":       "All SHAs match",
		},
	}, outputMap)
}

func TestGetAllChartVersionsCheckForUpdatesErrorGettingLatestChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetGitLabGetFileFunc(func(_, _, _ string) ([]byte, error) {
		return nil, errors.New("dummy error")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	outputMap, err := h.getAllChartVersions(true, false)
	assert.Equal(t, "error getting latest chart version: error getting Chart.yaml: dummy error", err.Error())
	assert.Empty(t, outputMap)
}

func TestGetAllChartVersionsCheckForUpdatesErrorCheckingIfBigBangUpdateIsNewer(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		switch repository {
		case bigBangRepo:
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "master", branch)
			return []byte("version: invalid"), nil
		case "big-bang/product/packages/grafana":
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "main", branch)
			return []byte("version: 2.0.0"), nil
		case "big-bang/product/packages/tempo":
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "main", branch)
			return []byte("version: 3.2.2"), nil
		}
		return nil, errors.New("invalid repository")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	output, err := h.getAllChartVersions(true, false)
	assert.Equal(t, `error checking for update: invalid version "invalid": Invalid Semantic Version`, err.Error())
	outputMap := output.ToMap()
	assert.Equal(t, map[string]any{
		"latestVersion":   "2.0.0",
		"updateAvailable": true,
		"version":         "1.0.3",
		"shasMatch":       "All SHAs match",
	}, outputMap["grafana"])
	assert.Empty(t, outputMap["bigbang"])
}

func TestGetAllChartVersionsCheckForUpdatesErrorCheckingIfPackageUpdateIsNewer(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		switch repository {
		case bigBangRepo:
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "master", branch)
			return []byte("version: 1.0.2"), nil
		case "big-bang/product/packages/grafana":
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "main", branch)
			return []byte("version: 2.0.0"), nil
		case "big-bang/product/packages/tempo":
			assert.Equal(t, "chart/Chart.yaml", path)
			assert.Equal(t, "main", branch)
			return []byte("version: invalid"), nil
		}
		return nil, errors.New("invalid repository")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	output, err := h.getAllChartVersions(true, false)
	assert.Equal(t, `error checking for update: invalid version "invalid": Invalid Semantic Version`, err.Error())

	outputMap := output.ToMap()
	assert.Equal(t, map[string]any{
		"latestVersion":   "2.0.0",
		"updateAvailable": true,
		"version":         "1.0.3",
		"shasMatch":       "All SHAs match",
	}, outputMap["grafana"])
	assert.Empty(t, outputMap["tempo"])
}

func TestGetChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	version, namespace, err := h.getChartVersion("bigbang")
	require.NoError(t, err)
	assert.Equal(t, "1.0.2", version)
	assert.Equal(t, "bigbang", namespace)

	// TODO check namespacs
	version, _, err = h.getChartVersion("invalid-chart")
	assert.Empty(t, version)
	assert.Equal(t, `error getting helmreleases: helmreleases.helm.toolkit.fluxcd.io "invalid-chart" not found`, err.Error())
}

func TestGetLatestChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, "big-bang/product/packages/grafana", repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "main", branch)
		return []byte("version: 1.0.3"), nil
	})

	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	version, err := h.getLatestChartVersion("grafana")
	require.NoError(t, err)
	assert.Equal(t, "1.0.3", version)
}

// TestGetLatestChartVersionBigBang tests that the special internal conditions required to check the version of Big Bang are met
func TestGetLatestChartVersionBigBang(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, bigBangRepo, repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "master", branch)
		return []byte("version: 1.0.3"), nil
	})

	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	version, err := h.getLatestChartVersion("bigbang")
	require.NoError(t, err)
	assert.Equal(t, "1.0.3", version)
}

func TestGetLatestChartVersionErrorListingGitRepos(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects(nil)

	factory.SetGitLabGetFileFunc(func(_, _, _ string) ([]byte, error) {
		return []byte("version: 1.0.3"), nil
	})

	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	version, err := h.getLatestChartVersion("grafana")
	assert.Empty(t, version)

	assert.Equal(t, `error getting chart URL: error getting gitrepositories: gitrepositories.source.toolkit.fluxcd.io "grafana" not found`, err.Error())
}

func TestGetLatestChartVersionErrorGettingChartFile(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, "big-bang/product/packages/grafana", repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "main", branch)
		return []byte("version: 1.0.3"), errors.New("dummy error")
	})

	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	version, err := h.getLatestChartVersion("grafana")
	assert.Empty(t, version)
	assert.Equal(t, "error getting Chart.yaml: dummy error", err.Error())
}

func TestGetLatestChartVersionErrorDecodingChartFile(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, "big-bang/product/packages/grafana", repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "main", branch)
		return []byte("not yaml"), nil
	})

	v, err := factory.GetViper()
	require.NoError(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	version, err := h.getLatestChartVersion("grafana")
	assert.Empty(t, version)
	assert.Equal(t, "failed to decode Chart.yaml: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `not yaml` into cmd.helmChartManifest", err.Error())
}

func TestUpdateIsNewer(t *testing.T) {
	tests := []struct {
		name           string
		current        string
		latest         string
		expectedResult bool
		expectedError  string
	}{
		{
			name:           "Latest is newer",
			current:        "1.0.0",
			latest:         "1.1.0",
			expectedResult: true,
			expectedError:  "",
		},
		{
			name:           "Current is newer",
			current:        "2.0.0",
			latest:         "1.9.9",
			expectedResult: false,
			expectedError:  "",
		},
		{
			name:           "Versions are equal",
			current:        "1.2.3",
			latest:         "1.2.3",
			expectedResult: false,
			expectedError:  "",
		},
		{
			name:           "Invalid current version",
			current:        "invalid",
			latest:         "1.0.0",
			expectedResult: false,
			expectedError:  `invalid version "invalid":`,
		},
		{
			name:           "Invalid latest version",
			current:        "1.0.0",
			latest:         "invalid",
			expectedResult: false,
			expectedError:  `invalid version "invalid":`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := updateIsNewer(tt.current, tt.latest)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestSplitChartName(t *testing.T) {
	tests := []struct {
		name           string
		fullName       string
		expectedResult string
	}{
		{
			name:           "Standard chart name with version",
			fullName:       "chart-1.2.3-bb.0",
			expectedResult: "chart",
		},
		{
			name:           "Chart name with multiple hyphens",
			fullName:       "my-awesome-chart-2.0.0",
			expectedResult: "my-awesome-chart",
		},
		{
			name:           "Chart name without version",
			fullName:       "simple-chart",
			expectedResult: "simple-chart",
		},
		{
			name:           "Chart name with version starting with zero",
			fullName:       "zero-chart-0.1.0",
			expectedResult: "zero-chart",
		},
		{
			name:           "Empty string",
			fullName:       "",
			expectedResult: "",
		},
		{
			name:           "Chart name with only numbers",
			fullName:       "123-456",
			expectedResult: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitChartName(tt.fullName)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCheckForUpdates(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, bigBangRepo, repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "master", branch)
		return []byte("version: 1.0.2"), nil
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	output, err := h.checkForUpdates("bigbang")
	require.NoError(t, err)

	outputMap := output.ToMap()
	assert.Equal(t, map[string]any{
		"version":         "1.0.2",
		"latestVersion":   "1.0.2",
		"updateAvailable": false,
	}, outputMap)
}

func TestCheckForUpdatesErrorGettingLatestChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(_, _, _ string) ([]byte, error) {
		return nil, errors.New("dummy error")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	outputMap, err := h.checkForUpdates("bigbang")
	assert.Empty(t, outputMap)
	assert.Equal(t, "error checking for latest chart version: error getting Chart.yaml: dummy error", err.Error())
}

func TestCheckForUpdatesErrorCheckingIfUpdateIsNewer(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildHelmReleasesFixture())

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, bigBangRepo, repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "master", branch)
		return []byte("version: invalid"), nil
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	outputMap, err := h.checkForUpdates("bigbang")
	assert.Equal(t, `error checking for update: invalid version "invalid": Invalid Semantic Version`, err.Error())
	assert.Empty(t, outputMap)
}

func TestCheckForUpdatesErrorGettingChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	// Remove grafana from the release fixture
	releases := buildHelmReleasesFixture()
	factory.SetHelmReleases(releases[:0])

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, "big-bang/product/packages/grafana", repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "main", branch)
		return []byte("version: 1.2.3"), nil
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	require.NoError(t, err)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	outputMap, err := h.checkForUpdates("grafana")
	assert.Equal(t, `error getting current chart version: error getting helmreleases: helmreleases.helm.toolkit.fluxcd.io "grafana" not found`, err.Error())
	assert.Empty(t, outputMap)
}

func TestGetSHAsForCurrentPods(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetObjects(
		[]runtime.Object{
			buildGitRepoFixture(),
			buildHelmReleaseFixture(false),
			buildPodFixtures(),
		})

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	versionHelper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	shas, err := versionHelper.getSHAsForCurrentPods("monitoring")
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"grafana:1.0.3": "1234567890",
	}, shas)
}

func TestMatchShas(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetObjects(
		[]runtime.Object{
			buildGitRepoFixture(),
			buildHelmReleaseFixture(false),
			buildPodFixtures(),
		},
	)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	versionHelper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	message, err := versionHelper.matchSHAs("grafana", "1.0.3", "monitoring")
	require.NoError(t, err)

	assert.Equal(t, "All SHAs match", message)
}

func TestMatchSHAsNotFoundInRelease(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetObjects(
		[]runtime.Object{
			buildGitRepoFixture(),
			buildHelmReleaseFixture(false),
			buildPodFixtures(),
		},
	)

	// Return a list that doesn't have any images posted to force a mismatch
	noImages := func(_ int, _, _ string) ([]byte, error) {
		return []byte{}, nil
	}
	factory.SetGitLabGetReleaseArtifactFunc(noImages)

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	versionHelper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	message, err := versionHelper.matchSHAs("grafana", "1.0.3", "monitoring")
	require.NoError(t, err)

	assert.Equal(t, `Error: SHA for running container "grafana:1.0.3" not found in published release artifacts`, message)
}

func TestMatchShasMismatchedSHAs(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("output-config.format", "yaml")

	factory.SetGVRToListKind(versionTestGVR)
	factory.SetHelmReleases(buildHelmReleasesFixture())
	factory.SetObjects(
		[]runtime.Object{
			buildGitRepoFixture(),
			buildHelmReleaseFixture(false),
			buildPodFixtures(),
		},
	)

	factory.SetIronBankGetImageSHAFunc(func(_ string) (string, error) {
		return "mismatched-sha", nil
	})

	cmd, err := NewVersionCmd(factory)
	require.NoError(t, err)

	versionHelper, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	require.NoError(t, err)

	message, err := versionHelper.matchSHAs("grafana", "1.0.3", "monitoring")
	require.NoError(t, err)

	assert.Equal(t, `Error: SHA mismatch for image "grafana:1.0.3". Local: "mismatched-sha", upstream: "1234567890"`, message)
}
