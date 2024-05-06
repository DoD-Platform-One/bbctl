package cmd

import (
	"bytes"
	"context"
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
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	remoteCommand "k8s.io/client-go/tools/remotecommand"
	cmdUtil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	metricsApi "k8s.io/metrics/pkg/apis/metrics"
)

var (
	preflightCheckUse = `preflight-check`

	preflightCheckShort = i18n.T(`Check cluster for expected configuration before installing big bang.`)

	preflightCheckLong = templates.LongDesc(i18n.T(`
		Check cluster for expected configuration before installing big bang.
		This command creates a job in preflight-check namespace to check system parameters.
		User needs to have RBAC permissions to create and delete namespace, secret and job resources.`))

	preflightCheckExample = templates.Examples(i18n.T(`
		# Check cluster for expected configuration
		bbctl preflight-check --registryserver <registry-server> --registryusername <username> --registrypassword <password>
		# Check cluster for expected configuration using environment variables for registry access
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

type preflightCheckFunc func(*cobra.Command, bbUtil.Factory, genericIOOptions.IOStreams, *schemas.GlobalConfiguration) preflightCheckStatus

type preflightCheckStatus string

const (
	failed  preflightCheckStatus = "Failed"  // check failed
	passed  preflightCheckStatus = "Passed"  // check passed
	unknown preflightCheckStatus = "Unknown" // check execution error
)

type preflightCheck struct {
	desc           string               // check description
	function       preflightCheckFunc   // function with check logic
	status         preflightCheckStatus // function execution status
	failureMessage string               // user friendly failure message
	successMessage string               // user friendly success message
}

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
			For more information refer to https://repo1.dso.mil/platform-one/big-bang/bigbang/-/blob/master/docs/guides/prerequisites/os_preconfiguration.md`)),
		successMessage: templates.LongDesc(i18n.T(`
			System parameters determined to be equal or greater than the recommended value. 
			This will allow for successful installation of packages that passed the check.
			For more information refer to https://repo1.dso.mil/platform-one/big-bang/bigbang/-/blob/master/docs/guides/prerequisites/os_preconfiguration.md`)),
	},
}

// system parameters
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

// system parameters
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

// pods created by flux controller
var fluxControllerPods []string = []string{
	"helm-controller",
	"kustomize-controller",
	"source-controller",
	"notification-controller",
}

// NewPreflightCheckCmd - new preflight check command
func NewPreflightCheckCmd(factory bbUtil.Factory, streams genericIOOptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     preflightCheckUse,
		Short:   preflightCheckShort,
		Long:    preflightCheckLong,
		Example: preflightCheckExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdUtil.CheckErr(bbPreflightCheck(cmd, factory, streams, cmd, preflightChecks))
		},
	}

	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(cmd)
	loggingClient.HandleError("error getting config client", err)

	loggingClient.HandleError(
		"error setting registryserver flag: %v",
		configClient.SetAndBindFlag(
			"registryserver",
			"",
			"Image registry server url",
		),
	)
	loggingClient.HandleError(
		"error setting registryusername flag: %v",
		configClient.SetAndBindFlag(
			"registryusername",
			"",
			"Image registry username",
		),
	)
	loggingClient.HandleError(
		"error setting registrypassword flag: %v",
		configClient.SetAndBindFlag(
			"registrypassword",
			"",
			"Image registry password",
		),
	)

	return cmd
}

// run sequence of predefined checks and summarize results
func bbPreflightCheck(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, command *cobra.Command, preflightChecks []preflightCheck) error {
	loggingClient := factory.GetLoggingClient()
	configClient, err := factory.GetConfigClient(command)
	loggingClient.HandleError("error getting config client", err)
	config := configClient.GetConfig()
	for i, check := range preflightChecks {
		status := check.function(cmd, factory, streams, config)
		preflightChecks[i].status = status
	}
	printPreflightCheckSummary(streams, preflightChecks)
	return nil
}

func checkMetricsServer(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
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

func checkDefaultStorageClass(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
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

func checkFluxController(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
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

func checkSystemParameters(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) preflightCheckStatus {
	pod, err := createResourcesForCommandExecution(cmd, factory, streams, config)
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
			if !checkSystemParameter(streams, bbPackage, param.name, paramValue, param.description, threshold) {
				status = failed
			}
		}
	}

	err = deleteResourcesForCommandExecution(cmd, factory, streams)
	if err != nil {
		fmt.Fprintf(streams.ErrOut, "%s", err.Error())
		return unknown
	}

	return status
}

func checkSystemParameter(streams genericIOOptions.IOStreams, bbPackage string, param string, value string, _ string, threshold int) bool {
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
	}

	return status

}

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

func createResourcesForCommandExecution(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams, config *schemas.GlobalConfiguration) (*coreV1.Pod, error) {
	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return nil, err
	}

	err = createNamespaceForCommandExecution(client, streams.Out)
	if err != nil {
		return nil, err
	}

	secret, err := createRegistrySecretForCommandExecution(client, streams.Out, config)
	if err != nil {
		return nil, err
	}

	return createJobForCommandExecution(client, streams.Out, secret)

}

func createNamespaceForCommandExecution(client kubernetes.Interface, w io.Writer) error {
	fmt.Fprintln(w, "Creating namespace for command execution...")

	_, err := bbUtilK8s.CreateNamespace(client, preflightPodNamespace)
	if err != nil {
		if api_errors.IsAlreadyExists(err) {
			fmt.Fprintf(w, "Namespace %s already exists...It will be recreated\n", preflightPodNamespace)
			err = bbUtilK8s.DeleteNamespace(client, preflightPodNamespace)
			if err != nil {
				return err
			}
			_, err = bbUtilK8s.CreateNamespace(client, preflightPodNamespace)
		}
	}

	return err
}

func createRegistrySecretForCommandExecution(client kubernetes.Interface, w io.Writer, config *schemas.GlobalConfiguration) (*coreV1.Secret, error) {
	fmt.Fprintln(w, "Creating registry secret for command execution...")

	server := config.PreflightCheckConfiguration.RegistryServer
	username := config.PreflightCheckConfiguration.RegistryUsername
	password := config.PreflightCheckConfiguration.RegistryPassword

	return bbUtilK8s.CreateRegistrySecret(client, preflightPodNamespace,
		preflightPodImagePullSecret, server, username, password)
}

func createJobForCommandExecution(client kubernetes.Interface, w io.Writer, secret *coreV1.Secret) (*coreV1.Pod, error) {
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

	for i := 0; i < 10; i++ {
		pods, _ := client.CoreV1().Pods(preflightPodNamespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: "job-name=preflightcheck"})
		for _, pod := range pods.Items {
			if pod.Status.Phase == coreV1.PodRunning {
				return &pod, nil
			}
		}
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("timeout waiting for command execution job to be ready")
}

func deleteResourcesForCommandExecution(cmd *cobra.Command, factory bbUtil.Factory, streams genericIOOptions.IOStreams) error {
	client, err := factory.GetK8sClientset(cmd)
	if err != nil {
		return err
	}

	fmt.Fprintln(streams.Out, "Deleting namespace for command execution...")

	return bbUtilK8s.DeleteNamespace(client, preflightPodNamespace)
}

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

func printPreflightCheckSummary(streams genericIOOptions.IOStreams, preflightChecks []preflightCheck) {
	fmt.Fprintf(streams.Out, "\n\n***Preflight Check Summary***\n\n")

	for _, check := range preflightChecks {
		message := check.failureMessage
		if check.status == passed {
			message = check.successMessage
		} else if check.status == unknown {
			message = fmt.Sprintf("System Error - Execute command again to run %s", check.desc)
		}
		fmt.Fprintf(streams.Out, "%s %s...\n%s\n\n", check.desc, check.status, message)
	}
}
