package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	bbUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util"
	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func policiesCmd(factory bbUtil.Factory, args []string) *cobra.Command {
	cmd, _ := NewPoliciesCmd(factory)
	cmd.SetArgs(args)
	return cmd
}

func TestGetPolicyCmdConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetFail.GetConfigClient = true
	// Act
	cmd, err := NewPoliciesCmd(factory)
	// Assert
	assert.Nil(t, cmd)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "Unable to get config client:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestGetPolicyUsage(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd := policiesCmd(factory, []string{})
	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "policy --PROVIDER CONSTRAINT_NAME", cmd.Use)
	assert.Contains(t, cmd.Example, "bbctl policy --gatekeeper")
	assert.Contains(t, cmd.Example, "bbctl policy --gatekeeper restrictedtainttoleration")
	assert.Contains(t, cmd.Example, "bbctl policy --kyverno")
	assert.Contains(t, cmd.Example, "bbctl policy --kyverno restrict-seccomp")
}

func TestInvalidArgsFunction(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	args := []string{"test"}
	// Act
	cmd := policiesCmd(factory, []string{})
	res, _ := cmd.ValidArgsFunction(cmd, args, "")
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
}

func gvrToListKindForPolicies() map[schema.GroupVersionResource]string {
	return map[schema.GroupVersionResource]string{
		{
			Group:    "apiextensions.k8s.io",
			Version:  "v1",
			Resource: "customresourcedefinitions",
		}: "customresourcedefinitionsList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "foos",
		}: "gkPolicyList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "foo",
		}: "gkPolicyList",
		{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: "nop",
		}: "gkPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "foos",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "foo",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "bars",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "bar",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1",
			Resource: "nop",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1beta1",
			Resource: "foos",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1beta1",
			Resource: "foo",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1beta1",
			Resource: "bars",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1beta1",
			Resource: "bar",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1beta1",
			Resource: "nop",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v2beta1",
			Resource: "foos",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v2beta1",
			Resource: "foo",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v2beta1",
			Resource: "bars",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v2beta1",
			Resource: "bar",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v2beta1",
			Resource: "nop",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1alpha2",
			Resource: "foos",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1alpha2",
			Resource: "foo",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1alpha2",
			Resource: "bars",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1alpha2",
			Resource: "bar",
		}: "kyvernoPolicyList",
		{
			Group:    "kyverno.io",
			Version:  "v1alpha2",
			Resource: "nop",
		}: "kyvernoPolicyList",
	}
}

func TestMatchingPolicyConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	// Act
	cmd, _ := NewPoliciesCmd(factory)
	factory.SetFail.GetConfigClient = true
	res, _ := cmd.ValidArgsFunction(cmd, []string{}, "")
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
}

func TestNoMatchingPrefix(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.GetViper().Set("big-bang-repo", "test")
	// Act
	cmd := policiesCmd(factory, []string{})
	res, _ := cmd.ValidArgsFunction(cmd, []string{}, "")
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
}

func TestGetK8sDynamicClientErrorGatekeeper(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("gatekeeper", true)
	factory.SetFail.GetK8sDynamicClient = true
	// Act
	cmd := policiesCmd(factory, []string{})
	res, _ := cmd.ValidArgsFunction(cmd, []string{}, "")
	err1 := cmd.RunE(cmd, []string{})
	err2 := cmd.RunE(cmd, []string{""})
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "failed to get K8sDynamicClient client") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "failed to get K8sDynamicClient client") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
}

func TestGetConfigClientError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	cmd := policiesCmd(factory, []string{})
	// Act
	factory.SetFail.GetConfigClient = true
	err1 := cmd.RunE(cmd, []string{})
	err2 := cmd.RunE(cmd, []string{""})
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "Unable to get config client:") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "Unable to get config client:") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
}

func TestGetK8sDynamicClientErrorKyverno(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kyverno", true)
	factory.SetFail.GetK8sDynamicClient = true
	// Act
	cmd := policiesCmd(factory, []string{})
	res, _ := cmd.ValidArgsFunction(cmd, []string{}, "")
	err1 := cmd.RunE(cmd, []string{})
	err2 := cmd.RunE(cmd, []string{""})
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "failed to get K8sDynamicClient client") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "failed to get K8sDynamicClient client") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
}

func TestNoPolicySpecified(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	// Act
	cmd := policiesCmd(factory, []string{})
	err1 := cmd.RunE(cmd, []string{})
	err2 := cmd.RunE(cmd, []string{""})
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "either --gatekeeper or --kyverno must be specified") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "either --gatekeeper or --kyverno must be specified, but not both") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
}

func TestFetchGatekeeperCrdsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("gatekeeper", true)
	factory.SetFail.GetPolicyClient = true
	factory.SetFail.GetCrds = true
	// Act
	cmd := policiesCmd(factory, []string{})
	res, _ := cmd.ValidArgsFunction(cmd, []string{}, "")
	err := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
	assert.Error(t, err)
	if !assert.Contains(t, err.Error(), "error getting gatekeeper crds:") {
		t.Errorf("unexpected output: %s", err.Error())
	}
}

func TestFetchGatekeeperConstraintsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("gatekeeper", true)
	factory.SetFail.GetPolicyClient = true
	// Act
	cmd := policiesCmd(factory, []string{})
	err1 := cmd.RunE(cmd, []string{""})
	err2 := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "error getting gatekeeper constraint:") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "error getting gatekeeper constraint:") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
}

func TestFetchKyvernoCrdsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kyverno", true)
	factory.SetFail.GetPolicyClient = true
	factory.SetFail.GetCrds = true
	// Act
	cmd := policiesCmd(factory, []string{})
	res, _ := cmd.ValidArgsFunction(cmd, []string{}, "")
	err1 := cmd.RunE(cmd, []string{})
	err2 := cmd.RunE(cmd, []string{""})
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "error getting kyverno crds:") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "error getting kyverno crds:") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
}

func TestFetchKyvernoPoliciesError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kyverno", true)
	factory.SetFail.GetPolicyClient = true
	// Act
	cmd := policiesCmd(factory, []string{})
	res, _ := cmd.ValidArgsFunction(cmd, []string{}, "")
	err1 := cmd.RunE(cmd, []string{""})
	err2 := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	assert.Nil(t, res)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "error getting kyverno policies:") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "error getting kyverno policies:") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
}

func TestFetchGatekeeperPolicyDescriptorError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("gatekeeper", true)
	factory.SetFail.GetPolicyClient = true
	factory.SetFail.GetDescriptor = true
	factory.SetFail.DescriptorType = "kind"

	// Act
	cmd := policiesCmd(factory, []string{})
	err1 := cmd.RunE(cmd, []string{""})
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "kind accessor error") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
}

func TestFetchGatekeeperPolicyDescriptorStringErrors(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("gatekeeper", true)
	factory.SetFail.GetPolicyClient = true
	factory.SetFail.GetDescriptor = true
	// Act
	cmd := policiesCmd(factory, []string{})
	factory.SetFail.DescriptorType = "name"
	err1 := cmd.RunE(cmd, []string{})
	factory.SetFail.DescriptorType = "desc"
	err2 := cmd.RunE(cmd, []string{})
	factory.SetFail.DescriptorType = "action"
	err3 := cmd.RunE(cmd, []string{})
	factory.SetFail.DescriptorType = "kind"
	err4 := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	//assert.Error(t, err)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "name accessor error") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "description accessor error") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
	assert.Error(t, err3)
	if !assert.Contains(t, err3.Error(), "Action accessor error") {
		t.Errorf("unexpected output: %s", err3.Error())
	}
	assert.Error(t, err4)
	if !assert.Contains(t, err4.Error(), "kind accessor error") {
		t.Errorf("unexpected output: %s", err4.Error())
	}
}

func TestFetchKyvernoPolicyDescriptorError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kyverno", true)
	factory.SetFail.GetPolicyClient = true
	factory.SetFail.GetDescriptor = true
	factory.SetFail.DescriptorType = "kind"
	// Act
	cmd := policiesCmd(factory, []string{})
	err1 := cmd.RunE(cmd, []string{"foo-1"})
	// Assert
	assert.NotNil(t, cmd)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "kind accessor error") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
}

func TestFetchKyvernoPolicyDescriptorStringErrors(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	viperInstance := factory.GetViper()
	viperInstance.Set("big-bang-repo", "test")
	viperInstance.Set("kyverno", true)
	factory.SetFail.GetPolicyClient = true
	factory.SetFail.GetDescriptor = true
	// Act
	cmd := policiesCmd(factory, []string{})
	factory.SetFail.DescriptorType = "name"
	err1 := cmd.RunE(cmd, []string{})
	factory.SetFail.DescriptorType = "namespace"
	err2 := cmd.RunE(cmd, []string{})
	factory.SetFail.DescriptorType = "desc"
	err3 := cmd.RunE(cmd, []string{})
	factory.SetFail.DescriptorType = "action"
	err4 := cmd.RunE(cmd, []string{})
	factory.SetFail.DescriptorType = "kind"
	err5 := cmd.RunE(cmd, []string{})
	// Assert
	assert.NotNil(t, cmd)
	//assert.Error(t, err)
	assert.Error(t, err1)
	if !assert.Contains(t, err1.Error(), "name accessor error") {
		t.Errorf("unexpected output: %s", err1.Error())
	}
	assert.Error(t, err2)
	if !assert.Contains(t, err2.Error(), "namespace accessor error") {
		t.Errorf("unexpected output: %s", err2.Error())
	}
	assert.Error(t, err3)
	if !assert.Contains(t, err3.Error(), "description accessor error") {
		t.Errorf("unexpected output: %s", err3.Error())
	}
	assert.Error(t, err4)
	if !assert.Contains(t, err4.Error(), "Action accessor error") {
		t.Errorf("unexpected output: %s", err4.Error())
	}
	assert.Error(t, err5)
	if !assert.Contains(t, err5.Error(), "kind accessor error") {
		t.Errorf("unexpected output: %s", err5.Error())
	}
}

func TestGatekeeperPolicies(t *testing.T) {
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.constraints.gatekeeper.sh",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd},
	}

	constraint1 := &unstructured.Unstructured{}
	constraint1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "constraints.gatekeeper.sh/v1beta1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "foo-1",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
			"annotations": map[string]interface{}{
				"constraints.gatekeeper/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"enforcementAction": "deny",
		},
	})

	constraint2 := &unstructured.Unstructured{}
	constraint2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "constraints.gatekeeper.sh/v1beta1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "foo-2",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
			"annotations": map[string]interface{}{
				"constraints.gatekeeper/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"enforcementAction": "dryrun",
		},
	})

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "gkPolicyList",
		},
		Items: []unstructured.Unstructured{*constraint1, *constraint2},
	}

	var tests = []struct {
		desc     string
		args     []string
		expected []string
		objects  []runtime.Object
	}{
		{
			"list all policies",
			[]string{"--gatekeeper"},
			[]string{"foos.constraints.gatekeeper.sh", "deny", "invalid config"},
			[]runtime.Object{crdList, constraintList},
		},
		{
			"list policy with given name",
			[]string{"--gatekeeper", "foos.constraints.gatekeeper.sh"},
			[]string{"foos.constraints.gatekeeper.sh", "foo-1", "foo-2", "deny", "dry", "invalid config"},
			[]runtime.Object{crdList, constraintList},
		},
		{
			"list non existent policy",
			[]string{"--gatekeeper", "nop"},
			[]string{"No constraints found"},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForPolicies())
			factory.GetViper().Set("big-bang-repo", "test")
			streams := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			cmd := policiesCmd(factory, test.args)
			err := cmd.Execute()
			assert.NoError(t, err)
			for _, exp := range test.expected {
				if !strings.Contains(buf.String(), exp) {
					t.Errorf("unexpected output: %s", buf.String())
				}
			}
		})
	}
}

func TestNoGatekeeperPolicies(t *testing.T) {
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.constraints.gatekeeper.sh",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
	})

	crdList1 := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd},
	}

	crdList2 := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{},
	}

	constraintList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "constraints.gatekeeper.sh/v1beta1",
			"kind":       "gkPolicyList",
		},
		Items: []unstructured.Unstructured{},
	}

	var tests = []struct {
		desc     string
		args     []string
		expected []string
		objects  []runtime.Object
	}{
		{
			"No Constraints",
			[]string{"--gatekeeper"},
			[]string{"No constraints found"},
			[]runtime.Object{crdList1, constraintList},
		},
		{
			"No Crds",
			[]string{"--gatekeeper"},
			[]string{"No Gatekeeper Policies Found"},
			[]runtime.Object{crdList2, constraintList},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForPolicies())
			factory.GetViper().Set("big-bang-repo", "test")
			streams := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			cmd := policiesCmd(factory, test.args)
			err := cmd.Execute()
			assert.NoError(t, err)
			for _, exp := range test.expected {
				if !strings.Contains(buf.String(), exp) {
					t.Errorf("unexpected output: %s", buf.String())
				}
			}
		})
	}
}

func TestKyvernoPolicies(t *testing.T) {
	crd1 := &unstructured.Unstructured{}
	crd1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.policies.kyverno.io",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
		},
		"spec": map[string]any{
			"group": "kyverno.io",
		},
	})

	crd2 := &unstructured.Unstructured{}
	crd2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "bars.policies.kyverno.io",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
		},
		"spec": map[string]any{
			"group": "kyverno.io",
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd1, *crd2},
	}

	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "foo-1",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
			"group":                   "kyverno.io",
		},
	})

	policy2 := &unstructured.Unstructured{}
	policy2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "bar",
		"metadata": map[string]interface{}{
			"name":      "bar-1",
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "audit",
			"group":                   "kyverno.io",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1, *policy2},
	}

	var tests = []struct {
		desc     string
		args     []string
		expected []string
		objects  []runtime.Object
	}{
		{
			"list all policies",
			[]string{"--kyverno"},
			[]string{"foos.policies.kyverno.io", "enforce", "invalid config"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"list policy with given name",
			[]string{"--kyverno", "bar-1"},
			[]string{"bar", "bar-1", "demo", "audit", "invalid config"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"list non existent policy",
			[]string{"--kyverno", "nop"},
			[]string{"No Matching Policy Found"},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForPolicies())
			factory.GetViper().Set("big-bang-repo", "test")
			streams := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			cmd := policiesCmd(factory, test.args)
			err := cmd.Execute()
			assert.NoError(t, err)
			for _, exp := range test.expected {
				if !strings.Contains(buf.String(), exp) {
					t.Errorf("unexpected output: %s", buf.String())
				}
			}
		})
	}
}

func TestNoKyvernoPolicies(t *testing.T) {
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.policies.kyverno.io",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
		},
		"spec": map[string]any{
			"group": "kyverno.io",
		},
	})

	crdList1 := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd},
	}

	crdList2 := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{},
	}

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{},
	}

	var tests = []struct {
		desc     string
		args     []string
		expected []string
		objects  []runtime.Object
	}{
		{
			"No Policies",
			[]string{"--kyverno"},
			[]string{"No policies found"},
			[]runtime.Object{crdList1, policyList},
		},
		{
			"No Crds",
			[]string{"--kyverno"},
			[]string{"No Kyverno Policies Found"},
			[]runtime.Object{crdList2, policyList},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.ResetIOStream()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForPolicies())
			factory.GetViper().Set("big-bang-repo", "test")
			streams := factory.GetIOStream()
			buf := streams.Out.(*bytes.Buffer)
			cmd := policiesCmd(factory, test.args)
			err := cmd.Execute()
			assert.NoError(t, err)
			for _, exp := range test.expected {
				if !strings.Contains(buf.String(), exp) {
					t.Errorf("unexpected output: %s", buf.String())
				}
			}
		})
	}
}

func TestGatekeeperPoliciesCompletion(t *testing.T) {
	crd1 := &unstructured.Unstructured{}
	crd1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.constraints.gatekeeper.sh",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
	})

	crd2 := &unstructured.Unstructured{}
	crd2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "fudges.constraints.gatekeeper.sh",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "gatekeeper",
			},
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd1, *crd2},
	}

	var tests = []struct {
		desc     string
		hint     string
		expected []string
		objects  []runtime.Object
	}{
		{
			"match all policies",
			"",
			[]string{"foos", "fudges"},
			[]runtime.Object{crdList},
		},
		{
			"match policies with given prefix",
			"f",
			[]string{"foos", "fudges"},
			[]runtime.Object{crdList},
		},
		{
			"match policy with given prefix",
			"fud",
			[]string{"fudges"},
			[]runtime.Object{crdList},
		},
		{
			"match no policy",
			"z",
			[]string{},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForPolicies())
			viperInstance := factory.GetViper()
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("gatekeeper", true)
			cmd, _ := NewPoliciesCmd(factory)
			suggestions, _ := cmd.ValidArgsFunction(cmd, []string{}, test.hint)
			if !reflect.DeepEqual(test.expected, suggestions) {
				t.Fatalf("expected: %v, got: %v", test.expected, suggestions)
			}
		})
	}
}

func TestKyvernoPoliciesCompletion(t *testing.T) {

	crd1 := &unstructured.Unstructured{}
	crd1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "customresourcedefinition",
		"metadata": map[string]interface{}{
			"name": "foos.policies.kyverno.io",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
		},
		"spec": map[string]any{
			"group": "kyverno.io",
		},
	})

	crdList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "customresourcedefinitionList",
		},
		Items: []unstructured.Unstructured{*crd1},
	}

	policy1 := &unstructured.Unstructured{}
	policy1.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name": "fu-bar",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "enforce",
			"group":                   "kyverno.io",
		},
	})

	policy2 := &unstructured.Unstructured{}
	policy2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kyverno.io/v1",
		"kind":       "foo",
		"metadata": map[string]interface{}{
			"name":      "fudge-bar",
			"namespace": "demo",
			"labels": map[string]interface{}{
				"app.kubernetes.io/name": "kyverno",
			},
			"annotations": map[string]interface{}{
				"policies.kyverno.io/description": "invalid config",
			},
		},
		"spec": map[string]interface{}{
			"validationFailureAction": "audit",
			"group":                   "kyverno.io",
		},
	})

	policyList := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "kyverno.io/v1",
			"kind":       "kyvernoPolicyList",
		},
		Items: []unstructured.Unstructured{*policy1, *policy2},
	}

	var tests = []struct {
		desc     string
		hint     string
		expected []string
		objects  []runtime.Object
	}{
		{
			"match all policies",
			"",
			[]string{"fu-bar", "fudge-bar"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"match policies with given prefix",
			"fu",
			[]string{"fu-bar", "fudge-bar"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"match policy with given prefix",
			"fud",
			[]string{"fudge-bar"},
			[]runtime.Object{crdList, policyList},
		},
		{
			"match no policy",
			"z",
			[]string{},
			[]runtime.Object{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKindForPolicies())
			viperInstance := factory.GetViper()
			viperInstance.Set("big-bang-repo", "test")
			viperInstance.Set("kyverno", true)
			cmd, _ := NewPoliciesCmd(factory)
			suggestions, _ := cmd.ValidArgsFunction(cmd, []string{}, test.hint)
			if !reflect.DeepEqual(test.expected, suggestions) {
				t.Fatalf("expected: %v, got: %v", test.expected, suggestions)
			}
		})
	}
}
