package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	mock "repo1.dso.mil/big-bang/product/packages/bbctl/mocks/repo1.dso.mil/big-bang/product/packages/bbctl/static"
	"repo1.dso.mil/big-bang/product/packages/bbctl/static"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	bbTestApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"

	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8sTesting "k8s.io/client-go/testing"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

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

	// Assert
	assert.NoError(t, res)
	if !assert.Contains(t, buf.String(), "bigbang: 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
	if !assert.Contains(t, buf.String(), "bbctl: ") {
		t.Errorf("unexpected output: %s", buf.String())
	}
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

	// Assert
	assert.NoError(t, err)
	if !assert.NotContains(t, buf.String(), "bigbang: 1.0.2") {
		t.Errorf("unexpected output: %s", buf.String())
	}
	if !assert.Contains(t, buf.String(), "bbctl: ") {
		t.Errorf("unexpected output: %s", buf.String())
	}
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
	assert.Error(t, err)
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
	if assert.Error(t, err) {
		assert.Equal(t, expected, err.Error())
	}
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
	assert.Error(t, res)
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
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	// Act
	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	// Set a config client to avoid this from failing on getting logging client
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: "test",
		}, nil
	}

	logger, err := factory.GetLoggingClient()
	assert.Nil(t, err)

	client, err := bbConfig.NewClient(getConfigFunc, nil, &logger, cmd, v)
	assert.Nil(t, err)

	factory.SetConfigClient(client)

	// Set constants client to fail now
	expectedError := fmt.Errorf("failed to get constants")
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
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	// Act
	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	// Set a config client to avoid this from failing on getting logging client
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: "test",
		}, nil
	}

	logger, err := factory.GetLoggingClient()
	assert.Nil(t, err)

	client, err := bbConfig.NewClient(getConfigFunc, nil, &logger, cmd, v)
	assert.Nil(t, err)

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
	assert.Nil(t, err)

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
	assert.Nil(t, err)

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
	assert.Error(t, err)
	assert.Equal(t, "error creating version helper: failed to get helm client", err.Error())
}

func TestVersionFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd, _ := NewVersionCmd(factory)
	viper, _ := factory.GetViper()
	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, fmt.Errorf("Dummy Error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, h)

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestVersionErrorBindingFlags(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")

	expectedError := fmt.Errorf("failed to set and bind flag")
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
			expectedErrMsg: fmt.Sprintf("error setting and binding client flag: %s", expectedError.Error()),
		},
		{
			flagName:       "all-charts",
			failOnCallNum:  2,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding all-charts flag: %s", expectedError.Error()),
		},
		{
			flagName:       "check-for-updates",
			failOnCallNum:  3,
			expectedCmd:    false,
			expectedErrMsg: fmt.Sprintf("error setting and binding check-for-updates flag: %s", expectedError.Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			callCount := 0
			setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, shortName string, value any, description string) error {
				callCount++
				if callCount == tt.failOnCallNum {
					return expectedError
				}
				return nil
			}

			configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, v)
			assert.Nil(t, err)
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
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			} else {
				assert.Nil(t, err)
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
	assert.Nil(t, err)

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
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, true)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	// Act
	err = h.bbVersion([]string{})

	// Assert
	assert.Error(t, err)
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
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, true)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

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
	v.Set("big-bang-repo", "test")
	v.Set("all-charts", true)
	streams, err := factory.GetIOStream()
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	assert.NoError(t, h.bbVersion([]string{}))

	output := fakeWriter.ActualBuffer.(*bytes.Buffer).String()

	assert.Contains(t, output, "bigbang:1.0.2")
	assert.Contains(t, output, "grafana:1.0.3")

}

func TestBBVersionErrorGettingAllCharts(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("all-charts", true)
	streams, err := factory.GetIOStream()
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &errorListHelmReleasesFunc)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	err = h.bbVersion([]string{})
	assert.NotNil(t, err)
	assert.Equal(t, "error getting all chart versions: error getting helmreleases: error in list crds", err.Error())
}

func TestBBVersionNoArgsErrorCheckingForUpdates(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("check-for-updates", true)
	streams, err := factory.GetIOStream()
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		return nil, fmt.Errorf("dummy error")
	})

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	err = h.bbVersion([]string{})

	assert.NotNil(t, err)
	assert.Equal(t, "error checking for updates: error checking for latest chart version: error getting Chart.yaml: dummy error", err.Error())
}

func TestBBVersionWithChartName(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	streams, err := factory.GetIOStream()
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	err = h.bbVersion([]string{"grafana"})
	assert.Nil(t, err)

	output := fakeWriter.ActualBuffer.(*bytes.Buffer).String()
	assert.Contains(t, output, "grafana:1.0.3")

}

func TestBBVersionWithChartNameErrorGettingChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	streams, err := factory.GetIOStream()
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetGVRToListKind(fluxResourceGVR)

	// Create a release with no version set
	releases := &unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": "helm.toolkit.fluxcd.io/v2",
			"kind":       "HelmRelease",
		},
		Items: []unstructured.Unstructured{newHelmRelease("grafana", "bigbang", "")},
	}

	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), releases})
	factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &errorListHelmReleasesFunc)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	err = h.bbVersion([]string{"grafana"})
	assert.Equal(t, `error getting chart version: error getting version for release "grafana": no version specified for the chart`, err.Error())
}

func TestBBVersionWithChartNameErrorCheckingForUpdates(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	v.Set("check-for-updates", true)
	streams, err := factory.GetIOStream()
	assert.Nil(t, err)
	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		return nil, fmt.Errorf("dummy error")
	})

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	err = h.bbVersion([]string{"grafana"})
	assert.Equal(t, `error checking for updates: error checking for latest chart version: error getting Chart.yaml: dummy error`, err.Error())
}

func TestBBVersionWithInvalidNumberOfArguments(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	v, _ := factory.GetViper()
	v.Set("big-bang-repo", "test")
	streams, err := factory.GetIOStream()
	assert.Nil(t, err)

	fakeWriter := bbTestApiWrappers.CreateFakeReaderWriter(t, false, false)
	streams.Out = fakeWriter
	factory.SetIOStream(streams)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

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
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	version, err := h.getReleaseVersion("grafana")
	assert.Empty(t, version)
	assert.Equal(t, `error getting version for release "grafana": no version specified for the chart`, err.Error())
}

func buildReleaseFixture() []*release.Release {
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

var fluxResourceGVR = map[runtimeSchema.GroupVersionResource]string{
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
}

var errorListHelmReleasesFunc = func(client *dynamicFake.FakeDynamicClient) {
	client.Fake.PrependReactor("list", "helmreleases", func(action k8sTesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("error in list crds")
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

func newHelmRelease(name, namespace, chartVersion string) unstructured.Unstructured {
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
		},
	}
}

func buildHelmReleaseFixture(includeBigBang bool) *unstructured.UnstructuredList {
	items := []unstructured.Unstructured{
		newHelmRelease("grafana", "bigbang", "1.0.3"),
		newHelmRelease("tempo", "bigbang", "3.2.1"),
	}

	if includeBigBang {
		items = append(items, newHelmRelease("bigbang", "bigbang", "1.0.2"))
	}

	return &unstructured.UnstructuredList{
		Object: map[string]any{
			"apiVersion": "helm.toolkit.fluxcd.io/v2",
			"kind":       "HelmRelease",
		},
		Items: items,
	}
}

func TestGetAllChartVersions(t *testing.T) {

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(false)
	assert.Nil(t, err)

	assert.Equal(t, outputMap["bigbang"], "1.0.2")
	assert.Equal(t, outputMap["grafana"], "1.0.3")
	assert.Equal(t, outputMap["tempo"], "3.2.1")
}

func TestGetAllChartVersionsBigBangManagedByFlux(t *testing.T) {

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(true)})
	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(false)
	assert.Nil(t, err)

	assert.Equal(t, outputMap["bigbang"], "1.0.2")
	assert.Equal(t, outputMap["grafana"], "1.0.3")
	assert.Equal(t, outputMap["tempo"], "3.2.1")
}

func TestGetAllChartVersionsErrorGettingBigBangVersion(t *testing.T) {

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(false)
	assert.Equal(t, "error getting Big Bang version: error fetching Big Bang release version: error getting helm information for release bigbang: release bigbang not found", err.Error())

	assert.Equal(t, outputMap["grafana"], "1.0.3")
	assert.Equal(t, outputMap["tempo"], "3.2.1")
}

func TestGetAllChartVersionsErrorListingReleases(t *testing.T) {
	releaseFixture := buildReleaseFixture()

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(releaseFixture)

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &errorListHelmReleasesFunc)

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(false)
	assert.Equal(t, "error getting helmreleases: error in list crds", err.Error())
	assert.Empty(t, len(outputMap))
}

func TestGetAllChartVersionsCheckForUpdates(t *testing.T) {

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)

	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		switch repository {
		case "big-bang/bigbang":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "master")
			return []byte("version: 1.0.2"), nil
		case "big-bang/product/packages/grafana":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "main")
			return []byte("version: 2.0.0"), nil
		case "big-bang/product/packages/tempo":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "main")
			return []byte("version: 3.2.2"), nil
		}

		return nil, fmt.Errorf("invalid repository")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(true)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"grafana": map[string]any{
			"version":         "1.0.3",
			"latest":          "2.0.0",
			"updateAvailable": true,
		},
		"bigbang": map[string]any{
			"version":         "1.0.2",
			"latest":          "1.0.2",
			"updateAvailable": false,
		},
		"tempo": map[string]any{
			"version":         "3.2.1",
			"latest":          "3.2.2",
			"updateAvailable": true,
		},
	}, outputMap)

}

func TestGetAllChartVersionsCheckForUpdatesErrorGettingLatestChartVersion(t *testing.T) {

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		return nil, fmt.Errorf("dummy error")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(true)
	assert.Equal(t, "error getting latest chart version: error getting Chart.yaml: dummy error", err.Error())
	assert.Empty(t, outputMap)

}

func TestGetAllChartVersionsCheckForUpdatesErrorCheckingIfBigBangUpdateIsNewer(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		switch repository {
		case "big-bang/bigbang":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "master")
			return []byte("version: invalid"), nil
		case "big-bang/product/packages/grafana":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "main")
			return []byte("version: 2.0.0"), nil
		case "big-bang/product/packages/tempo":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "main")
			return []byte("version: 3.2.2"), nil
		}
		return nil, fmt.Errorf("invalid repository")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(true)
	assert.Equal(t, `error checking for update: invalid version "invalid": Invalid Semantic Version`, err.Error())
	assert.Equal(t, map[string]any{
		"latest":          "2.0.0",
		"updateAvailable": true,
		"version":         "1.0.3",
	}, outputMap["grafana"])
	assert.Empty(t, outputMap["bigbang"])

}

func TestGetAllChartVersionsCheckForUpdatesErrorCheckingIfPackageUpdateIsNewer(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		switch repository {
		case "big-bang/bigbang":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "master")
			return []byte("version: 1.0.2"), nil
		case "big-bang/product/packages/grafana":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "main")
			return []byte("version: 2.0.0"), nil
		case "big-bang/product/packages/tempo":
			assert.Equal(t, path, "chart/Chart.yaml")
			assert.Equal(t, branch, "main")
			return []byte("version: invalid"), nil
		}
		return nil, fmt.Errorf("invalid repository")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.getAllChartVersions(true)
	assert.Equal(t, `error checking for update: invalid version "invalid": Invalid Semantic Version`, err.Error())
	assert.Equal(t, map[string]any{
		"latest":          "2.0.0",
		"updateAvailable": true,
		"version":         "1.0.3",
	}, outputMap["grafana"])
	assert.Empty(t, outputMap["tempo"])

}

func TestGetChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())
	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	version, err := h.getChartVersion("bigbang")
	assert.Nil(t, err)
	assert.Equal(t, version, "1.0.2")

	version, err = h.getChartVersion("invalid-chart")
	assert.Empty(t, version)
	assert.Equal(t, `error getting helmreleases: helmreleases.helm.toolkit.fluxcd.io "invalid-chart" not found`, err.Error())
}

// func TestGetChartVersionErrorListingReleases(t *testing.T) {
// 	factory := bbTestUtil.GetFakeFactory()
// 	factory.ResetIOStream()
// 	factory.SetHelmReleases(buildReleaseFixture())
// 	factory.SetGVRToListKind(fluxResourceGVR)
// 	factory.SetObjects([]runtime.Object{buildGitRepoFixture(), buildHelmReleaseFixture(false)})
// 	factory.SetFail.GetK8sDynamicClientPrepFuncs = append(factory.SetFail.GetK8sDynamicClientPrepFuncs, &errorListHelmReleasesFunc)
//
// 	v, err := factory.GetViper()
// 	v.Set("big-bang-repo", "test")
// 	assert.Nil(t, err)
//
// 	cmd, err := NewVersionCmd(factory)
// 	assert.Nil(t, err)
//
// 	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
// 	assert.Nil(t, err)
//
// 	outputMap, err := h.getChartVersion("bigbang")
// 	fmt.Println(outputMap)
// 	assert.Equal(t, "error getting helm information for all releases: dummy error", err.Error())
// 	assert.Empty(t, len(outputMap))
// }

func TestGetLatestChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, repository, "big-bang/product/packages/grafana")
		assert.Equal(t, path, "chart/Chart.yaml")
		assert.Equal(t, branch, "main")
		return []byte("version: 1.0.3"), nil
	})

	v, err := factory.GetViper()
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	version, err := h.getLatestChartVersion("grafana")
	assert.Nil(t, err)
	assert.Equal(t, "1.0.3", version)
}

// TestGetLatestChartVersionBigBang tests that the special internal conditions required to check the version of Big Bang are met
func TestGetLatestChartVersionBigBang(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, "big-bang/bigbang", repository)
		assert.Equal(t, "chart/Chart.yaml", path)
		assert.Equal(t, "master", branch)
		return []byte("version: 1.0.3"), nil
	})

	v, err := factory.GetViper()
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	version, err := h.getLatestChartVersion("bigbang")
	assert.Nil(t, err)
	assert.Equal(t, version, "1.0.3")
}

func TestGetLatestChartVersionErrorListingGitRepos(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects(nil)

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		return []byte("version: 1.0.3"), nil
	})

	v, err := factory.GetViper()
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	version, err := h.getLatestChartVersion("grafana")
	assert.Empty(t, version)
	assert.Equal(t, "error getting gitrepositories: gitrepositories.source.toolkit.fluxcd.io \"grafana\" not found", err.Error())
}

func TestGetLatestChartVersionErrorGettingChartFile(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, repository, "big-bang/product/packages/grafana")
		assert.Equal(t, path, "chart/Chart.yaml")
		assert.Equal(t, branch, "main")
		return []byte("version: 1.0.3"), fmt.Errorf("dummy error")
	})

	v, err := factory.GetViper()
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	version, err := h.getLatestChartVersion("grafana")
	assert.Empty(t, version)
	assert.Equal(t, "error getting Chart.yaml: dummy error", err.Error())
}

func TestGetLatestChartVersionErrorDecodingChartFile(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, repository, "big-bang/product/packages/grafana")
		assert.Equal(t, path, "chart/Chart.yaml")
		assert.Equal(t, branch, "main")
		return []byte("not yaml"), nil
	})

	v, err := factory.GetViper()
	assert.Nil(t, err)
	v.Set("big-bang-repo", "test")

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

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
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
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
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, "big-bang/bigbang", repository)
		assert.Equal(t, path, "chart/Chart.yaml")
		assert.Equal(t, branch, "master")
		return []byte("version: 1.0.2"), nil
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.checkForUpdates("bigbang")
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"version":         "1.0.2",
		"latest":          "1.0.2",
		"updateAvailable": false,
	}, outputMap)

}

func TestCheckForUpdatesErrorGettingLatestChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		return nil, fmt.Errorf("dummy error")
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.checkForUpdates("bigbang")
	assert.Empty(t, outputMap)
	assert.Equal(t, "error checking for latest chart version: error getting Chart.yaml: dummy error", err.Error())
}

func TestCheckForUpdatesErrorCheckingIfUpdateIsNewer(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	factory.SetHelmReleases(buildReleaseFixture())

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, "big-bang/bigbang", repository)
		assert.Equal(t, path, "chart/Chart.yaml")
		assert.Equal(t, branch, "master")
		return []byte("version: invalid"), nil
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.checkForUpdates("bigbang")
	assert.Equal(t, `error checking for update: invalid version "invalid": Invalid Semantic Version`, err.Error())
	assert.Empty(t, outputMap)

}

func TestCheckForUpdatesErrorGettingChartVersion(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()

	// Remove grafana from the release fixture
	releases := buildReleaseFixture()
	factory.SetHelmReleases(releases[:0])

	factory.SetGVRToListKind(fluxResourceGVR)
	factory.SetObjects([]runtime.Object{buildGitRepoFixture()})

	factory.SetGitLabGetFileFunc(func(repository string, path string, branch string) ([]byte, error) {
		assert.Equal(t, repository, "big-bang/product/packages/grafana")
		assert.Equal(t, path, "chart/Chart.yaml")
		assert.Equal(t, branch, "main")
		return []byte("version: 1.2.3"), nil
	})

	v, err := factory.GetViper()
	v.Set("big-bang-repo", "test")
	assert.Nil(t, err)

	cmd, err := NewVersionCmd(factory)
	assert.Nil(t, err)

	h, err := newVersionCmdHelper(cmd, factory, static.DefaultClient)
	assert.Nil(t, err)

	outputMap, err := h.checkForUpdates("grafana")
	assert.Equal(t, `error getting current chart version: error getting helmreleases: helmreleases.helm.toolkit.fluxcd.io "grafana" not found`, err.Error())
	assert.Empty(t, outputMap)

}
