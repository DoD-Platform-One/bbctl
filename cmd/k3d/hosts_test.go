package k3d

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	apiV1Beta1 "istio.io/api/networking/v1beta1"
	apisV1Beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	bbAwsUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/aws"
	bbConfig "repo1.dso.mil/big-bang/product/packages/bbctl/util/config"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"
	apiWrappers "repo1.dso.mil/big-bang/product/packages/bbctl/util/test/apiwrappers"
)

const (
	callerIdentityAccount = "123456789012"
	callerIdentityArn     = "arn:aws:iam::123456789012:user/developer"
	privateIPConst        = "192.192.192.192"
	publicIPConst         = "172.172.172.172"
)

func TestK3d_NewHostsCmd(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, err := NewHostsCmd(factory)
	// Assert
	assert.Nil(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "hosts", cmd.Use)
}

func TestK3d_NewHostsCmd_Run(t *testing.T) {
	testCases := []struct {
		name       string
		vsName     string
		vsGateways []string
		svcName    string
		svcType    coreV1.ServiceType
		shouldFail bool
		shouldErr  bool
	}{
		{
			name:       "Pass case",
			vsName:     "test",
			vsGateways: []string{"test-gateway"},
			svcName:    "test-gateway",
			svcType:    coreV1.ServiceTypeLoadBalancer,
			shouldFail: false,
			shouldErr:  false,
		},
		{
			name:       "No service case",
			vsName:     "test",
			vsGateways: []string{"test-gateway"},
			svcName:    "test-gateway",
			svcType:    coreV1.ServiceTypeClusterIP,
			shouldFail: true,
			shouldErr:  false,
		},
		{
			name:       "No matching gateway case",
			vsName:     "test",
			vsGateways: []string{"bad-test-gateway"},
			svcName:    "test-gateway",
			svcType:    coreV1.ServiceTypeLoadBalancer,
			shouldFail: true,
			shouldErr:  false,
		},
		{
			name:       "No gateway case",
			vsName:     "test",
			vsGateways: []string{},
			svcName:    "test-gateway",
			svcType:    coreV1.ServiceTypeLoadBalancer,
			shouldFail: true,
			shouldErr:  false,
		},
		{
			name:       "Error on io.Writer case",
			vsName:     "test",
			vsGateways: []string{"test-gateway"},
			svcName:    "test-gateway",
			svcType:    coreV1.ServiceTypeLoadBalancer,
			shouldFail: true,
			shouldErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			streams, _ := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)
			if tc.shouldErr {
				streams.Out = apiWrappers.CreateFakeWriterFromReaderWriter(t, false, tc.shouldErr, out)
			}
			privateIP := privateIPConst

			vs := apisV1Beta1.VirtualService{
				TypeMeta: metaV1.TypeMeta{
					Kind:       "VirtualService",
					APIVersion: "networking.istio.io/v1beta1",
				},
				ObjectMeta: metaV1.ObjectMeta{
					Name:      tc.vsName,
					Namespace: "test",
				},
				Spec: apiV1Beta1.VirtualService{
					Hosts:    []string{"test1", "test2"},
					Gateways: tc.vsGateways,
				},
			}
			vsList := apisV1Beta1.VirtualServiceList{
				Items: []*apisV1Beta1.VirtualService{
					&vs,
				},
			}
			svc := coreV1.Service{
				TypeMeta: metaV1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: metaV1.ObjectMeta{
					Name:      tc.svcName,
					Namespace: "test",
				},
				Spec: coreV1.ServiceSpec{
					Type: tc.svcType,
					ClusterIPs: []string{
						privateIP,
					},
				},
			}
			svcList := coreV1.ServiceList{
				Items: []coreV1.Service{
					svc,
				},
			}
			factory.SetObjects([]runtime.Object{&svcList})
			factory.SetVirtualServices(&vsList)
			viperInstance, _ := factory.GetViper()
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("output-config.format", "yaml")
			viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
			cmd, cmdErr := NewHostsCmd(factory)
			assert.Nil(t, cmdErr)
			cmd.SetArgs([]string{"--private-ip"})
			var err error

			// Act
			if os.Getenv("BE_CRASHER") == "1" {
				if !tc.shouldErr {
					return
				}
				assert.Nil(t, cmd.Execute())
				return
			}

			if tc.shouldErr {
				runCrasherCommand := exec.Command(os.Args[0], "-test.run=TestK3d_NewHostsCmd_Run")
				runCrasherCommand.Env = append(os.Environ(), "BE_CRASHER=1")
				runCrasherCommand.Stderr = errOut
				runCrasherCommand.Stdout = out
				runCrasherCommand.Stdin = in
				err = runCrasherCommand.Run()
			} else {
				err = cmd.Execute()
			}

			// Assert
			assert.NotNil(t, cmd)
			assert.Equal(t, "hosts", cmd.Use)
			if tc.shouldFail {
				if tc.shouldErr {
					assert.NotNil(t, err)
					assert.Contains(t, out.String(), "FAIL")
					assert.Contains(t, errOut.String(), (&apiWrappers.FakeWriterError{}).Error())
				} else {
					assert.Nil(t, err)
					assert.Equal(t, "hosts: {}\n", out.String())
					assert.Empty(t, errOut.String())
				}
				assert.Empty(t, in.String())
			} else {
				assert.Nil(t, err)
				assert.Empty(t, errOut.String())
				assert.Empty(t, in.String())
				// hosts:\n  192.192.192.192:\n  - test1\n  - test2\n
				assert.Equal(t, fmt.Sprintf("hosts:\n  %v:\n  - %v\n  - %v\n", privateIP, vs.Spec.Hosts[0], vs.Spec.Hosts[1]), out.String())
			}
		})
	}
}

func TestK3d_hostsListClusterErrors(t *testing.T) {
	goodkubeconfig := "../../util/test/data/kube-config.yaml"
	badkubeconfig := "../test/data/bad-kube-config.yaml"
	tests := []struct {
		name string
		// errorFunc is a function that will be called with the awsClient and factory
		// at the start of a test case to allow setting flags to force errors
		errorFunc func(factory *bbTestUtil.FakeFactory)
		errmsg    string
	}{
		{
			name: "ErrorGettingIOStreams",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.GetIOStreams = 1
			},
			errmsg: "unable to get output client: failed to get streams",
		},
		{
			name: "ErrorGettingLoggingClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.GetLoggingClient = true
			},
			errmsg: "unable to get logging client: failed to get logging client",
		},
		{
			name: "ErrorGettingConfigClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.GetConfigClient = 1
			},
			errmsg: "unable to get config client: failed to get config client",
		},
		{
			name: "ErrorBuildingK8sConfig",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				viperInstance, viperErr := factory.GetViper()
				assert.Nil(t, viperErr)
				viperInstance.Set("kubeconfig", badkubeconfig)
			},
			errmsg: fmt.Sprintf(
				"unable to build k8s configuration: stat %s: no such file or directory",
				badkubeconfig,
			),
		},
		{
			name: "ErrorCreatingIstioClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.GetIstioClient = true
			},
			errmsg: "unable to create istio client: failed to get istio clientset",
		},
		{
			name: "ErrorListingIstioClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.Istio.GetList = true
			},
			errmsg: "unable to list istio services: failed to list istio services",
		},
		{
			name: "ErrorCreatingK8sClient",
			errorFunc: func(factory *bbTestUtil.FakeFactory) {
				factory.SetFail.GetK8sClientset = true
			},
			errmsg: "unable to create k8s client: testing error",
		},
	}

	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	streams, streamsErr := factory.GetIOStream()
	out := streams.Out.(*bytes.Buffer)
	assert.Nil(t, streamsErr)
	streams.Out = apiWrappers.CreateFakeWriterFromReaderWriter(t, false, true, out)
	account := callerIdentityAccount
	arn := callerIdentityArn
	callerIdentity := bbAwsUtil.CallerIdentity{
		GetCallerIdentityOutput: sts.GetCallerIdentityOutput{
			Account: &account,
			Arn:     &arn,
		},
		Username: "developer",
	}
	clusterIPs := []bbAwsUtil.ClusterIP{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetCallerIdentity(&callerIdentity)
			factory.SetClusterIPs(&clusterIPs)
			viperInstance, viperErr := factory.GetViper()
			assert.Nil(t, viperErr)
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", goodkubeconfig)

			// Trigger our errors
			test.errorFunc(factory)

			cmd, _ := NewHostsCmd(factory)
			err := hostsListCluster(cmd, factory)

			assert.NotNil(t, err)
			assert.Equal(t, test.errmsg, err.Error())
		})
	}
}

func TestK3d_HostsListCluster_ListAllError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	bigBangRepoLocation := "test"
	viperInstance, viperErr := factory.GetViper()
	assert.Nil(t, viperErr)
	viperInstance.Set("big-bang-repo", bigBangRepoLocation)
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	// Act
	cmd, err := NewHostsCmd(factory)
	assert.NotNil(t, cmd)
	assert.Nil(t, err)

	listAllErr := HostsListCluster(cmd, factory, true)

	// Assert
	assert.Error(t, listAllErr)
	if !assert.Contains(t, listAllErr.Error(), "unable to list all services:") {
		t.Errorf("unexpected output: %s", listAllErr.Error())
	}
}

func TestK3d_NewHostsCmd_BindFlagsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance, viperErr := factory.GetViper()
	assert.Nil(t, viperErr)
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")

	expectedError := fmt.Errorf("failed to set and bind flag")
	setAndBindFlagFunc := func(client *bbConfig.ConfigClient, name string, shortHand string, value interface{}, description string) error {
		if name == "private-ip" {
			return expectedError
		}
		return nil
	}

	logClient, logClientErr := factory.GetLoggingClient()
	assert.Nil(t, logClientErr)
	configClient, err := bbConfig.NewClient(nil, setAndBindFlagFunc, &logClient, nil, viperInstance)
	assert.Nil(t, err)
	factory.SetConfigClient(configClient)

	// Act
	cmd, err := NewHostsCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("unable to bind flags: %s", expectedError.Error()), err.Error())
}

func TestK3d_NewHostsCmd_ConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.ResetIOStream()
	bigBangRepoLocation := "test"
	viperInstance, _ := factory.GetViper()
	viperInstance.Set("big-bang-repo", bigBangRepoLocation)
	viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
	factory.SetFail.GetConfigClient = 1

	// Act
	cmd, err := NewHostsCmd(factory)

	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestHostsFailToGetConfig(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	loggingClient, _ := factory.GetLoggingClient()
	cmd, _ := NewHostsCmd(factory)
	viper, _ := factory.GetViper()
	expected := ""
	getConfigFunc := func(client *bbConfig.ConfigClient) (*schemas.GlobalConfiguration, error) {
		return &schemas.GlobalConfiguration{
			BigBangRepo: expected,
		}, fmt.Errorf("Dummy Error")
	}
	client, _ := bbConfig.NewClient(getConfigFunc, nil, &loggingClient, cmd, viper)
	factory.SetConfigClient(client)

	// Act
	err := hostsListCluster(cmd, factory)

	// Assert
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting config:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}
