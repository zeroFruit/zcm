package cri

import (
	"context"
	"fmt"
	"path"
	"simpleconman/pkg/container"
	"simpleconman/pkg/oci"
	"time"

	"github.com/pkg/errors"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type runtimeService struct {
	runtime         oci.Runtime
	containerGetter container.Getter

	rootDir   string
	logDir    string
	exitDir   string
	attachDir string

	timeout time.Duration

	store container.Store
}

// CreateContainer creates a new container in specified PodSandbox.
// FIXME: currently this method is not atomic
func (s *runtimeService) CreateContainer(ctx context.Context,
	req *runtimeapi.CreateContainerRequest) (*runtimeapi.CreateContainerResponse, error) {
	handle, err := container.NewHandle(
		s.containerGetter,
		s.containerDir,
		s.logFile,
		s.attachFile,
		s.exitFile,
	)
	if err != nil {
		return nil, err
	}
	spec, err := oci.NewSpec(oci.SpecOptions{
		Command:      req.GetConfig().GetCommand(),
		Args:         req.GetConfig().GetArgs(),
		RootPath:     handle.RootfsDir(),
		RootReadonly: req.GetConfig().GetLinux().GetSecurityContext().GetReadonlyRootfs(),
	})
	if err != nil {
		return nil, err
	}
	if err := handle.Bundle(spec, s.rootDir); err != nil {
		return nil, err
	}

	_, err = s.runtime.CreateContainer(handle, req.GetConfig().GetStdin(),
		req.GetConfig().GetStdinOnce(), s.timeout)
	if err != nil {
		return nil, err
	}
	if err := handle.Created(); err != nil {
		return nil, err
	}
	if err := s.store.Put(handle); err != nil {
		return nil, err
	}
	return &runtimeapi.CreateContainerResponse{
		ContainerId: handle.Id().String(),
	}, nil
}

func (s *runtimeService) containerDir(id container.Id) string {
	return path.Join(s.containersDir(), id.String())
}

func (s *runtimeService) containersDir() string {
	return path.Join(s.rootDir, "containers")
}

func (s *runtimeService) logFile(id container.Id) string {
	return path.Join(s.logDir, id.String()+".log")
}

func (s *runtimeService) attachFile(id container.Id) string {
	return path.Join(s.attachDir, id.String())
}

func (s *runtimeService) exitFile(id container.Id) string {
	return path.Join(s.exitDir, id.String())
}

// StartContainer starts the container.
func (s *runtimeService) StartContainer(ctx context.Context,
	req *runtimeapi.StartContainerRequest) (*runtimeapi.StartContainerResponse, error) {
	id := container.Id(req.ContainerId)
	cont, handle, err := s.containerGetter.Get(id)
	if err != nil {
		return nil, err
	}
	if !cont.CanStart() {
		return nil, fmt.Errorf("cannot start container. container status [%s]",
			cont.Status.String())
	}
	if err := s.runtime.StartContainer(handle); err != nil {
		return nil, errors.Wrap(err, "cannot start container")
	}
	if err := handle.Started(); err != nil {
		return nil, errors.Wrap(err, "cannot update status to started")
	}

	delays := 200 * time.Millisecond
	for {
		// TODO: add timeout
		time.Sleep(delays)
		cont, _, err := s.containerGetter.Get(id)
		if err != nil {
			return nil, err
		}
		if cont.Status == container.Running {
			break
		}
		if cont.Status != container.Created {
			return nil, fmt.Errorf("failed to start container. status=%v", cont.Status)
		}
		// exponential back-off delays
		delays *= 2
	}

	return &runtimeapi.StartContainerResponse{}, nil
}

// StopContainer stops a running container with a grace period (i.e., timeout).
func (s *runtimeService) StopContainer(ctx context.Context, r *runtimeapi.StopContainerRequest) (*runtimeapi.StopContainerResponse, error) {
	return nil, nil
}

// RemoveContainer removes the container.
func (s *runtimeService) RemoveContainer(ctx context.Context, r *runtimeapi.RemoveContainerRequest) (*runtimeapi.RemoveContainerResponse, error) {
	return nil, nil
}

// ListContainers lists all containers by filters.
func (s *runtimeService) ListContainers(ctx context.Context, r *runtimeapi.ListContainersRequest) (*runtimeapi.ListContainersResponse, error) {
	iter := s.store.Iter()
	result := []*runtimeapi.Container{}
	for iter.HasNext() {
		handle := iter.Next()
		cont, err := s.runtime.Container(handle)
		if err != nil {
			return nil, err
		}
		result = append(result, &runtimeapi.Container{
			Id:           handle.Id().String(),
			PodSandboxId: "",
			Image: &runtimeapi.ImageSpec{
				Image:       "",
				Annotations: nil,
			},
			ImageRef:    "",
			CreatedAt:   cont.CreatedAt.UTC().Unix(),
			Labels:      nil,
			Annotations: nil,
		})
	}
	return &runtimeapi.ListContainersResponse{
		Containers: result,
	}, nil
}

// ContainerStatus returns the status of the container.
func (s *runtimeService) ContainerStats(ctx context.Context, r *runtimeapi.ContainerStatsRequest) (*runtimeapi.ContainerStatsResponse, error) {
	handle, err := s.store.Get(container.Id(r.ContainerId))
	if err != nil {
		return nil, err
	}
	_, err = s.runtime.Container(handle)
	if err != nil {
		return nil, err
	}
	return &runtimeapi.ContainerStatsResponse{
		Stats: &runtimeapi.ContainerStats{
			Attributes: &runtimeapi.ContainerAttributes{
				Id: handle.Id().String(),
			},
			Cpu: &runtimeapi.CpuUsage{
				Timestamp: time.Now().UTC().Unix(),
			},
			Memory: &runtimeapi.MemoryUsage{
				Timestamp: time.Now().UTC().Unix(),
			},
		},
	}, nil
}
