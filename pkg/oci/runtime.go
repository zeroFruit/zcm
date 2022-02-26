package oci

import (
	"simpleconman/pkg/container"
	"time"
)

type Runtime interface {
	CreateContainer(handle *container.Handle, stdin bool,
		stdinOnce bool, timeout time.Duration) (*container.Instance, error)
	StartContainer(*container.Handle) error
	Container(handle *container.Handle) (*container.Instance, error)
}
