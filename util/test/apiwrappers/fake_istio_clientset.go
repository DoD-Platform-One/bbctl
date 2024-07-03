package apiwrappers

import (
	"context"

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
}

// FakeVirtualService
type FakeVirtualService struct {
	typedV1Beta1.VirtualServiceExpansion

	VirtualServicesList *apisV1Beta1.VirtualServiceList
}

// NewFakeIstioClientSet intializes and returns a new FakeIstioClientSet
func NewFakeIstioClientSet(vsList *apisV1Beta1.VirtualServiceList) *FakeIstioClientSet {
	return &FakeIstioClientSet{VirtualServicesList: vsList}
}

// Fake Clientset functions

// NetworkingV1beta1 intializes and returns a new FakeNetworkingV1beta1 object containing the configured list of virtual services
func (f *FakeIstioClientSet) NetworkingV1beta1() typedV1Beta1.NetworkingV1beta1Interface {
	return &FakeNetworkingV1beta1{
		VirtualServicesList: f.VirtualServicesList,
	}
}

// Fake NetworkingV1beta1Interface functions

// DestinationRules returns nil
func (f *FakeNetworkingV1beta1) DestinationRules(namespace string) typedV1Beta1.DestinationRuleInterface {
	return nil
}

// Gateways returns nil
func (f *FakeNetworkingV1beta1) Gateways(namespace string) typedV1Beta1.GatewayInterface {
	return nil
}

// ProxyConfigs returns nil
func (f *FakeNetworkingV1beta1) ProxyConfigs(namespace string) typedV1Beta1.ProxyConfigInterface {
	return nil
}

// RESTClient returns nil
func (f *FakeNetworkingV1beta1) RESTClient() rest.Interface {
	return nil
}

// ServiceEntries returns nil
func (f *FakeNetworkingV1beta1) ServiceEntries(namespace string) typedV1Beta1.ServiceEntryInterface {
	return nil
}

// Sidecars returns nil
func (f *FakeNetworkingV1beta1) Sidecars(namespace string) typedV1Beta1.SidecarInterface {
	return nil
}

// VirtualServices returns nil
func (f *FakeNetworkingV1beta1) VirtualServices(namespace string) typedV1Beta1.VirtualServiceInterface {
	return &FakeVirtualService{
		VirtualServicesList: f.VirtualServicesList,
	}
}

// WorkloadEntries returns nil
func (f *FakeNetworkingV1beta1) WorkloadEntries(namespace string) typedV1Beta1.WorkloadEntryInterface {
	return nil
}

// WorkloadGroups returns nil
func (f *FakeNetworkingV1beta1) WorkloadGroups(namespace string) typedV1Beta1.WorkloadGroupInterface {
	return nil
}

// Fake VirtualService functions

// Create returns nil, nil
func (f *FakeVirtualService) Create(ctx context.Context, virtualService *apisV1Beta1.VirtualService, opts v1.CreateOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil
}

// Update returns nil, nil
func (f *FakeVirtualService) Update(ctx context.Context, virtualService *apisV1Beta1.VirtualService, opts v1.UpdateOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil
}

// UpdateStatus returns nil, nil
func (f *FakeVirtualService) UpdateStatus(ctx context.Context, virtualService *apisV1Beta1.VirtualService, opts v1.UpdateOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil
}

// Delete returns nil
func (f *FakeVirtualService) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return nil
}

// DeleteCollection returns nil
func (f *FakeVirtualService) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	return nil
}

// Get returns nil, nil
func (f *FakeVirtualService) Get(ctx context.Context, name string, opts v1.GetOptions) (*apisV1Beta1.VirtualService, error) {
	return nil, nil
}

// List returns a list of virtual service resources
func (f *FakeVirtualService) List(ctx context.Context, opts v1.ListOptions) (*apisV1Beta1.VirtualServiceList, error) {
	return f.VirtualServicesList, nil
}

// Watch returns nil, nil
func (f *FakeVirtualService) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

// Patch returns nil, nil
func (f *FakeVirtualService) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, args ...string) (*apisV1Beta1.VirtualService, error) {
	return nil, nil
}

// Apply returns nil, nil
func (f *FakeVirtualService) Apply(ctx context.Context, virtualService *networkingV1Beta1.VirtualServiceApplyConfiguration, opts v1.ApplyOptions) (result *apisV1Beta1.VirtualService, err error) {
	return nil, nil
}

// ApplyStatus returns nil, nil
func (f *FakeVirtualService) ApplyStatus(ctx context.Context, virtualService *networkingV1Beta1.VirtualServiceApplyConfiguration, opts v1.ApplyOptions) (result *apisV1Beta1.VirtualService, err error) {
	return nil, nil
}
