package apiwrappers

import (
	"bytes"
	"io"
	"testing"
)

// ReaderWriter is an interface that combines the io.Reader and io.Writer interfaces
type ReaderWriter interface {
	io.Reader
	io.Writer
}

// FakeReaderWriter is a fake implementation of io.Writer that can be used for testing purposes
type FakeReaderWriter struct {
	ActualBuffer       ReaderWriter
	shouldErrorOnRead  bool
	shouldErrorOnWrite bool
	t                  *testing.T
}

// CreateFakeReaderWriter creates a new FakeReaderWriter instance with a backing &bytes.Buffer
func CreateFakeReaderWriter(t *testing.T, shouldErrorOnRead bool, shouldErrorOnWrite bool) *FakeReaderWriter {
	t.Helper()
	return &FakeReaderWriter{
		ActualBuffer:       &bytes.Buffer{},
		shouldErrorOnRead:  shouldErrorOnRead,
		shouldErrorOnWrite: shouldErrorOnWrite,
		t:                  t,
	}
}

// CreateFakeWriterFromReaderWriter creates a new FakeReaderWriter instance from an existing ReaderWriter
func CreateFakeWriterFromReaderWriter(t *testing.T, shouldErrorOnRead bool, shouldErrorOnWrite bool, actualBuffer ReaderWriter) *FakeReaderWriter {
	t.Helper()
	return &FakeReaderWriter{
		ActualBuffer:       actualBuffer,
		shouldErrorOnRead:  shouldErrorOnRead,
		shouldErrorOnWrite: shouldErrorOnWrite,
		t:                  t,
	}
}

// Write writes the given byte slice to the buffer
func (f *FakeReaderWriter) Write(p []byte) (int, error) {
	if f.shouldErrorOnWrite || f.t == nil || f.ActualBuffer == nil {
		return 0, &FakeWriterError{badInitialization: f.t == nil || f.ActualBuffer == nil}
	}
	return f.ActualBuffer.Write(p)
}

// Read reads the given byte slice from the buffer
func (f *FakeReaderWriter) Read(p []byte) (int, error) {
	if f.shouldErrorOnRead || f.t == nil || f.ActualBuffer == nil {
		return 0, &FakeWriterError{badInitialization: f.t == nil || f.ActualBuffer == nil}
	}
	return f.ActualBuffer.Read(p)
}

// FakeWriterError is an error that is returned when the FakeWriter is intentionally errored
type FakeWriterError struct {
	badInitialization bool
}

// Error returns the error message string
func (f *FakeWriterError) Error() string {
	if f.badInitialization {
		return "FakeWriter not properly initialized"
	}
	return "FakeWriter intentionally errored"
}
