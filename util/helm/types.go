package helm

import (
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

// Options defines the options for constructing a client via REST config
type Options struct {
	Namespace        string
	RepositoryConfig string
	RepositoryCache  string
	Debug            bool
	Linting          bool
	DebugLog         action.DebugLog
	RestConfig       *rest.Config
}

// RESTClientGetter defines the values of a helm REST client
type RESTClientGetter struct {
	namespace  string
	kubeConfig []byte
	restConfig *rest.Config
}

// HelmClient Client defines the values of a helm client
type HelmClient struct {
	Settings     *cli.EnvSettings
	Providers    getter.Providers
	storage      *repo.File
	ActionConfig *action.Configuration
	linting      bool
	DebugLog     action.DebugLog
}
