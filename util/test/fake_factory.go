package test

import (
	"context"
	"io"
	"strings"

	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	fakehelm "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/helm"

	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakectlrclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// GetFakeFactory - get fake factory
func GetFakeFactory(helmReleases []*release.Release, objs []runtime.Object,
	gvrToListKind map[schema.GroupVersionResource]string, resources []*metav1.APIResourceList) *FakeFactory {
	return &FakeFactory{
		helmReleases:  helmReleases,
		objs:          objs,
		gvrToListKind: gvrToListKind,
		resources:     resources,
	}
}

// FakeFactory - fake factory
type FakeFactory struct {
	helmReleases  []*release.Release
	objs          []runtime.Object
	gvrToListKind map[schema.GroupVersionResource]string
	resources     []*metav1.APIResourceList
}

// GetHelmClient - get helm client
func (f *FakeFactory) GetHelmClient(namespace string) (helm.Client, error) {
	return fakehelm.NewFakeClient(f.helmReleases)
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
	cs := fake.NewSimpleClientset(f.objs...)
	if f.resources != nil {
		cs.Fake.Resources = f.resources
	}
	return cs, nil
}

// GetK8sDynamicClient - get k8s dynamic client
func (f *FakeFactory) GetK8sDynamicClient() (dynamic.Interface, error) {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, f.gvrToListKind, f.objs...), nil
}

// GetRestConfig - get rest config
func (f *FakeFactory) GetRestConfig() (*rest.Config, error) {
	return &rest.Config{}, nil
}

// GetCommandExecutor - execute command in a Pod
func (f *FakeFactory) GetCommandExecutor(pod *corev1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remotecommand.Executor, error) {
	fakeCommandExecutor.Command = strings.Join(command, " ")
	return fakeCommandExecutor, nil
}

// GetFakeCommandExecutor - get fake command executor
func GetFakeCommandExecutor() *FakeCommandExecutor {
	return fakeCommandExecutor
}

// FakeCommandExecutor - fake command executor
type FakeCommandExecutor struct {
	Command       string
	CommandResult map[string]string
}

// Stream - stream command result
func (f *FakeCommandExecutor) Stream(options remotecommand.StreamOptions) error {
	stdout := options.Stdout
	output := f.CommandResult[f.Command]
	stdout.Write([]byte(output))
	return nil
}

// StreamWithContext - stream command result with given context
func (f *FakeCommandExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	stdout := options.Stdout
	output := f.CommandResult[f.Command]
	stdout.Write([]byte(output))
	return nil
}

var fakeCommandExecutor = &FakeCommandExecutor{}
