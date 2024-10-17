package apiwrappers

import (
	"context"
	"errors"

	apisV1Beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	networkingV1Beta1 "istio.io/client-go/pkg/applyconfiguration/networking/v1beta1"
	typedV1Beta1 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

// FakeIstioClientSet
type FakeIstioClientSet struct {
	VirtualServicesList *apisV1Beta1.VirtualServiceList

	setFail SetFail
}

// Flags to control the fake client behavior and force functions to fail
type SetFail struct {
	GetList bool
}

// FakeNetworkingV1beta1
type FakeNetworkingV1beta1 struct {
	DestinationRulesGetter typedV1Beta1.DestinationRulesGetter
	GatewaysGetter         typedV1Beta1.GatewaysGetter
	ProxyConfigsGetter     typedV1Beta1.ProxyConfigsGetter
	ServiceEntriesGetter   typedV1Beta1.ServiceEntriesGetter
	SidecarsGetter         typedV1Beta1.SidecarsGetter
	VirtualServicesGetter  typedV1Beta1.VirtualServicesGetter
	WorkloadEntriesGetter  typedV1Beta1.WorkloadEntriesGetter
	WorkloadGroupsGetter   typedV1Beta1.WorkloadGroupsGetter

	VirtualServicesList *apisV1Beta1.VirtualServiceList
	setFail             SetFail
}

// FakeVirtualService
type FakeVirtualService struct {
	typedV1Beta1.VirtualServiceExpansion
	VirtualServicesList *apisV1Beta1.VirtualServiceList

	setFail SetFail
}

// NewFakeIstioClientSet intializes and returns a new FakeIstioClientSet
func NewFakeIstioClientSet(vsList *apisV1Beta1.VirtualServiceList, sf SetFail) *FakeIstioClientSet {
	return &FakeIstioClientSet{
		VirtualServicesList: vsList,
		setFail:             sf,
	}
}

// Fake Clientset functions

// NetworkingV1beta1 intializes and returns a new FakeNetworkingV1beta1 object containing the configured list of virtual services
func (f *FakeIstioClientSet) NetworkingV1beta1() typedV1Beta1.NetworkingV1beta1Interface {
	return &FakeNetworkingV1beta1{
		VirtualServicesList: f.VirtualServicesList,
		setFail:             f.setFail,
	}
}

// Fake NetworkingV1beta1Interface functions

// DestinationRules returns nil
func (f *FakeNetworkingV1beta1) DestinationRules(_ string) typedV1Beta1.DestinationRuleInterface {
	return nil
}

// Gateways returns nil
func (f *FakeNetworkingV1beta1) Gateways(_ string) typedV1Beta1.GatewayInterface {
	return nil
}

// ProxyConfigs returns nil
func (f *FakeNetworkingV1beta1) ProxyConfigs(_ string) typedV1Beta1.ProxyConfigInterface {
	return nil
}

// RESTClient returns nil
func (f *FakeNetworkingV1beta1) RESTClient() rest.Interface {
	return nil
}

// ServiceEntries returns nil
func (f *FakeNetworkingV1beta1) ServiceEntries(_ string) typedV1Beta1.ServiceEntryInterface {
	return nil
}

// Sidecars returns nil
func (f *FakeNetworkingV1beta1) Sidecars(_ string) typedV1Beta1.SidecarInterface {
	return nil
}

// VirtualServices returns nil
func (f *FakeNetworkingV1beta1) VirtualServices(_ string) typedV1Beta1.VirtualServiceInterface {
	return &FakeVirtualService{
		VirtualServicesList: f.VirtualServicesList,
		setFail:             f.setFail,
	}
}

// WorkloadEntries returns nil
func (f *FakeNetworkingV1beta1) WorkloadEntries(_ string) typedV1Beta1.WorkloadEntryInterface {
	return nil
}

// WorkloadGroups returns nil
func (f *FakeNetworkingV1beta1) WorkloadGroups(_ string) typedV1Beta1.WorkloadGroupInterface {
	return nil
}

// Fake VirtualService functions

// Create returns nil, nil
func (f *FakeVirtualService) Create(_ context.Context, _ *apisV1Beta1.VirtualService, _ v1.CreateOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil //nolint:nilnil
}

// Update returns nil, nil
func (f *FakeVirtualService) Update(_ context.Context, _ *apisV1Beta1.VirtualService, _ v1.UpdateOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil //nolint:nilnil
}

// UpdateStatus returns nil, nil
func (f *FakeVirtualService) UpdateStatus(_ context.Context, _ *apisV1Beta1.VirtualService, _ v1.UpdateOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil //nolint:nilnil
}

// Delete returns nil
func (f *FakeVirtualService) Delete(_ context.Context, _ string, _ v1.DeleteOptions) error {
	return nil
}

// DeleteCollection returns nil
func (f *FakeVirtualService) DeleteCollection(_ context.Context, _ v1.DeleteOptions, _ v1.ListOptions) error {
	return nil
}

// Get returns nil, nil
func (f *FakeVirtualService) Get(_ context.Context, _ string, _ v1.GetOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil //nolint:nilnil
}

// List returns a list of virtual service resources
func (f *FakeVirtualService) List(_ context.Context, _ v1.ListOptions) (*apisV1Beta1.VirtualServiceList, error) {
	if f.setFail.GetList {
		return nil, errors.New("failed to list istio services")
	}
	return f.VirtualServicesList, nil
}

// Watch returns nil, nil
func (f *FakeVirtualService) Watch(_ context.Context, _ v1.ListOptions) (watch.Interface, error) {
	return nil, nil //nolint:nilnil
}

// Patch returns nil, nil
func (f *FakeVirtualService) Patch(_ context.Context, _ string, _ types.PatchType, _ []byte, _ v1.PatchOptions, _ ...string) (*apisV1Beta1.VirtualService, error) {
	return nil, nil //nolint:nilnil
}

// Apply returns nil, nil
func (f *FakeVirtualService) Apply(_ context.Context, _ *networkingV1Beta1.VirtualServiceApplyConfiguration, _ v1.ApplyOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil //nolint:nilnil
}

// ApplyStatus returns nil, nil
func (f *FakeVirtualService) ApplyStatus(_ context.Context, _ *networkingV1Beta1.VirtualServiceApplyConfiguration, _ v1.ApplyOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil //nolint:nilnil
}
