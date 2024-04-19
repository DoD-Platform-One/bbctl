package apiwrappers

import (
	v1Beta1 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1beta1"
)

// IstioClientset interface
type IstioClientset interface {
	NetworkingV1beta1() v1Beta1.NetworkingV1beta1Interface
}
