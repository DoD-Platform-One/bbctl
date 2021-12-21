package helm

import (
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

	return &ClientHelm{
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

// GetRelease - GetRelease returns a release specified by name.
func (c *ClientHelm) GetRelease(name string) (*release.Release, error) {
	return c.getRelease(name)
}

// getRelease returns a release matching the provided 'name'.
func (c *ClientHelm) getRelease(name string) (*release.Release, error) {
	getReleaseClient := action.NewGet(c.ActionConfig)

	return getReleaseClient.Run(name)
}

// GetList - getList returns a list of releases
func (c *ClientHelm) GetList() ([]*release.Release, error) {
	return c.getList()
}

// getList returns a list of releases
func (c *ClientHelm) getList() ([]*release.Release, error) {
	getListClient := action.NewList(c.ActionConfig)

	return getListClient.Run()
}

// GetValues - getValues returns release values
func (c *ClientHelm) GetValues(name string, allValues bool) (interface{}, error) {
	return c.getValues(name, allValues)
}

// getValues returns release values
func (c *ClientHelm) getValues(name string, allValues bool) (map[string]interface{}, error) {
	getValuesClient := action.NewGetValues(c.ActionConfig)
	getValuesClient.AllValues = allValues
	return getValuesClient.Run(name)
}
