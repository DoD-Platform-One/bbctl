package k3d

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	apiV1Beta1 "istio.io/api/networking/v1beta1"
	apisV1Beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

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
	cmd := NewHostsCmd(factory)
	// Assert
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
			streams := factory.GetIOStream()
			in := streams.In.(*bytes.Buffer)
			out := streams.Out.(*bytes.Buffer)
			errOut := streams.ErrOut.(*bytes.Buffer)
			if tc.shouldErr {
				streams.Out = apiWrappers.CreateFakeWriterFromStream(t, tc.shouldErr, streams.Out)
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
			viperInstance := factory.GetViper()
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kubeconfig", "../../util/test/data/kube-config.yaml")
			cmd := NewHostsCmd(factory)
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
					assert.Empty(t, out.String())
					assert.Contains(t, errOut.String(), (&apiWrappers.FakeWriterError{}).Error())
				} else {
					assert.Nil(t, err)
					assert.Empty(t, out.String())
					assert.Empty(t, errOut.String())
				}
				assert.Empty(t, in.String())
			} else {
				assert.Nil(t, err)
				assert.Empty(t, errOut.String())
				assert.Empty(t, in.String())
				assert.Equal(t, fmt.Sprintf("%v\t%v\t%v\n", privateIP, vs.Spec.Hosts[0], vs.Spec.Hosts[1]), out.String())
			}
		})
	}
}
