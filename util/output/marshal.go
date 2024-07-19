package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

// Outputable interface is used to define the methods that an object must implement in order to be able to be marshaled into different formats
type Outputable interface {
	// MashalYaml marshals the object into a yaml format
	MarshalYaml() ([]byte, error)
	// MarshalJson marshals the object into a json format
	MarshalJson() ([]byte, error)
	// MarshalHumanReadable marshals the object into a human readable format
	MarshalHumanReadable() (string, error)
	//
}

// BasicOutput is a simple struct that contains a name field that can be marshaled into different formats
//
// This struct is used to demonstrate how to implement the Outputable interface and can be embedded into other structs
type BasicOutput struct {
	Vals map[string]interface{}
}

// MashalYaml marshals the BasicOutput object into a yaml format
func (b *BasicOutput) MarshalYaml() ([]byte, error) {
	return yaml.Marshal(b.Vals)
}

// MarshalJson marshals the BasicOutput object into a json format
func (b *BasicOutput) MarshalJson() ([]byte, error) {
	return json.Marshal(b.Vals)
}

// MarshalHumanReadable marshals the BasicOutput object into a human readable format by calling the String method
func (b *BasicOutput) MarshalHumanReadable() (string, error) {
	return b.String(), nil
}

// String returns the name of the BasicOutput object in a human readable format
func (b *BasicOutput) String() string {
	return fmt.Sprintf("Vals: %s", b.Vals)
}
