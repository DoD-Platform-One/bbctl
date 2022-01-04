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

// GetFakeFactory - get fake factory
func GetFakeFactory(helmReleases []*release.Release, objs []runtime.Object, gvrToListKind map[schema.GroupVersionResource]string) *FakeFactory {
	return &FakeFactory{helmReleases: helmReleases, objs: objs, gvrToListKind: gvrToListKind}
}

// FakeFactory - fake factory
type FakeFactory struct {
	helmReleases  []*release.Release
	objs          []runtime.Object
	gvrToListKind map[schema.GroupVersionResource]string
}

// GetHelmClient - get helm client
func (f *FakeFactory) GetHelmClient(namespace string) (helm.Client, error) {
	return NewFakeClient(f.helmReleases)
}

// GetClientSet - get clientset
func (f *FakeFactory) GetClientSet() (kubernetes.Interface, error) {
	fakeClient := fake.NewSimpleClientset()
	return fakeClient, nil
}

// GetRuntimeClient - get runtime client
func (f *FakeFactory) GetRuntimeClient(scheme *runtime.Scheme) (client.Client, error) {

	cb := fakectlrclient.NewClientBuilder()
	rc := cb.WithScheme(scheme).Build()
	return rc, nil
}

// GetK8sClientset - get k8s clientset
func (f *FakeFactory) GetK8sClientset() (kubernetes.Interface, error) {
	return fake.NewSimpleClientset(f.objs...), nil
}

// GetK8sDynamicClient - get k8s dynamic client
func (f *FakeFactory) GetK8sDynamicClient() (dynamic.Interface, error) {
	scheme := runtime.NewScheme()
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, f.gvrToListKind, f.objs...), nil
}
