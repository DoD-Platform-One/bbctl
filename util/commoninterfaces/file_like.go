package commoninterfaces

import (
	"io"
	"io/fs"
	"syscall"
	"time"
)

type FileLike interface {
	Chdir() error
	Chmod(mode fs.FileMode) error
	Chown(uid int, gid int) error
	Close() error
	Fd() uintptr
	Name() string
	Read(b []byte) (int, error)
	ReadAt(b []byte, off int64) (int, error)
	ReadDir(n int) ([]fs.DirEntry, error)
	ReadFrom(r io.Reader) (int64, error)
	Readdir(n int) ([]fs.FileInfo, error)
	Readdirnames(n int) ([]string, error)
	Seek(offset int64, whence int) (int64, error)
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	Stat() (fs.FileInfo, error)
	Sync() error
	SyscallConn() (syscall.RawConn, error)
	Truncate(size int64) error
	Write(b []byte) (int, error)
	WriteAt(b []byte, off int64) (int, error)
	WriteString(s string) (int, error)
	WriteTo(w io.Writer) (int64, error)
}
