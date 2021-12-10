package util_test

import (
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "k8s.io/client-go/kubernetes"
	fakeclientgo "k8s.io/client-go/kubernetes/fake"
	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakectlrclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type FakeFactory struct {
	HelmReleases []*release.Release
}

func (f *FakeFactory) GetHelmClient(namespace string) (helm.Client, error) {
	return helm.NewFakeClient(f.HelmReleases)
}

func (f *FakeFactory) GetClientSet() (fakeclient.Interface, error) {
	fakeClient := fakeclientgo.NewSimpleClientset()
	return fakeClient, nil
}

func (f *FakeFactory) GetRuntimeClient(scheme *runtime.Scheme) (client.Client, error) {

	cb := fakectlrclient.NewClientBuilder()
	rc := cb.WithScheme(scheme).Build()
	return rc, nil
}