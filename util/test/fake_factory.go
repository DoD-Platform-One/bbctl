package test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	bbUtilApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/apiwrappers"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	fakeApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
	fakeAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/aws"
	fakeHelm "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/helm"
	fakeLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/release"
	apisV1Beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	remoteCommand "k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeControllerClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// GetEmptyFakeFactory - get empty fake factory
func GetFakeFactory() *FakeFactory {
	factory := &FakeFactory{}
	// Set required default values
	factory.SetAWSConfig(nil)
	factory.SetClusterIPs(nil)
	factory.SetLoggingFunc(nil)
	factory.viperInstance = viper.New()
	return factory
}

// Set functions

// SetHelmReleases - set helm releases
func (f *FakeFactory) SetHelmReleases(helmReleases []*release.Release) {
	f.helmReleases = helmReleases
}

// SetObjects - set objects
func (f *FakeFactory) SetObjects(objects []runtime.Object) {
	f.objects = objects
}

// SetGVRToListKind - set gvr to list kind
func (f *FakeFactory) SetGVRToListKind(gvrToListKind map[schema.GroupVersionResource]string) {
	f.gvrToListKind = gvrToListKind
}

// SetResources - set resources
func (f *FakeFactory) SetResources(resources []*metaV1.APIResourceList) {
	f.resources = resources
}

// SetAWSConfig - set aws config
func (f *FakeFactory) SetAWSConfig(awsConfig *aws.Config) {
	var awsConfigToUse aws.Config
	if awsConfig == nil {
		awsConfigToUse = aws.Config{}
	} else {
		awsConfigToUse = *awsConfig
	}
	f.awsConfig = awsConfigToUse
}

// SetCallerIdentity - set caller identity
func (f *FakeFactory) SetCallerIdentity(callerIdentity *bbAws.CallerIdentity) {
	f.callerIdentity = callerIdentity
}

// SetClusterIPs - set cluster IPs
func (f *FakeFactory) SetClusterIPs(clusterIPs *[]bbAws.ClusterIP) {
	var clusterIPsToUse []bbAws.ClusterIP
	if clusterIPs == nil {
		clusterIPsToUse = []bbAws.ClusterIP{}
	} else {
		clusterIPsToUse = *clusterIPs
	}
	f.clusterIPs = clusterIPsToUse
}

// SetEC2Client - set ec2 client
func (f *FakeFactory) SetEC2Client(ec2Client *ec2.Client) {
	f.ec2Client = ec2Client
}

// SetLoggingFunc - set logging function
func (f *FakeFactory) SetLoggingFunc(loggingFunc fakeLog.LoggingFunction) {
	var loggingFuncToUse fakeLog.LoggingFunction
	if loggingFunc == nil {
		loggingFuncToUse = func(args ...string) {
			fmt.Println(args)
		}
	} else {
		loggingFuncToUse = loggingFunc
	}
	f.loggingFunc = loggingFuncToUse
}

// SetSTSClient - set sts client
func (f *FakeFactory) SetSTSClient(stsClient *sts.Client) {
	f.stsClient = stsClient
}

// SetVirtualServices - set virtual services
func (f *FakeFactory) SetVirtualServices(virtualServices *apisV1Beta1.VirtualServiceList) {
	f.virtualServiceList = virtualServices
}

// FakeFactory - fake factory
type FakeFactory struct {
	awsConfig          aws.Config
	callerIdentity     *bbAws.CallerIdentity
	clusterIPs         []bbAws.ClusterIP
	ec2Client          *ec2.Client
	helmReleases       []*release.Release
	loggingFunc        fakeLog.LoggingFunction
	objects            []runtime.Object
	gvrToListKind      map[schema.GroupVersionResource]string
	resources          []*metaV1.APIResourceList
	stsClient          *sts.Client
	virtualServiceList *apisV1Beta1.VirtualServiceList
	viperInstance      *viper.Viper
	configClient       *bbConfig.ConfigClient

	SetFail struct {
		GetConfigClient bool
	}
}

// GetCredentialHelper - get credential helper
func (f *FakeFactory) GetCredentialHelper() func(string, string) string {
	return func(arg1 string, arg2 string) string {
		return ""
	}
}

// GetAWSClient - get aws client
func (f *FakeFactory) GetAWSClient() bbAws.Client {
	fakeClient, err := fakeAws.NewFakeClient(f.clusterIPs, &f.awsConfig, f.ec2Client, f.callerIdentity, f.stsClient)
	if err != nil {
		panic(err)
	}
	return fakeClient
}

// GetHelmClient - get helm client
func (f *FakeFactory) GetHelmClient(cmd *cobra.Command, namespace string) (helm.Client, error) {
	return fakeHelm.NewFakeClient(f.helmReleases)
}

// GetClientSet - get clientset
func (f *FakeFactory) GetClientSet() (kubernetes.Interface, error) {
	fakeClient := fake.NewSimpleClientset()
	return fakeClient, nil
}

// GetK8sClientset - get k8s clientset
func (f *FakeFactory) GetK8sClientset(cmd *cobra.Command) (kubernetes.Interface, error) {
	cs := fake.NewSimpleClientset(f.objects...)
	if f.resources != nil {
		cs.Fake.Resources = f.resources
	}
	return cs, nil
}

// GetK8sDynamicClient - get k8s dynamic client
func (f *FakeFactory) GetK8sDynamicClient(cmd *cobra.Command) (dynamic.Interface, error) {
	scheme := runtime.NewScheme()
	err := coreV1.AddToScheme(scheme)
	f.GetLoggingClient().HandleError("failed to add coreV1 to scheme", err)
	return dynamicFake.NewSimpleDynamicClientWithCustomListKinds(scheme, f.gvrToListKind, f.objects...), nil
}

// GetLoggingClient - get logging client
func (f *FakeFactory) GetLoggingClient() bbLog.Client {
	return f.GetLoggingClientWithLogger(nil)
}

// GetLoggingClientWithLogger - get logging client providing logger
func (f *FakeFactory) GetLoggingClientWithLogger(logger *slog.Logger) bbLog.Client {
	var localFunc fakeLog.LoggingFunction
	if f.loggingFunc == nil {
		localFunc = func(args ...string) {
			fmt.Println(args)
		}
	} else {
		localFunc = f.loggingFunc
	}

	client := fakeLog.NewFakeClient(localFunc)
	return client
}

// GetRestConfig - get rest config
func (f *FakeFactory) GetRestConfig(cmd *cobra.Command) (*rest.Config, error) {
	return &rest.Config{}, nil
}

// GetRuntimeClient - get runtime client
func (f *FakeFactory) GetRuntimeClient(scheme *runtime.Scheme) (client.Client, error) {

	cb := fakeControllerClient.NewClientBuilder()
	rc := cb.WithScheme(scheme).Build()
	return rc, nil
}

// GetCommandExecutor - execute command in a Pod
func (f *FakeFactory) GetCommandExecutor(cmd *cobra.Command, pod *coreV1.Pod, container string, command []string, stdout io.Writer, stderr io.Writer) (remoteCommand.Executor, error) {
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
func (f *FakeCommandExecutor) Stream(options remoteCommand.StreamOptions) error {
	stdout := options.Stdout
	output := f.CommandResult[f.Command]
	stdout.Write([]byte(output))
	return nil
}

// StreamWithContext - stream command result with given context
func (f *FakeCommandExecutor) StreamWithContext(ctx context.Context, options remoteCommand.StreamOptions) error {
	stdout := options.Stdout
	output := f.CommandResult[f.Command]
	stdout.Write([]byte(output))
	return nil
}

var fakeCommandExecutor = &FakeCommandExecutor{}

// GetCommandWrapper - get command wrapper
func (f *FakeFactory) GetCommandWrapper(name string, args ...string) *bbUtilApiWrappers.Command {
	return fakeApiWrappers.NewFakeCommand(name, args...)
}

// GetIstioClientSet - get istio clientset
func (f *FakeFactory) GetIstioClientSet(cfg *rest.Config) (bbUtilApiWrappers.IstioClientset, error) {
	return fakeApiWrappers.NewFakeIstioClientSet(f.virtualServiceList), nil
}

// SetConfigClient sets the configuration client returned by the fake factory.
// This may be useful for tests that set configuration values directly that bypass
// the viper instance.
func (f *FakeFactory) SetConfigClient(configClient *bbConfig.ConfigClient) {
	f.configClient = configClient
}

// GetConfigClient - get config client
func (f *FakeFactory) GetConfigClient(command *cobra.Command) (*bbConfig.ConfigClient, error) {
	// if SetConfigClient has been previously called and an alternative client
	// has been attached, return it
	if f.configClient != nil {
		return f.configClient, nil

	}

	if f.SetFail.GetConfigClient {
		return nil, fmt.Errorf("failed to get config client")
	}
	clientGetter := bbConfig.ClientGetter{}
	loggingClient := f.GetLoggingClient()
	client, err := clientGetter.GetClient(command, &loggingClient, f.GetViper())
	return client, err
}

// GetViper - get viper
func (f *FakeFactory) GetViper() *viper.Viper {
	return f.viperInstance
}
