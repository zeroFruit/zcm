package runc

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"simpleconman/runtime"
	"syscall"

	"github.com/pkg/errors"
)

type service struct{}

func (s *service) Start(ctx context.Context, id string) (_ string, retErr error) {
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// TODO: fill the args to run the shim daemon
	args := []string{}
	cmd := exec.Command(self, args...)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	addr, err := runtime.SocketAddr(ctx, "zcm", id)
	if err != nil {
		return "", err
	}
	sock, err := runtime.NewSocket(addr)
	if err != nil {
		if !runtime.SocketEaddrinuse(err) {
			return "", errors.Wrap(err, "create new shim socket")
		}
		if err := runtime.RemoveSocket(addr); err != nil {
			return "", errors.Wrap(err, "remove pre-existing socket")
		}
		if sock, err = runtime.NewSocket(addr); err != nil {
			return "", errors.Wrap(err, "try create new shim socket 2x")
		}
	}
	defer func() {
		if retErr != nil {
			sock.Close()
			_ = runtime.RemoveSocket(addr)
		}
	}()
	if err := runtime.WriteAddr("address", addr); err != nil {
		return "", err
	}

	f, err := sock.File()
	if err != nil {
		return "", err
	}

	cmd.ExtraFiles = append(cmd.ExtraFiles, f)

	if err := cmd.Start(); err != nil {
		f.Close()
		return "", err
	}
	defer func() {
		if retErr != nil {
			cmd.Process.Kill()
		}
	}()

	go cmd.Wait()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return addr, nil
	}

	// TODO: parse data with runc options.Options then do something
	fmt.Println("data", data)
	return addr, nil
}
