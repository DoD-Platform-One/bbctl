package k3d

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbK8s "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"
)

var (
	hostsUse = `hosts`

	hostsShort = i18n.T(`Generates /etc/hosts entries for your k3d cluster`)

	hostsLong = templates.LongDesc(i18n.T(`Generates a list of hosts that reference your k3d cluster suitable for use in /etc/hosts`))

	hostsExample = templates.Examples(i18n.T(`
	    # Generate a list of hosts that reference your k3d cluster suitable for use in /etc/hosts
		bbctl k3d hosts`))
)

// NewHostsCmd - Returns a command to generate a hosts list for your k3d cluster using hostsListCluster
func NewHostsCmd(factory bbUtil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     hostsUse,
		Short:   hostsShort,
		Long:    hostsLong,
		Example: hostsExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(hostsListCluster(cmd, factory))
		},
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)

	loggingClient.HandleError(
		"Unable to add flags to command: %v",
		configClient.SetAndBindFlag(
			"private-ip",
			false,
			"Use the private IP instead of the public IP",
		),
	)

	return cmd
}

// hostsListCluster - Returns the error (nil if no error) when generating a hosts list for your k3d cluster
func hostsListCluster(cmd *cobra.Command, factory bbUtil.Factory) error {
	streams := factory.GetIOStream()
	var virtualServices []string

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("Unable to get config client: %v", err)
	config := configClient.GetConfig()
	k8sConfig, err := bbK8s.BuildKubeConfig(config)
	loggingClient.HandleError("Unable to build k8s configuration: %v", err)
	istioClientSet, err := factory.GetIstioClientSet(k8sConfig)
	loggingClient.HandleError("Unable to create istio client: %v", err)
	istioServices, err := istioClientSet.NetworkingV1beta1().VirtualServices("").List(context.TODO(), metaV1.ListOptions{})
	loggingClient.HandleError("Unable to list istio services: %v", err)
	k8sClient, err := factory.GetK8sClientset(cmd)
	loggingClient.HandleError("Unable to create k8s client: %v", err)
	allServices, err := k8sClient.CoreV1().Services("").List(context.TODO(), metaV1.ListOptions{})
	loggingClient.HandleError("Unable to list services: %v", err)

	// One line per service
	for _, loadBalancer := range allServices.Items {
		virtualServices = []string{}
		// Only considers services of type LoadBalancer
		if loadBalancer.Spec.Type != coreV1.ServiceTypeLoadBalancer {
			loggingClient.Debug("Skipping service %s of type %s\n", loadBalancer.Name, loadBalancer.Spec.Type)
			continue
		}
		// Checks all virtual istio services for a loadbalancer name or namespace matches
		for _, virtualService := range istioServices.Items {
			// Skips virtual services without hosts or gateways
			if len(virtualService.Spec.Hosts) == 0 ||
				len(virtualService.Spec.Gateways) == 0 {
				loggingClient.Warn("Skipping virtual service %s without hosts or gateways\n", virtualService.Name)
				continue
			}
			// Checks if the load balancer name or namespace matches the virtual service
			for _, gateway := range virtualService.Spec.Gateways {
				combinedName := fmt.Sprintf("%s/%s", loadBalancer.Namespace, loadBalancer.Name)
				if strings.Contains(loadBalancer.Name, gateway) ||
					// Adds the virtual service hosts to the list if a match is found
					strings.Contains(combinedName, gateway) {
					virtualServices = slices.Insert(virtualServices,
						0,
						virtualService.Spec.Hosts...,
					)
					break
				}
			}
		}
		// Skips if no virtual services were found in loadBalancer name or namespace
		if len(virtualServices) == 0 {
			loggingClient.Warn("Skipping service %s without virtual services\n", loadBalancer.Name)
			continue
		}
		// Print each cluster IP with the associated virtual services
		for _, clusterIP := range loadBalancer.Spec.ClusterIPs {
			msg := fmt.Sprintf("%s\t%s\n",
				clusterIP,
				strings.Join(virtualServices, "\t"),
			)
			_, err = streams.Out.Write([]byte(msg))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
