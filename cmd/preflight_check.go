package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	"repo1.dso.mil/big-bang/product/packages/bbctl/util/config/schemas"
	bbUtilK8s "repo1.dso.mil/big-bang/product/packages/bbctl/util/k8s"

	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	remoteCommand "k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	metricsApi "k8s.io/metrics/pkg/apis/metrics"
)

var (
	preflightCheckUse = `preflight-check`

	preflightCheckShort = i18n.T(`Check cluster for the minimum required configurations before installing Big Bang.`)

	preflightCheckLong = templates.LongDesc(i18n.T(`
		Check cluster for the minimum required configurations before installing Big Bang.
		This command creates a job in the 'preflight-check' namespace to check system parameters.
		User needs to have RBAC permissions to create and delete namespace, secret and job resources.`))

	preflightCheckExample = templates.Examples(i18n.T(`
		# Check cluster for the minimum required configurations
		bbctl preflight-check --registryserver <registry-server> --registryusername <username> --registrypassword <password>
		# Export registry credentials as environment variables before running this command to configure registry access
		# export REGISTRYSERVER=registry1.dso.mil
		# export REGISTRYUSERNAME=<username>
		# export REGISTRYPASSWORD=<password>
		bbctl preflight-check `))

	preflightPodImage = "registry1.dso.mil/ironbank/redhat/ubi/ubi8-minimal:8.4"

	preflightPodNamespace = "preflight-check"

	preflightPodImagePullSecret = "registry-secret"

	preflightPodName = "preflightcheck"

	fluxNamespace = "flux-system"
)

// preflightCheckFunc is a type definition that allows each preflight check step to provide its own implementation
type preflightCheckFunc func(*cobra.Command, bbUtil.Factory, *schemas.GlobalConfiguration) preflightCheckStatus

// preflightCheckStatus is a type definition that represents the output value of a single preflight check step
//
// Should only contain one of the constant values defined below: `Failed`, `Passed`, or `Unknown`
type preflightCheckStatus string

// Define all the possible preflightCheckStatus values as constants
const (
	failed  preflightCheckStatus = "Failed"  // check failed
	passed  preflightCheckStatus = "Passed"  // check passed
	unknown preflightCheckStatus = "Unknown" // check execution error
)

// preflightCheck defines the format for each step of the preflight check command including the implementation function
// and the output messages to display on success or failure
//
// The status value is populated using the return value of the `function` call as part of the bbPreflightCheck function
type preflightCheck struct {
	desc           string               // check description
	function       preflightCheckFunc   // function with check logic
	status         preflightCheckStatus // function execution status
	failureMessage string               // user friendly failure message
	successMessage string               // user friendly success message
}

// preflightChecks defines all the steps to run in the bbPreflightCheck function
var preflightChecks []preflightCheck = []preflightCheck{
	{
		desc:     "Metrics Server Check",
		function: checkMetricsServer,
		failureMessage: templates.LongDesc(i18n.T(`
			Metrics Server needs to be running in the cluster for Horizontal Pod Autoscaler to work.`)),
		successMessage: templates.LongDesc(i18n.T(`
			Metrics Server is running in the cluster for Horizontal Pod Autoscaler to work.`)),
	},
	{
		desc:     "Default Storage Class Check",
		function: checkDefaultStorageClass,
		failureMessage: templates.LongDesc(i18n.T(`
			A Default Storage Class must be defined for Stateful workloads. 
			If you don't have a need for Persistent Volumes, this error can be ignored.`)),
		successMessage: templates.LongDesc(i18n.T(`
			Default Storage Class exists for Stateful workloads to work.`)),
	},
	{
		desc:     "Flux Controller Check",
		function: checkFluxController,
		failureMessage: templates.LongDesc(i18n.T(`
			Flux Controller is required for successful installation of Big Bang packages using GitOps.`)),
		successMessage: templates.LongDesc(i18n.T(`
			Flux Controller is running and allows for successful installation of Big Bang packages using GitOps.`)),
	},
	{
		desc:     "System Parameters Check",
		function: checkSystemParameters,
		failureMessage: templates.LongDesc(i18n.T(`
			Some packages installed by Big Bang require system parameters to be equal or greater than the recommended value. 
			You can ignore this error if not planning to install packages that failed the check.
			For more information refer to https://repo1.dso.mil/big-bang/bigbang/-/blob/master/docs/prerequisites/os-preconfiguration.md`)),
		successMessage: templates.LongDesc(i18n.T(`
			System parameters determined to be equal or greater than the recommended value. 
			This will allow for successful installation of packages that passed the check.
			For more information refer to https://repo1.dso.mil/big-bang/bigbang/-/blob/master/docs/prerequisites/os-preconfiguration.md`)),
	},
}

// systemParameter defines the format for each of the system parameter checks
type systemParameter struct {
	// parameter name
	name string
	// command to execute to get paramater value
	command []string
	// parameter description
	description string
	// map of package name and minimum parameter value
	threshold map[string]int
}

// sysParams defines all the system parameter checks to run as part of the checkSystemParameters step
var sysParams []systemParameter = []systemParameter{
	{
		"vm.max_map_count",
		[]string{"cat", "/proc/sys/vm/max_map_count"},
		"max number of memory map areas",
		map[string]int{
			"ECK":       262144,
			"Sonarqube": 524288,
		},
	},
	{
		"fs.file-max",
		[]string{"cat", "/proc/sys/fs/file-max"},
		"max number of file handles",
		map[string]int{
			"Sonarqube": 131072,
		},
	},
	{
		"ulimit -n",
		[]string{"ulimit", "-n"},
		"max number of open files",
		map[string]int{
			"Sonarqube": 131072,
		},
	},
	{
		"ulimit -u",
		[]string{"ulimit", "-u"},
		"max number of user processes",
		map[string]int{
			"Sonarqube": 8192,
		},
	},
}

// fluxControllerPods lists all the required pods that must be running in order to confirm Flux is installed in the cluster
var fluxControllerPods []string = []string{
	"helm-controller",
	"kustomize-controller",
	"source-controller",
	"notification-controller",
}

// NewPreflightCheckCmd - Creates a new Cobra command which implements the `bbctl preflight-check` functionality
func NewPreflightCheckCmd(factory bbUtil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     preflightCheckUse,
		Short:   preflightCheckShort,
		Long:    preflightCheckLong,
		Example: preflightCheckExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return bbPreflightCheck(cmd, factory, cmd, preflightChecks)
		},
	}

	configClient, clientError := factory.GetConfigClient(cmd)
	if clientError != nil {
		return nil, fmt.Errorf("Unable to get config client: %w", clientError)
	}

	registryServerError := configClient.SetAndBindFlag(
		"registryserver",
		"",
		"Image registry server url",
	)
	if registryServerError != nil {
		return nil, fmt.Errorf("Error setting registryserver flag: %w", registryServerError)
	}

	registryUserError := configClient.SetAndBindFlag(
		"registryusername",
		"",
		"Image registry username",
	)
	if registryUserError != nil {
		return nil, fmt.Errorf("Error setting registryusername flag: %w", registryUserError)
	}

	registryPasswordError := configClient.SetAndBindFlag(
		"registrypassword",
		"",
		"Image registry password",
	)
	if registryPasswordError != nil {
		return nil, fmt.Errorf("Error setting registrypassword flag: %w", registryPasswordError)
	}

	return cmd, nil
}

// Internal helper function to implement the preflight check command
//
// Runs the sequence of predefined checks in the preflightChecks array and summarizes the results
func bbPreflightCheck(cmd *cobra.Command, factory bbUtil.Factory, command *cobra.Command, preflightChecks []preflightCheck) error {
	configClient, err := factory.GetConfigClient(command)
	if err != nil {
		return fmt.Errorf("Unable to get config client: %w", err)
	}
	config, configErr := configClient.GetConfig()
	if configErr != nil {
		return fmt.Errorf("error getting config: %w", configErr)
	}
	for i, check := range preflightChecks {
		status := check.function(cmd, factory, config)
		preflightChecks[i].status = status
	}
	return printPreflightCheckSummary(factory, preflightChecks)
}

// Internal helper function to implement the metrics server check step
func checkMetricsServer(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
	streams := factory.GetIOStream()
	fmt.Fprintln(streams.Out, "Checking metrics server...")

	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	apiGroups, err := client.Discovery().ServerGroups()
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	metricsAPIAvailable := supportedMetricsAPIVersionAvailable(apiGroups)
	if !metricsAPIAvailable {
		fmt.Fprintln(streams.Out, "Check Failed - Metrics API not available.")
		return failed
	}

	fmt.Fprintln(streams.Out, "Check Passed - Metrics API available.")
	return passed
}

// Internal helper function to implement the storage class check step
func checkDefaultStorageClass(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
	streams := factory.GetIOStream()
	fmt.Fprintln(streams.Out, "Checking default storage class...")

	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	storageClasses, err := client.StorageV1().StorageClasses().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	defaultStorageClassName := ""
	for _, storageClass := range storageClasses.Items {
		if storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			defaultStorageClassName = storageClass.Name
		}
	}

	if defaultStorageClassName == "" {
		fmt.Fprintln(streams.Out, "Check Failed - Default storage class not found.")
		return failed
	}

	fmt.Fprintf(streams.Out, "Check Passed - Default storage class %s found.\n", defaultStorageClassName)
	return passed
}

// Internal helper function to implement the Flux installation check step
func checkFluxController(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
	streams := factory.GetIOStream()
	fmt.Fprintln(streams.Out, "Checking flux installation...")

	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	fluxStatus := make(map[string]string)
	pods, err := client.CoreV1().Pods(fluxNamespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	for _, pod := range pods.Items {
		fluxStatus[pod.Labels["app"]] = string(pod.Status.Phase)
	}

	status := passed
	for _, fluxPod := range fluxControllerPods {
		podStatus := fluxStatus[fluxPod]
		if podStatus != "" {
			if podStatus != string(coreV1.PodRunning) {
				fmt.Fprintf(streams.Out, "Check Failed - flux %s pod not in running state.\n", fluxPod)
				status = failed
			} else {
				fmt.Fprintf(streams.Out, "Check Passed - flux %s pod running.\n", fluxPod)
			}
		} else {
			fmt.Fprintf(streams.Out, "Check Failed - flux %s pod not found in %s namespace.\n", fluxPod, fluxNamespace)
			status = failed
		}
	}

	return status
}

// Internal helper function to implement the system parameters check step
func checkSystemParameters(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) preflightCheckStatus {
	streams := factory.GetIOStream()
	pod, err := createResourcesForCommandExecution(cmd, factory, config)
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	fmt.Fprintln(streams.Out, "Checking system parameters...")

	status := passed
	for _, param := range sysParams {
		fmt.Fprintf(streams.Out, "Checking %s\n", param.name)
		var stdout, stderr bytes.Buffer
		err := execCommand(cmd, factory, &stdout, &stderr, pod, param.command)
		if err != nil {
			fmt.Fprintf(streams.ErrOut, "%s", err)
			status = unknown
			continue
		}
		for bbPackage, threshold := range param.threshold {
			paramValue := strings.ReplaceAll(stdout.String(), "\n", "")
			if !checkSystemParameter(factory, bbPackage, param.name, paramValue, param.description, threshold) {
				status = failed
			}
		}
	}

	err = deleteResourcesForCommandExecution(cmd, factory)
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	return status
}

// Internal helper function to implement the individual system parameter checks
func checkSystemParameter(factory bbUtil.Factory, bbPackage string, param string, value string, _ string, threshold int) bool {
	streams := factory.GetIOStream()
	fmt.Fprintf(streams.Out, "%s = %s\n", param, value)

	if value == "unlimited" {
		fmt.Fprintf(streams.Out, "Check Passed - %s %s is suitable for %s to work.\n", param, value, bbPackage)
		return true
	}

	status := true

	val, err := strconv.Atoi(value)
	if err == nil {
		if val < threshold {
			fmt.Fprintf(streams.Out, "Check Failed - %s needs to be at least %d for %s to work.\n", param, threshold, bbPackage)
			status = false
		} else {
			fmt.Fprintf(streams.Out, "Check Passed - %s %d is suitable for %s to work.\n", param, val, bbPackage)
		}
	} else {
		fmt.Fprintf(streams.Out, "Check Undetermined - %s needs to be at least %d for %s to work. Current value %s\n", param, threshold, bbPackage, value)
		status = false
	}

	return status
}

// Internal helper function to verfiy the k8s metrics API version as part of the checkMetricsServer step
func supportedMetricsAPIVersionAvailable(discoveredAPIGroups *metaV1.APIGroupList) bool {
	supportedMetricsAPIVersions := []string{"v1beta1"}

	for _, discoveredAPIGroup := range discoveredAPIGroups.Groups {
		if discoveredAPIGroup.Name != metricsApi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}

	return false
}

// Internal helper function to create the preflight check job resources in the k8s cluster
//
// Create a new namespace, a container registry credentials secret, and the preflight check job
func createResourcesForCommandExecution(cmd *cobra.Command, factory bbUtil.Factory, config *schemas.GlobalConfiguration) (*coreV1.Pod, error) {
	streams := factory.GetIOStream()
	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return nil, err
	}

	err = createNamespaceForCommandExecution(client, streams.Out, config)
	if err != nil {
		return nil, err
	}

	secret, err := createRegistrySecretForCommandExecution(client, streams.Out, config)
	if err != nil {
		return nil, err
	}

	return createJobForCommandExecution(client, streams.Out, secret, config)
}

// Internal helper function to create a new namespace in the k8s cluster for the preflight check job
//
// Will attempt to delete the namespace and recreate it if it already exists
func createNamespaceForCommandExecution(client kubernetes.Interface, w io.Writer, config *schemas.GlobalConfiguration) error {
	fmt.Fprintln(w, "Creating namespace for command execution...")

	_, err := bbUtilK8s.CreateNamespace(client, preflightPodNamespace)
	if err != nil {
		if api_errors.IsAlreadyExists(err) {
			fmt.Fprintf(w, "Namespace %s already exists... It will be recreated\n", preflightPodNamespace)
			err = bbUtilK8s.DeleteNamespace(client, preflightPodNamespace)
			if err != nil {
				return err
			}
			// Give the namespace deletion some time to finish before trying to recreate the namespace
			for retry := 0; retry <= config.PreflightCheckConfiguration.RetryCount; retry++ {
				_, err = bbUtilK8s.CreateNamespace(client, preflightPodNamespace)
				if err != nil {
					time.Sleep(time.Duration(config.PreflightCheckConfiguration.RetryDelay) * time.Second)
				} else {
					break
				}
			}
		}
	}

	return err
}

// Internal helper function to create a new container registry credentials secret in the k8s cluster
func createRegistrySecretForCommandExecution(client kubernetes.Interface, w io.Writer, config *schemas.GlobalConfiguration) (*coreV1.Secret, error) {
	fmt.Fprintln(w, "Creating registry secret for command execution...")

	server := config.PreflightCheckConfiguration.RegistryServer
	username := config.PreflightCheckConfiguration.RegistryUsername
	password := config.PreflightCheckConfiguration.RegistryPassword

	if server == "" || username == "" || password == "" {
		return nil, errors.New("\n***Invalid registry credentials provided. Ensure the registry server, username, and password values are all set!***")
	}

	return bbUtilK8s.CreateRegistrySecret(client, preflightPodNamespace,
		preflightPodImagePullSecret, server, username, password)
}

// Internal helper function to create the preflight check job in the k8s cluster
func createJobForCommandExecution(client kubernetes.Interface, w io.Writer, secret *coreV1.Secret, config *schemas.GlobalConfiguration) (*coreV1.Pod, error) {
	fmt.Fprintln(w, "Creating job for command execution...")

	jobDesc := &bbUtilK8s.JobDesc{
		Name:               preflightPodName,
		ContainerName:      "executor",
		ContainerImage:     preflightPodImage,
		ImagePullSecret:    secret.Name,
		Command:            []string{"/bin/sleep"},
		Args:               []string{"30"},
		TTLSecondsOnFinish: 0,
	}

	job, err := bbUtilK8s.CreateJob(client, preflightPodNamespace, jobDesc)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(w, "Waiting for job %s to be ready...\n", job.Name)

	for i := 0; i < config.PreflightCheckConfiguration.RetryCount; i++ {
		pods, err := client.CoreV1().Pods(preflightPodNamespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: "job-name=preflightcheck"})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			if pod.Status.Phase == coreV1.PodRunning {
				return &pod, nil
			}
		}
		time.Sleep(time.Duration(config.PreflightCheckConfiguration.RetryDelay) * time.Second)
	}

	return nil, fmt.Errorf("timeout waiting for command execution job to be ready")
}

// Internal helper function to cleanup k8s resources after the system parameters check is complete
func deleteResourcesForCommandExecution(cmd *cobra.Command, factory bbUtil.Factory) error {
	streams := factory.GetIOStream()
	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return err
	}

	fmt.Fprintln(streams.Out, "Deleting namespace for command execution...")

	return bbUtilK8s.DeleteNamespace(client, preflightPodNamespace)
}

// Internal helper function to execute CLI commands in a Pod running on the k8s cluster to verify system parameter check values
func execCommand(cmd *cobra.Command, factory bbUtil.Factory, out io.Writer, errOut io.Writer, pod *coreV1.Pod, command []string) error {
	exec, err := factory.GetCommandExecutor(cmd, pod, "", command, out, errOut)
	if err != nil {
		return err
	}

	err = exec.StreamWithContext(context.TODO(), remoteCommand.StreamOptions{
		Stdin:  nil,
		Stdout: out,
		Stderr: errOut,
	})

	return err
}

// Internal helper function to print the results of every preflight check step out to the console
func printPreflightCheckSummary(factory bbUtil.Factory, preflightChecks []preflightCheck) error {
	streams := factory.GetIOStream()
	var errorsList []error
	_, err := fmt.Fprintf(streams.Out, "\n\nPreflight Check Summary\n\n")
	if err != nil {
		errorsList = append(errorsList, err)
	}

	for _, check := range preflightChecks {
		message := check.failureMessage
		if check.status == passed {
			message = check.successMessage
		} else if check.status == unknown {
			message = fmt.Sprintf("System Error - Execute command again to run %s", check.desc)
		}
		_, err := fmt.Fprintf(streams.Out, "%s %s...\n%s\n\n", check.desc, check.status, message)
		if err != nil {
			errorsList = append(errorsList, err)
		}
	}
	return errors.Join(errorsList...)
}
