package oci

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"simpleconman/pkg/container"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type runcRuntime struct {
	// shimmyPath is path to shimy executable
	shimmyPath string

	// runtimePath is path to runc executable
	runtimePath string

	// rootPath is directory path to store container state
	rootPath string
}

func NewRuncRuntime(shimmyPath string, runtimePath string, rootPath string) *runcRuntime {
	return &runcRuntime{
		shimmyPath:  shimmyPath,
		runtimePath: runtimePath,
		rootPath:    rootPath,
	}
}

func (r *runcRuntime) CreateContainer(handle *container.Handle,
	stdin bool, stdinOnce bool, timeout time.Duration) (*container.Instance, error) {
	cmd := exec.Command(
		r.shimmyPath,
		"--shimmy-pidfile", path.Join(handle.BundleDir(), "shimmy.pid"),
		"--shimmy-log-level", strings.ToUpper(logrus.GetLevel().String()),
		"--runtime", r.runtimePath,
		fmt.Sprintf("--runtime-arg='--root=%s'", r.rootPath),
		"--bundle", handle.BundleDir(),
		"--container-id", handle.Id().String(),
		"--container-pidfile", path.Join(handle.BundleDir(), "container.pid"),
		"--container-logfile", handle.LogFile(),
		"--container-exitfile", handle.ExitFile(),
		"--container-attachfile", handle.AttachFile(),
	)
	if stdin {
		cmd.Args = append(cmd.Args, "--stdin")
	}
	if stdinOnce {
		cmd.Args = append(cmd.Args, "--stdin-once")
	}

	syncPipeRead, syncPipeWrite, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer syncPipeRead.Close()
	defer syncPipeWrite.Close()

	cmd.ExtraFiles = append(cmd.ExtraFiles, syncPipeWrite)
	cmd.Args = append(
		cmd.Args,
		"--syncpipe-fd", strconv.Itoa(2+len(cmd.ExtraFiles)),
	)

	if _, err := runCommand(cmd); err != nil {
		return nil, err
	}

	type Report struct {
		Kind   string `json:"kind"`
		Status string `json:"status"`
		Stderr string `json:"stderr"`
		Pid    int    `json:"pid"`
	}

	type Result struct {
		Err    error
		Report Report
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ch := make(chan Result, 1)

	go func() {
		defer cancel()

		b, err := ioutil.ReadAll(syncPipeRead)
		if err != nil {
			ch <- Result{Err: err}
		}
		syncPipeRead.Close()

		report := &Report{}
		if err := json.Unmarshal(b, report); err != nil {
			ch <- Result{
				Err: errors.Wrap(err,
					fmt.Sprintf("failed to decode report [%v]. raw [%v]", string(b), b)),
			}
		}

		if report.Kind != "container_pid" && report.Pid <= 0 {
			ch <- Result{Err: errors.Errorf("%+v", report)}
		}

		ch <- Result{Report: *report}
	}()

	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "timeout")
	case result := <-ch:
		if result.Err != nil {
			return nil, err
		}
		// FIXME: maybe we don't need to return as *Instance type
		return &container.Instance{Pid: uint32(result.Report.Pid)}, nil
	}
}

func (r *runcRuntime) StartContainer(handle *container.Handle) error {
	cmd := exec.Command(
		r.runtimePath,
		"--root", r.rootPath,
		"start", handle.Id().String(),
	)
	_, err := runCommand(cmd)
	return err
}

func (r *runcRuntime) Container(handle *container.Handle) (*container.Instance, error) {
	cmd := exec.Command(
		r.runtimePath,
		"--root", r.rootPath,
		"state",
		handle.Id().String(),
	)
	b, err := runCommand(cmd)
	if err != nil {
		return nil, err
	}
	result := &container.Instance{}
	return result, json.Unmarshal(b, result)
}

func runCommand(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.Output()
	debugLog(cmd, output, err)
	return output, wrappedError(err)
}

func debugLog(cmd *exec.Cmd, stdout []byte, err error) {
	stderr := []byte{}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = ee.Stderr
		}
	}

	logrus.WithFields(logrus.Fields{
		"stdout": string(stdout),
		"stderr": string(stderr),
		"error":  err,
	}).Debugf("exec %s", strings.Join(cmd.Args, " "))
}

func wrappedError(err error) error {
	if err == nil {
		return nil
	}

	msg := "OCI runtime (runc) execution failed"
	if ee, ok := err.(*exec.ExitError); ok {
		msg = fmt.Sprintf("%v, stderr=[%v]", msg, string(ee.Stderr))
	}

	return errors.Wrap(err, msg)
}

type runcContainerGetter struct {
	readOnlyStore container.ReadOnlyStore
	runcRuntime   *runcRuntime
}

func (r *runcContainerGetter) Get(id container.Id) (*container.Instance, *container.Handle, error) {
	handle, err := r.readOnlyStore.Get(id)
	if err != nil {
		return nil, nil, err
	}
	cont, err := r.runcRuntime.Container(handle)
	if err != nil {
		return nil, nil, err
	}
	return cont, handle, nil
}

func (r *runcContainerGetter) List() ([]*container.Instance, error) {
	panic("not implemented") // TODO: Implement
}
