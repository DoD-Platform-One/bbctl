package test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbUtilApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/apiwrappers"
	bbAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/commoninterfaces"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	bbGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/gitlab"
	helm "repo1.dso.mil/big-bang/product/packages/bbctl/util/helm"
	bbLog "repo1.dso.mil/big-bang/product/packages/bbctl/util/log"
	bbOutput "repo1.dso.mil/big-bang/product/packages/bbctl/util/output"
	fakeApiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
	fakeAws "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/aws"
	fakeGitLab "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/gitlab"
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
	factory.fakeCommandExecutor = &FakeCommandExecutor{}
	return factory
}

// Set functions

// SetHelmReleases - set helm releases
func (f *FakeFactory) SetHelmReleases(helmReleases []*release.Release) {
	f.helm.releases = helmReleases
}

// SetHelmGetReleaseFunc sets the GetRelease function on the fake helm client
func (f *FakeFactory) SetHelmGetReleaseFunc(getReleaseFunc helm.GetReleaseFunc) {
	f.helm.getRelease = getReleaseFunc
}

// SetHelmGetValuesFunc sets the GetValues function on the fake helm client
func (f *FakeFactory) SetHelmGetValuesFunc(getValuesFunc helm.GetValuesFunc) {
	f.helm.getValues = getValuesFunc
}

// SetHelmGetListFunc sets the GetList function on the fake helm client
func (f *FakeFactory) SetHelmGetListFunc(getListFunc helm.GetListFunc) {
	f.helm.getList = getListFunc
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
		streams, err := f.GetIOStream()
		if err != nil {
			panic(err)
		}
		loggingFuncToUse = func(args ...string) {
			_, err = streams.ErrOut.Write([]byte(strings.Join(args, "\n")))
			if err != nil {
				panic(err)
			}
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
	awsConfig           aws.Config
	callerIdentity      *bbAws.CallerIdentity
	clusterIPs          []bbAws.ClusterIP
	ec2Client           *ec2.Client
	loggingFunc         fakeLog.LoggingFunction
	objects             []runtime.Object
	gvrToListKind       map[schema.GroupVersionResource]string
	resources           []*metaV1.APIResourceList
	stsClient           *sts.Client
	virtualServiceList  *apisV1Beta1.VirtualServiceList
	viperInstance       *viper.Viper
	configClient        *bbConfig.ConfigClient
	fakeCommandExecutor *FakeCommandExecutor
	pipeReader          commonInterfaces.FileLike
	pipeWriter          commonInterfaces.FileLike

	SetFail struct {
		GetCredentialHelper          bool
		GetConfigClient              int // the number of times to pass before returning an error every time, 0 is never fail
		getConfigClientCount         int // the number of times the GetConfigClient function has been called
		GetHelmClient                bool
		GetOutputClient              bool
		GetK8sDynamicClient          bool
		GetK8sDynamicClientPrepFuncs []*func(clientset *dynamicFake.FakeDynamicClient)
		GetK8sClientset              bool
		GetK8sClientsetPrepFuncs     []*func(clientset *fake.Clientset)
		GetCommandExecutor           bool
		GetCommandWrapper            bool
		SetCommandWrapperRunError    bool
		GetPolicyClient              bool
		GetCrds                      bool
		GetDescriptor                bool
		DescriptorType               string
		GetAWSClient                 bool
		GetIstioClient               bool
		GetIOStreams                 int // the number of times to pass before returning an error every time, 0 is never fail
		getIOStreamsCount            int // the number of times the GetIOStreams function has been called
		GetLoggingClient             bool
		GetPipe                      bool
		GetRuntimeClient             bool
		GetViper                     int // the number of times to pass before returning an error every time, 0 is never fail
		GetViperCount                int // the number of times the GetViper function has been called
		GetGitLabClient              bool

		// configure the AWS fake client and fake istio client to fail on certain calls
		// configure the AWS fake client to fail on certain calls
		AWS   fakeAws.SetFail
		Istio fakeApiWrappers.SetFail
	}

	helm struct {
		releases   []*release.Release
		getRelease helm.GetReleaseFunc
		getList    helm.GetListFunc
		getValues  helm.GetValuesFunc
	}

	credentialHelper bbUtil.CredentialHelper
	gitlab           struct {
		getFileFunc fakeGitLab.GetFileFunc
	}
}

// GetCredentialHelper - get credential helper
func (f *FakeFactory) GetCredentialHelper() (bbUtil.CredentialHelper, error) {
	if f.SetFail.GetCredentialHelper {
		return nil, errors.New("failed to get credential helper")
	}
	if f.credentialHelper == nil {
		f.credentialHelper = func(_ string, _ string) (string, error) {
			return "", nil
		}
	}
	return f.credentialHelper, nil
}

func (f *FakeFactory) SetCredentialHelper(credentialHelper bbUtil.CredentialHelper) {
	f.credentialHelper = credentialHelper
}

// GetAWSClient constructs a fake AWS client
func (f *FakeFactory) GetAWSClient() (bbAws.Client, error) {
	if f.SetFail.GetAWSClient {
		return nil, errors.New("failed to get AWS client")
	}
	fakeClient, err := fakeAws.NewFakeClient(
		f.clusterIPs,
		&f.awsConfig,
		f.ec2Client,
		f.callerIdentity,
		f.stsClient,
		f.SetFail.AWS,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS client: %w", err)
	}
	return fakeClient, nil
}

// SetGitLabGetFileFunc sets the GetFile function on the fake GitLab client
func (f *FakeFactory) SetGitLabGetFileFunc(getFileFunc fakeGitLab.GetFileFunc) {
	f.gitlab.getFileFunc = getFileFunc
}

// GetGitLabClient constructs a fake GitLab client
func (f *FakeFactory) GetGitLabClient() (bbGitLab.Client, error) {
	// Fail if the GetGitLabClient function has been called with a set fail
	if f.SetFail.GetGitLabClient {
		return nil, errors.New("failed to get GitLab client")
	}

	fakeClient, err := fakeGitLab.NewFakeClient("https://localhost.com", "", f.gitlab.getFileFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitLab client: %w", err)
	}
	return fakeClient, nil
}

// GetHelmClient - get helm client
func (f *FakeFactory) GetHelmClient(_ *cobra.Command, _ string) (helm.Client, error) {
	if f.SetFail.GetHelmClient {
		return nil, errors.New("failed to get helm client")
	}

	return fakeHelm.NewFakeClient(
		f.helm.getRelease,
		f.helm.getList,
		f.helm.getValues,
		f.helm.releases,
	)
}

// GetClientSet - get clientset
func (f *FakeFactory) GetClientSet() (kubernetes.Interface, error) {
	fakeClient := fake.NewSimpleClientset()
	return fakeClient, nil
}

// GetOutputClient
func (f *FakeFactory) GetOutputClient(cmd *cobra.Command) (bbOutput.Client, error) {
	if f.SetFail.GetOutputClient {
		return nil, errors.New("failed to get output client")
	}
	streams, err := f.GetIOStream()
	if err != nil {
		return nil, err
	}
	configClient, err := f.GetConfigClient(cmd)
	if err != nil {
		return nil, err
	}
	config, err := configClient.GetConfig()
	if err != nil {
		return nil, err
	}
	outputClientGetter := bbOutput.ClientGetter{}
	outputClient := outputClientGetter.GetClient(config.OutputConfiguration.Format, *streams)

	return outputClient, nil
}

// GetK8sClientset - get k8s clientset
func (f *FakeFactory) GetK8sClientset(_ *cobra.Command) (kubernetes.Interface, error) {
	if f.SetFail.GetK8sClientset {
		return nil, errors.New("testing error")
	}
	cs := fake.NewSimpleClientset(f.objects...)
	if f.resources != nil {
		cs.Fake.Resources = f.resources
	}
	if len(f.SetFail.GetK8sClientsetPrepFuncs) > 0 {
		for _, prepFunc := range f.SetFail.GetK8sClientsetPrepFuncs {
			(*prepFunc)(cs)
		}
	}
	return cs, nil
}

// GetK8sDynamicClient - get k8s dynamic client
func (f *FakeFactory) GetK8sDynamicClient(_ *cobra.Command) (dynamic.Interface, error) {
	if f.SetFail.GetK8sDynamicClient {
		return nil, errors.New("failed to get K8sDynamicClient client")
	}

	if f.SetFail.GetPolicyClient {
		client := &badClient{}
		if f.SetFail.GetCrds {
			client.FailCrd = true
		}
		if f.SetFail.GetDescriptor {
			client.FailDescriptor = true
			v, err := f.GetViper()
			if err != nil {
				return nil, err
			}
			if v.Get("gatekeeper") == true {
				client.Gatekeeper = true
			}
			client.DescriptorType = f.SetFail.DescriptorType
		}
		return client, nil
	}

	scheme := runtime.NewScheme()
	err := coreV1.AddToScheme(scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to add coreV1 to scheme: %w", err)
	}
	client := dynamicFake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		f.gvrToListKind,
		f.objects...)
	for _, prepFunc := range f.SetFail.GetK8sDynamicClientPrepFuncs {
		(*prepFunc)(client)
	}
	return client, nil
}

// GetLoggingClient - get logging client
func (f *FakeFactory) GetLoggingClient() (bbLog.Client, error) {
	if f.SetFail.GetLoggingClient {
		return nil, errors.New("failed to get logging client")
	}
	return f.GetLoggingClientWithLogger(nil)
}

// GetLoggingClientWithLogger - get logging client providing logger
func (f *FakeFactory) GetLoggingClientWithLogger(_ *slog.Logger) (bbLog.Client, error) {
	client := fakeLog.NewFakeClient(f.loggingFunc)
	return client, nil
}

// GetRestConfig - get rest config
func (f *FakeFactory) GetRestConfig(_ *cobra.Command) (*rest.Config, error) {
	return &rest.Config{}, nil
}

// GetRuntimeClient - get runtime client
func (f *FakeFactory) GetRuntimeClient(scheme *runtime.Scheme) (client.Client, error) {
	if f.SetFail.GetRuntimeClient {
		return nil, errors.New("test error")
	}
	cb := fakeControllerClient.NewClientBuilder()
	rc := cb.WithScheme(scheme).Build()
	return rc, nil
}

// GetCommandExecutor - execute command in a Pod
func (f *FakeFactory) GetCommandExecutor(
	_ *cobra.Command,
	_ *coreV1.Pod,
	_ string,
	command []string,
	_ io.Writer,
	_ io.Writer,
) (remoteCommand.Executor, error) {
	if f.SetFail.GetCommandExecutor {
		return nil, errors.New("testing error")
	}
	f.fakeCommandExecutor.Command = strings.Join(command, " ")
	return f.fakeCommandExecutor, nil
}

// GetFakeCommandExecutor - get fake command executor
func (f *FakeFactory) GetFakeCommandExecutor() (*FakeCommandExecutor, error) {
	if f.SetFail.GetCommandExecutor {
		return nil, errors.New("testing error")
	}
	return f.fakeCommandExecutor, nil
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
func (f *FakeCommandExecutor) StreamWithContext(
	_ context.Context,
	options remoteCommand.StreamOptions,
) error {
	stdout := options.Stdout
	output := f.CommandResult[f.Command]
	stdout.Write([]byte(output))
	return nil
}

// GetCommandWrapper - get command wrapper
func (f *FakeFactory) GetCommandWrapper(
	name string,
	args ...string,
) (*bbUtilApiWrappers.Command, error) {
	if f.SetFail.GetCommandWrapper {
		return nil, errors.New("failed to get command wrapper")
	}
	wrapper := fakeApiWrappers.NewFakeCommand(name, f.SetFail.SetCommandWrapperRunError, args...)
	streams, err := f.GetIOStream()
	if err != nil {
		return nil, err
	}
	wrapper.SetStdout(streams.Out)
	wrapper.SetStderr(streams.ErrOut)
	wrapper.SetStdin(streams.In)
	return wrapper, nil
}

// GetIstioClientSet - get istio clientset
func (f *FakeFactory) GetIstioClientSet(_ *rest.Config) (bbUtilApiWrappers.IstioClientset, error) {
	if f.SetFail.GetIstioClient {
		return nil, errors.New("failed to get istio clientset")
	}
	return fakeApiWrappers.NewFakeIstioClientSet(f.virtualServiceList, f.SetFail.Istio), nil
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
	f.SetFail.getConfigClientCount++
	if f.SetFail.GetConfigClient > 0 && f.SetFail.getConfigClientCount >= f.SetFail.GetConfigClient {
		return nil, errors.New("failed to get config client")
	}
	clientGetter := bbConfig.ClientGetter{}
	loggingClient, err := f.GetLoggingClient()
	if err != nil {
		return nil, err
	}
	v, err := f.GetViper()
	if err != nil {
		return nil, err
	}
	client, err := clientGetter.GetClient(command, &loggingClient, v)
	return client, err
}

// GetViper - get viper
func (f *FakeFactory) GetViper() (*viper.Viper, error) {
	f.SetFail.GetViperCount++
	if f.SetFail.GetViper > 0 && f.SetFail.GetViperCount >= f.SetFail.GetViper {
		return nil, errors.New("failed to get viper")
	}
	return f.viperInstance, nil
}

// SetViper sets the viper instance
func (f *FakeFactory) SetViper(v *viper.Viper) error {
	f.viperInstance = v
	return nil
}

// Temporary Singleton for IO Streams until implementation of bbctl #214
var (
	streams   *genericIOOptions.IOStreams //nolint:gochecknoglobals
	oneStream sync.Once                   //nolint:gochecknoglobals
)

// ResetIOStream resets the IOStreams singleton
func (f *FakeFactory) ResetIOStream() {
	streams = nil
	oneStream = sync.Once{}
}

// GetIOStream initializes and returns a new IOStreams object used to interact with console input, output, and error output
func (f *FakeFactory) GetIOStream() (*genericIOOptions.IOStreams, error) {
	f.SetFail.getIOStreamsCount++
	if f.SetFail.GetIOStreams > 0 && f.SetFail.getIOStreamsCount >= f.SetFail.GetIOStreams {
		return nil, errors.New("failed to get streams")
	}
	oneStream.Do(func() {
		streams = &genericIOOptions.IOStreams{
			In:     &bytes.Buffer{},
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		}
	})
	return streams, nil
}

func (f *FakeFactory) SetIOStream(stream *genericIOOptions.IOStreams) {
	streams = stream
}

// GetPipe - get the pipe reader and writer
func (f *FakeFactory) GetPipe() (commonInterfaces.FileLike, commonInterfaces.FileLike, error) {
	if f.SetFail.GetPipe {
		return nil, nil, errors.New("failed to get pipe")
	}
	if f.pipeReader != nil && f.pipeWriter != nil {
		return f.pipeReader, f.pipeWriter, nil
	}
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get pipe: %w", err)
	}
	err = f.SetPipe(r, w)
	return r, w, err
}

// SetPipe - set the pipe reader and writer
func (f *FakeFactory) SetPipe(reader commonInterfaces.FileLike, writer commonInterfaces.FileLike) error {
	if reader == nil || writer == nil {
		return errors.New("reader and writer must not be nil")
	}
	f.pipeReader = reader
	f.pipeWriter = writer
	return nil
}

// ResetPipe resets the pipe reader and writer to nil
func (f *FakeFactory) ResetPipe() {
	f.pipeReader = nil
	f.pipeWriter = nil
}
