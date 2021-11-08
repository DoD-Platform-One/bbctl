package helm

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Client holds the method signatures for a Helm client.
type Client interface {
	GetRelease(name string) (*release.Release, error)
	GetList() ([]*release.Release, error)
	GetValues(name string, allValues bool) (interface{}, error)
}

var storage = repo.File{}

const (
	defaultCachePath            = "/tmp/.helmcache"
	defaultRepositoryConfigPath = "/tmp/.helmrepo"
)

// New returns a new Helm client with the provided options
func New(options *Options) (Client, error) {
	settings := cli.New()

	err := setEnvSettings(options, settings)
	if err != nil {
		return nil, err
	}

	clientGetter := NewRESTClientGetter(options.Namespace, nil, options.RestConfig)

	return newClient(options, clientGetter, settings)
}

// newClient returns a new Helm client via the provided options
func newClient(options *Options, clientGetter genericclioptions.RESTClientGetter, settings *cli.EnvSettings) (Client, error) {
	debugLog := options.DebugLog
	if debugLog == nil {
		debugLog = func(format string, v ...interface{}) {
			log.Printf(format, v...)
		}
	}

	actionConfig := new(action.Configuration)
	err := actionConfig.Init(
		clientGetter,
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)
	if err != nil {
		return nil, err
	}

	return &HelmClient{
		Settings:     settings,
		Providers:    getter.All(settings),
		storage:      &storage,
		ActionConfig: actionConfig,
		linting:      options.Linting,
		DebugLog:     debugLog,
	}, nil
}

// setEnvSettings sets the client's environment settings based on the provided client configuration
func setEnvSettings(options *Options, settings *cli.EnvSettings) error {
	if options == nil {
		options = &Options{
			RepositoryConfig: defaultRepositoryConfigPath,
			RepositoryCache:  defaultCachePath,
			Linting:          true,
		}
	}

	if options.Namespace != "" {
		pflags := pflag.NewFlagSet("", pflag.ContinueOnError)
		settings.AddFlags(pflags)
		err := pflags.Parse([]string{"-n", options.Namespace})
		if err != nil {
			return err
		}
	}

	if options.RepositoryConfig == "" {
		options.RepositoryConfig = defaultRepositoryConfigPath
	}

	if options.RepositoryCache == "" {
		options.RepositoryCache = defaultCachePath
	}

	settings.RepositoryCache = options.RepositoryCache
	settings.RepositoryConfig = defaultRepositoryConfigPath
	settings.Debug = options.Debug

	return nil
}

// GetRelease returns a release specified by name.
func (c *HelmClient) GetRelease(name string) (*release.Release, error) {
	return c.getRelease(name)
}

// getRelease returns a release matching the provided 'name'.
func (c *HelmClient) getRelease(name string) (*release.Release, error) {
	getReleaseClient := action.NewGet(c.ActionConfig)

	return getReleaseClient.Run(name)
}

// getList returns a list of releases
func (c *HelmClient) GetList() ([]*release.Release, error) {
	return c.getList()
}

// getList returns a list of releases
func (c *HelmClient) getList() ([]*release.Release, error) {
	getListClient := action.NewList(c.ActionConfig)

	return getListClient.Run()
}

// getValues returns release values
func (c *HelmClient) GetValues(name string, allValues bool) (interface{}, error) {
	return c.getValues(name, allValues)
}

// getValues returns release values
func (c *HelmClient) getValues(name string, allValues bool) (map[string]interface{}, error) {
	getValuesClient := action.NewGetValues(c.ActionConfig)
	getValuesClient.AllValues = allValues
	return getValuesClient.Run(name)
}

// NewFakeClient returns a new Fale Helm client with the provided options
func NewFakeClient(releases []*release.Release) (Client, error) {
	return &FakeClient{releases: releases}, nil
}

type FakeClient struct {
	releases []*release.Release
}

// GetRelease returns a release specified by name.
func (c *FakeClient) GetRelease(name string) (*release.Release, error) {
	for _, r := range c.releases {
		if r.Name == name {
			return r, nil
		}
	}

	return nil, fmt.Errorf("release %s not found", name)
}

// getList returns a list of releases
func (c *FakeClient) GetList() ([]*release.Release, error) {
	return c.releases, nil
}

// getList returns a list of releases
func (c *FakeClient) GetValues(name string, allValues bool) (interface{}, error) {
	for _, r := range c.releases {
		if r.Name == name {
			return r.Chart.Values, nil
		}
	}

	return nil, fmt.Errorf("release %s not found", name)
}
