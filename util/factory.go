package util

import (
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/release"
	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/helm"
	bbk8sutil "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/k8s"
)

// Factory creates utility objects
type Factory interface {
	GetHelmClient(namespace string) (helm.Client, error)
}

func NewFactory(flags *pflag.FlagSet) *utilFactory {
	return &utilFactory{flags: flags}
}

type utilFactory struct {
	flags *pflag.FlagSet
}

func (f *utilFactory) GetHelmClient(namespace string) (helm.Client, error) {
	config, err := bbk8sutil.BuildKubeConfigFromFlags(f.flags)
	if err != nil {
		return nil, err
	}

	opt := &helm.Options{
		Namespace:        namespace,
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            true,
		Linting:          true,
		RestConfig:       config,
	}

	return helm.New(opt)
}

func FakeFactory(helmReleases []*release.Release) *fakeFactory {
	return &fakeFactory{helmReleases: helmReleases}
}

type fakeFactory struct {
	helmReleases []*release.Release
}

func (f *fakeFactory) GetHelmClient(namespace string) (helm.Client, error) {
	return helm.NewFakeClient(f.helmReleases)
}
