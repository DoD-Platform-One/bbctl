package util

import (
	"io"
	"log"
	"os"

	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"

	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbk8sutil "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Factory interface
type Factory interface {
	GetHelmClient(namespace string) (helm.Client, error)
	GetK8sClientset() (kubernetes.Interface, error)
	GetRuntimeClient(*runtime.Scheme) (runtimeclient.Client, error)
	GetK8sDynamicClient() (dynamic.Interface, error)
	GetRestConfig() (*rest.Config, error)
	GetCommandExecutor(pod *corev1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remotecommand.Executor, error)
}

// NewFactory - new factory method
func NewFactory(flags *pflag.FlagSet) *UtilityFactory {
	return &UtilityFactory{flags: flags}
}

// UtilityFactory - util factory
type UtilityFactory struct {
	flags *pflag.FlagSet
}

// GetHelmClient - get helm client
func (f *UtilityFactory) GetHelmClient(namespace string) (helm.Client, error) {

	actionConfig, err := f.getHelmConfig(namespace)
	if err != nil {
		return nil, err
	}

	getReleaseClient := action.NewGet(actionConfig)

	getListClient := action.NewList(actionConfig)

	getValuesClient := action.NewGetValues(actionConfig)
	getValuesClient.AllValues = true

	return helm.NewClient(getReleaseClient.Run, getListClient.Run, getValuesClient.Run)
}

// GetK8sClientset - get k8s clientset
func (f *UtilityFactory) GetK8sClientset() (kubernetes.Interface, error) {

	config, err := f.GetRestConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// GetK8sDynamicClient - get k8s dynamic client
func (f *UtilityFactory) GetK8sDynamicClient() (dynamic.Interface, error) {
	return bbk8sutil.BuildDynamicClientFromFlags(f.flags)
}

// GetRuntimeClient - get runtime client
func (f *UtilityFactory) GetRuntimeClient(scheme *runtime.Scheme) (runtimeclient.Client, error) {

	// init runtime cotroller client
	runtimeClient, err := runtimeclient.New(ctrl.GetConfigOrDie(), runtimeclient.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return runtimeClient, err
}

// GetRestConfig - get rest config
func (f *UtilityFactory) GetRestConfig() (*rest.Config, error) {
	return bbk8sutil.BuildKubeConfigFromFlags(f.flags)
}

// GetCommandExecutor - get executor to run command in a Pod
func (f *UtilityFactory) GetCommandExecutor(pod *corev1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remotecommand.Executor, error) {

	client, err := f.GetK8sClientset()
	if err != nil {
		return nil, err
	}

	req := client.Discovery().RESTClient().Post().
		Prefix("/api/v1").
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")

	req.SpecificallyVersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     false,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       false,
	}, scheme.ParameterCodec, schema.GroupVersion{Version: "v1"})

	config, err := f.GetRestConfig()
	if err != nil {
		return nil, err
	}

	return remotecommand.NewSPDYExecutor(config, "POST", req.URL())
}

func (f *UtilityFactory) getHelmConfig(namespace string) (*action.Configuration, error) {

	config, err := bbk8sutil.BuildKubeConfigFromFlags(f.flags)
	if err != nil {
		return nil, err
	}

	// TODO: add support for an alternate warning handler and then just default nil
	clientGetter := helm.NewRESTClientGetter(config, namespace, nil)

	debugLog := func(format string, v ...interface{}) {
		log.Printf(format, v...)
	}

	actionConfig := new(action.Configuration)
	err = actionConfig.Init(
		clientGetter,
		namespace,
		os.Getenv("HELM_DRIVER"),
		debugLog,
	)
	if err != nil {
		return nil, err
	}

	return actionConfig, nil
}
