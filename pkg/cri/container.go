package cri

import (
	"fmt"
	"simpleconman/pkg/container"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func Status(s container.Status) runtimeapi.ContainerState {
	switch s {
	case container.Initial, container.Created:
		return runtimeapi.ContainerState_CONTAINER_CREATED
	case container.Running:
		return runtimeapi.ContainerState_CONTAINER_RUNNING
	case container.Stopped:
		return runtimeapi.ContainerState_CONTAINER_EXITED
	case container.Unknown:
		return runtimeapi.ContainerState_CONTAINER_UNKNOWN
	}
	panic(fmt.Sprintf("unknown state: %s", s.String()))
}
