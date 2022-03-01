package runtime

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

type socket string

func (s socket) path() string {
	return strings.TrimPrefix(string(s), "unix://")
}

func SocketPath(addr string) string {
	return socket(addr).path()
}

const socketRoot = "/run/zcm"

func SocketAddr(ctx context.Context, socketPath, id string) (string, error) {
	d := sha256.Sum256([]byte(filepath.Join(socketPath, id)))
	return fmt.Sprintf("unix://%s/%x", filepath.Join(socketRoot, "s"), d), nil
}

func NewSocket(addr string) (*net.UnixListener, error) {
	var (
		sock = socket(addr)
		path = sock.path()
	)
	if err := os.MkdirAll(filepath.Dir(path), 0600); err != nil {
		return nil, errors.Wrapf(err, "%s", path)
	}
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0600); err != nil {
		os.Remove(sock.path())
		l.Close()
		return nil, err
	}
	return l.(*net.UnixListener), nil
}

func RemoveSocket(addr string) error {
	sock := socket(addr)
	return os.Remove(sock.path())
}

func WriteAddr(path, addr string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	tmpPath := filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s", filepath.Base(path)))
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	_, err = f.WriteString(addr)
	f.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func SocketEaddrinuse(err error) bool {
	netErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	if netErr.Op != "listen" {
		return false
	}
	syscallErr, ok := netErr.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	errno, ok := syscallErr.Err.(syscall.Errno)
	if !ok {
		return false
	}
	return errno == syscall.EADDRINUSE
}
