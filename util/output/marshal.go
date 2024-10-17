package output

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
)

// Outputable interface is used to define the methods that an object must implement in order to be able to be marshaled into different formats
type Outputable interface {
	// MashalYaml marshals the object into a yaml format
	EncodeYAML() ([]byte, error)
	// EncodeJSON marshals the object into a json format
	EncodeJSON() ([]byte, error)
	// EncodeText marshals the object into a human readable format
	EncodeText() ([]byte, error)
	//
}

// BasicOutput is a simple struct that contains a name field that can be marshaled into different formats
//
// This struct is used to demonstrate how to implement the Outputable interface and can be embedded into other structs
type BasicOutput struct {
	Vals map[string]interface{}
}

// MashalYaml marshals the BasicOutput object into a yaml format
func (b *BasicOutput) EncodeYAML() ([]byte, error) {
	return yaml.Marshal(b.Vals)
}

// EncodeJSON marshals the BasicOutput object into a json format
func (b *BasicOutput) EncodeJSON() ([]byte, error) {
	return json.Marshal(b.Vals)
}

// EncodeText marshals the BasicOutput object into a human readable format by calling the String method
func (b *BasicOutput) EncodeText() ([]byte, error) {
	return []byte(b.String()), nil
}

// String returns the name of the BasicOutput object in a human readable format
func (b *BasicOutput) String() string {
	return fmt.Sprintf("Vals: %s", b.Vals)
}
