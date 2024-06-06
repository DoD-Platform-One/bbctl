package apiwrappers

import (
	"bytes"
	"io"
	"testing"
)

// FakeWriter is a fake implementation of io.Writer that can be used for testing purposes.
type FakeWriter struct {
	ActualBuffer io.Writer
	ShouldError  bool
	t            *testing.T
}

// CreateFakeWriter creates a new FakeWriter instance.
func CreateFakeWriter(t *testing.T, shouldError bool) *FakeWriter {
	return &FakeWriter{
		ActualBuffer: &bytes.Buffer{},
		ShouldError:  shouldError,
		t:            t,
	}
}

// CreateFakeWriter creates a new FakeWriter instance.
func CreateFakeWriterFromStream(t *testing.T, shouldError bool, actualBuffer io.Writer) *FakeWriter {
	return &FakeWriter{
		ActualBuffer: actualBuffer,
		ShouldError:  shouldError,
		t:            t,
	}
}

// Write writes the given byte slice to the buffer.
func (f *FakeWriter) Write(p []byte) (n int, err error) {
	if f.ShouldError || f.t == nil {
		return 0, &FakeWriterError{badInitialization: f.t == nil}
	}
	return f.ActualBuffer.Write(p)
}

// FakeWriterError is an error that is returned when the FakeWriter is intentionally errored.
type FakeWriterError struct {
	badInitialization bool
}

// Error returns the error message.
func (f *FakeWriterError) Error() string {
	if f.badInitialization {
		return "FakeWriter not properly initialized"
	}
	return "FakeWriter intentionally errored"
}
