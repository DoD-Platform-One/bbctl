// Package yamler implements a YAML marshaler with configurable indentation... we wanted to stick with two spaces.
package yamler

import (
	"bufio"
	"bytes"
	"io"
	"log"

	upstream "gopkg.in/yaml.v3"
)

// MinIndent is the lowest number of spaces we can indent a level of YAML output. This is an implementation detail
// of the underlying yaml.v3 library.
const MinIndent = 2

// MaxIndent is the highest number of spaces we can indent a level of YAML output. This is an implementation detail
// of the underlying yaml.v3 library.
const MaxIndent = 9

// DefaultIndent is 2 spaces because BigBang's YAML configs were already that way when this yamler package was written.
const DefaultIndent = 2

type Marshaler struct {
	indentCols int
}

// Marshal writes a YAML-annotated struct to a byte slice. Indentation width will be equal to DefaultIndent in spaces.
func Marshal(in interface{}) ([]byte, error) {
	marshaler := NewMarshaler(DefaultIndent)
	return marshaler.Marshal(in)
}

// Unmarshal loads a struct from a YAML input byte slice, provided
// the target struct is YAML-annotated.
// Wraps yaml.v3's Unmarshal() method in the interest of allowing
// YAMLer to serve as a drop-in replacement for current bbctl use cases.
func Unmarshal(in []byte, out interface{}) error {
	return upstream.Unmarshal(in, out)
}

// NewMarshaler creates a new Marshaler that can convert YAML-annotated inputs
// to YAML output having a specific indent width between 2 and 9.
//
// indentCols will be silently set to DefaultIndent if it falls beyond
// MinIndent or MaxIndent.
func NewMarshaler(indentCols int) *Marshaler {
	if indentCols < MinIndent || indentCols > MaxIndent {
		indentCols = DefaultIndent
	}
	return &Marshaler{indentCols: indentCols}
}

// Marshal writes a given YAML-annotated object out to a byte slice.
func (yd *Marshaler) Marshal(in interface{}) ([]byte, error) {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	encoder := upstream.NewEncoder(w)
	encoder.SetIndent(yd.indentCols)

	err := encoder.Encode(in)
	if err != nil {
		return nil, err
	}

	err = encoder.Close()
	if err != nil {
		log.Fatalf("failed to close encoder: %v", err)
	}
	err = w.Flush()
	if err != nil {
		log.Fatalf("failed to flush: %v", err)
	}

	return b.Bytes(), nil
}

// MarshalToWriter writes a given YAML-annotated object out to a writer such as io.Stdout.
func (yd *Marshaler) MarshalToWriter(in interface{}, w io.Writer) error {
	encoder := upstream.NewEncoder(w)
	encoder.SetIndent(yd.indentCols)

	err := encoder.Encode(in)
	if err != nil {
		return err
	}

	return encoder.Close()
}
