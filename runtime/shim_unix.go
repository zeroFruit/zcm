// +build !windows

package runtime

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// setupSignals creates a new signal handler for all signals and sets the shim as a
// sub-reaper so that the container processes are reparented
func setupSignals() (chan os.Signal, error) {
	sigChan := make(chan os.Signal, 32)
	signals := []os.Signal{unix.SIGTERM, unix.SIGINT, unix.SIGPIPE, unix.SIGCHLD}
	signal.Notify(sigChan, signals...)
	return sigChan, nil
}

func handleSignals(ctx context.Context, sigChan chan os.Signal) error {
	logrus.Info("starting signal loop")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-sigChan:
			switch s {
			case unix.SIGCHLD:
				// TODO: reap child
			case unix.SIGPIPE:
			}
		}
	}
}
