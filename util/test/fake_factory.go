package test

import (
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	helm "repo1.dso.mil/platform-one/big-bang/apps/product-tools/bbctl/util/helm"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakectlrclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func FakeFactory(helmReleases []*release.Release, objs []runtime.Object, gvrToListKind map[schema.GroupVersionResource]string) *fakeFactory {
	return &fakeFactory{helmReleases: helmReleases, objs: objs, gvrToListKind: gvrToListKind}
}

type fakeFactory struct {
	helmReleases  []*release.Release
	objs          []runtime.Object
	gvrToListKind map[schema.GroupVersionResource]string
}

func (f *fakeFactory) GetHelmClient(namespace string) (helm.Client, error) {
	return helm.NewFakeClient(f.helmReleases)
}

func (f *fakeFactory) GetClientSet() (kubernetes.Interface, error) {
	fakeClient := fake.NewSimpleClientset()
	return fakeClient, nil
}

func (f *fakeFactory) GetRuntimeClient(scheme *runtime.Scheme) (client.Client, error) {

	cb := fakectlrclient.NewClientBuilder()
	rc := cb.WithScheme(scheme).Build()
	return rc, nil
}

func (f *fakeFactory) GetK8sClientset() (kubernetes.Interface, error) {
	return fake.NewSimpleClientset(f.objs...), nil
}

func (f *fakeFactory) GetK8sDynamicClient() (dynamic.Interface, error) {
	scheme := runtime.NewScheme()
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, f.gvrToListKind, f.objs...), nil
}
