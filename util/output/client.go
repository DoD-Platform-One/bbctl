// Package output provides utilities for outputting data in different formats such as JSON, YAML, and text.
// It includes functionality for creating an output client that writes data to an io.Writer based on a specified format.
package output

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	genericIOOptions "k8s.io/cli-runtime/pkg/genericiooptions"
)

// Format defines constants for supported output formats.
type Format string

const (
	JSON Format = "json"
	TEXT Format = "text"
	YAML Format = "yaml"
)

// Client is the interface that wraps the basic Output method.
//
// Output takes an Outputable object and writes it to an io.Writer based on a specified format.
type Client interface {
	Output(data Outputable) error
}

// outputClient is the implementation of the Client interface.
// It manages the output formatting and handles writing data to an io.Writer.
type outputClient struct {
	Format Format    // The output format (JSON, YAML, TEXT)
	Writer io.Writer // The writer to output the data
}

// NewOutputClient creates a new output client based on the specified format and io streams.
//
// format: The desired output format (JSON, YAML, TEXT)
// streams: The generic I/O streams for input/output operations.
func NewOutputClient(format Format, streams genericIOOptions.IOStreams) Client {
	return &outputClient{
		Format: format,
		Writer: streams.Out,
	}
}

// Output writes the given data to the client's io.Writer in the specified output format.
//
// data: The data to be outputted, which must implement the Outputable interface.
func (oc *outputClient) Output(data Outputable) error {
	switch oc.Format {
	case TEXT:
		return oc.writeText(data)
	case JSON:
		return oc.writeJSON(data)
	case YAML:
		return oc.writeYAML(data)
	default:
		return fmt.Errorf("unsupported format: %s", oc.Format)
	}
}

// WriteJson writes the given data as JSON to the client's io.Writer.
//
// data: The data to be written, which must implement the Outputable interface.
func (oc *outputClient) writeJSON(data Outputable) error {
	jsonData, err := data.EncodeJSON()
	if err != nil {
		return errors.Wrap(err, "unable to write JSON output")
	}

	_, err = oc.Writer.Write(jsonData)
	if err != nil {
		return errors.Wrap(err, "unable to write JSON output")
	}

	return nil
}

// WriteYaml writes the given data as YAML to the client's io.Writer.
//
// data: The data to be written, which must implement the Outputable interface.
func (oc *outputClient) writeYAML(data Outputable) error {
	yamlData, err := data.EncodeYAML()
	if err != nil {
		return errors.Wrap(err, "unable to write YAML output")
	}

	_, err = oc.Writer.Write(yamlData)
	if err != nil {
		return errors.Wrap(err, "unable to write YAML output")
	}

	return nil
}

// WriteText writes the given data as human-readable text to the client's io.Writer.
//
// data: The data to be written, which must implement the Outputable interface.
func (oc *outputClient) writeText(data Outputable) error {
	output, err := data.EncodeText()
	if err != nil {
		return errors.Wrap(err, "unable to write human-readable output")
	}

	_, err = fmt.Fprintln(oc.Writer, string(output))
	if err != nil {
		return errors.Wrap(err, "unable to write human-readable output")
	}

	return nil
}
