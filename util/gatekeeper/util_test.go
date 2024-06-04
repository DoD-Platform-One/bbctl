package gatekeeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	bbTestUtil "repo1.dso.mil/big-bang/product/packages/bbctl/util/test"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestFetchGatekeeperCrds(t *testing.T) {
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{crdList(t)})
	factory.SetGVRToListKind(gvrToListKind(t))
	client, _ := factory.GetK8sDynamicClient(nil)
	crds, _ := FetchGatekeeperCrds(client)

	assert.Equal(t, "foos.constraints.gatekeeper.sh", crds.Items[0].GetName())
}

func TestFetchGatekeeperConstraints(t *testing.T) {
	var tests = []struct {
		desc     string
		arg      string
		expected []string
		objects  []runtime.Object
	}{
		{
			"no constraints exist",
			"foos.constraints.gatekeeper.sh",
			[]string{},
			[]runtime.Object{crdList(t)},
		},
		{
			"constraints exist",
			"foos.constraints.gatekeeper.sh",
			[]string{"foo-1", "foo-2"},
			[]runtime.Object{crdList(t), constraintList(t)},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			factory := bbTestUtil.GetFakeFactory()
			factory.SetObjects(test.objects)
			factory.SetGVRToListKind(gvrToListKind(t))
			client, _ := factory.GetK8sDynamicClient(nil)
			constraints, _ := FetchGatekeeperConstraints(client, test.arg)
			for i, constraint := range constraints.Items {
				assert.Equal(t, test.expected[i], constraint.GetName())
			}
		})
	}
}

func TestFetchGatekeeperConstraintsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{crdList(t)})
	factory.SetGVRToListKind(gvrToListKind(t))
	client := &badClient{}

	// Act
	result, err := FetchGatekeeperConstraints(client, "nop.constraints.gatekeeper.sh")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting gatekeeper constraint")
	assert.Nil(t, result)
}

func TestFetchGatekeeperCrdsError(t *testing.T) {
	// Arrange
	factory := bbTestUtil.GetFakeFactory()
	factory.SetObjects([]runtime.Object{constraintList(t)})
	factory.SetGVRToListKind(gvrToListKind(t))
	client := &badClient{}

	// Act
	result, err := FetchGatekeeperCrds(client)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error getting gatekeeper crds")
	assert.Nil(t, result)
}
