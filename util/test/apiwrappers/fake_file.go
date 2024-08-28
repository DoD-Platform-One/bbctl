package apiwrappers

import (
	"io"
	"io/fs"
	"os"
	"syscall"
	"testing"
	"time"

	commonInterfaces "repo1.dso.mil/big-bang/product/packages/bbctl/util/common_interfaces"
)

// FakeFile is a fake implementation of os.File that can be used for testing purposes
type FakeFile struct {
	File               commonInterfaces.FileLike
	shouldErrorOnRead  bool
	shouldErrorOnWrite bool
	t                  *testing.T
	SetFail            struct {
		Chdir            bool
		Chmod            bool
		Chown            bool
		Close            bool
		Fd               bool
		Name             bool
		Read             bool
		ReadAt           bool
		ReadDir          bool
		ReadFrom         bool
		Readdir          bool
		Readdirnames     bool
		Seek             bool
		SetDeadline      bool
		SetReadDeadline  bool
		SetWriteDeadline bool
		Stat             bool
		Sync             bool
		SyscallConn      bool
		Truncate         bool
		Write            bool
		WriteAt          bool
		WriteString      bool
		WriteTo          bool
	}
}

// CreateFakeFileFromOSPipe creates a new FakeFile instance from a call to os.Pipe()
// the failOnRead and failOnWrite parameters determine if the file should error on read or write
// r will only error on read, w will only error on write (even if you set both to fail)
//
// If you need to test both see CreateFakeFileFromOSPipeExtended
func CreateFakeFileFromOSPipe(t *testing.T, errOnRead bool, errOnWrite bool) (r *FakeFile, w *FakeFile, err error) {
	osR, osW, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return &FakeFile{
			File:               osR,
			shouldErrorOnRead:  errOnRead,
			shouldErrorOnWrite: false,
			t:                  t,
		}, &FakeFile{
			File:               osW,
			shouldErrorOnRead:  false,
			shouldErrorOnWrite: errOnWrite,
			t:                  t,
		}, nil
}

// CreateFakeFileFromOSPipeExtended creates a new FakeFile instance from a call to os.Pipe()
// the failOnRead and failOnWrite parameters determine if the file should error on read or write
func CreateFakeFileFromOSPipeExtended(t *testing.T, rErrOnRead bool, rErrOnWrite bool, wErrOnRead bool, wErrOnWrite bool) (r *FakeFile, w *FakeFile, err error) {
	osR, osW, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return &FakeFile{
			File:               osR,
			shouldErrorOnRead:  rErrOnRead,
			shouldErrorOnWrite: rErrOnWrite,
			t:                  t,
		}, &FakeFile{
			File:               osW,
			shouldErrorOnRead:  wErrOnRead,
			shouldErrorOnWrite: wErrOnWrite,
			t:                  t,
		}, nil
}

// CreateFakeFileFromFileLike creates a new FakeFile instance from an existing FileLike (e.g. os.File)
func CreateFakeFileFromFileLike(t *testing.T, shouldErrorOnRead bool, shouldErrorOnWrite bool, actualFile commonInterfaces.FileLike) (*FakeFile, error) {
	if actualFile == nil {
		return nil, &FakeFileError{badInitialization: true}
	}
	return &FakeFile{
		File:               actualFile,
		shouldErrorOnRead:  shouldErrorOnRead,
		shouldErrorOnWrite: shouldErrorOnWrite,
		t:                  t,
	}, nil
}

// Write writes the given byte slice to the file
func (f *FakeFile) Chdir() error {
	if f.SetFail.Chdir || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Chdir()
}

// Chmod changes the mode of the file
func (f *FakeFile) Chmod(mode fs.FileMode) error {
	if f.SetFail.Chmod || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Chmod(mode)
}

// Chown changes the owner and group of the file
func (f *FakeFile) Chown(uid int, gid int) error {
	if f.SetFail.Chown || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Chown(uid, gid)
}

// Close closes the file
func (f *FakeFile) Close() error {
	if f.SetFail.Close || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Close()
}

// Fd returns the file descriptor
func (f *FakeFile) Fd() uintptr {
	if f.SetFail.Fd || f.t == nil {
		return 0
	}
	return f.File.Fd()
}

// Name returns the file name
func (f *FakeFile) Name() string {
	if f.SetFail.Name || f.t == nil {
		return ""
	}
	return f.File.Name()
}

// Read reads the given byte slice from the file
func (f *FakeFile) Read(b []byte) (n int, err error) {
	if f.shouldErrorOnRead || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Read(b)
}

// ReadAt reads the given byte slice from the file at the given offset
func (f *FakeFile) ReadAt(b []byte, off int64) (n int, err error) {
	if f.SetFail.ReadAt || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.ReadAt(b, off)
}

// ReadDir reads the directory
func (f *FakeFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.SetFail.ReadDir || f.t == nil {
		return nil, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.ReadDir(n)
}

// ReadFrom reads from the given reader
func (f *FakeFile) ReadFrom(r io.Reader) (n int64, err error) {
	if f.SetFail.ReadFrom || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.ReadFrom(r)
}

// Readdir reads the directory
func (f *FakeFile) Readdir(n int) ([]fs.FileInfo, error) {
	if f.SetFail.Readdir || f.t == nil {
		return nil, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Readdir(n)
}

// Readdirnames reads the directory names
func (f *FakeFile) Readdirnames(n int) (names []string, err error) {
	if f.SetFail.Readdirnames || f.t == nil {
		return nil, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Readdirnames(n)
}

// Seek seeks to the given offset
func (f *FakeFile) Seek(offset int64, whence int) (ret int64, err error) {
	if f.SetFail.Seek || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Seek(offset, whence)
}

// SetDeadline sets the deadline
func (f *FakeFile) SetDeadline(t time.Time) error {
	if f.SetFail.SetDeadline || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.SetDeadline(t)
}

// SetReadDeadline sets the read deadline
func (f *FakeFile) SetReadDeadline(t time.Time) error {
	if f.SetFail.SetReadDeadline || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (f *FakeFile) SetWriteDeadline(t time.Time) error {
	if f.SetFail.SetWriteDeadline || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.SetWriteDeadline(t)
}

// Stat returns the file info
func (f *FakeFile) Stat() (fs.FileInfo, error) {
	if f.SetFail.Stat || f.t == nil {
		return nil, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Stat()
}

// Sync syncs the file
func (f *FakeFile) Sync() error {
	if f.SetFail.Sync || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Sync()
}

// SyscallConn returns the raw connection
func (f *FakeFile) SyscallConn() (syscall.RawConn, error) {
	if f.SetFail.SyscallConn || f.t == nil {
		return nil, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.SyscallConn()
}

// Truncate truncates the file
func (f *FakeFile) Truncate(size int64) error {
	if f.SetFail.Truncate || f.t == nil {
		return &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Truncate(size)
}

// Write writes the given byte slice to the file
func (f *FakeFile) Write(b []byte) (n int, err error) {
	if f.shouldErrorOnWrite || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.Write(b)
}

// WriteAt writes the given byte slice to the file at the given offset
func (f *FakeFile) WriteAt(b []byte, off int64) (n int, err error) {
	if f.SetFail.WriteAt || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.WriteAt(b, off)
}

// WriteString writes the given string to the file
func (f *FakeFile) WriteString(s string) (n int, err error) {
	if f.SetFail.WriteString || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.WriteString(s)
}

// WriteTo writes the file to the given writer
func (f *FakeFile) WriteTo(w io.Writer) (n int64, err error) {
	if f.SetFail.WriteTo || f.t == nil {
		return 0, &FakeFileError{badInitialization: f.t == nil}
	}
	return f.File.WriteTo(w)
}

// FakeFileError is an error that is returned when the FakeFile is intentionally errored
type FakeFileError struct {
	badInitialization bool
}

// Error returns the error message string
func (f *FakeFileError) Error() string {
	if f.badInitialization {
		return "FakeFile not properly initialized"
	}
	return "FakeFile intentionally errored"
}
