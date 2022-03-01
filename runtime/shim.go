package runtime

import (
	"context"
	"flag"
	"os"
)

type Shim interface {
	Start(ctx context.Context, id string) (string, error)
}

type Bootstrapper interface {
	Bootstrap() Shim
}

var (
	shimPidFile string
	runtime     string
	bundle      string
	action      string

	containerId       string
	containerPidFile  string
	containerLogFile  string
	containerExitFile string
)

func parseFlags() {
	flag.StringVar(&shimPidFile, "shim-pid", "", "path to file which contains shim pid")
	flag.StringVar(&runtime, "runtime", "", "path to runtime binary")
	flag.StringVar(&bundle, "bundle", "", "path to bundle")
	flag.StringVar(&action, "action", "", "action for shim")
	flag.StringVar(&containerId, "id", "", "container id")
	flag.StringVar(&containerPidFile, "pid-file", "", "path to container pid file")
	flag.StringVar(&containerLogFile, "log-file", "", "path to container log file")
	flag.StringVar(&containerExitFile, "exit-file", "", "path to container exit file")
}

func Run(bootstrapper Bootstrapper) {
	if err := run(bootstrapper); err != nil {
		// TODO: write error to stderr
	}
}

func run(bootstrapper Bootstrapper) error {
	parseFlags()

	shim := bootstrapper.Bootstrap()
	ctx := context.Background()

	sigChan, err := setupSignals()
	if err != nil {
		return err
	}

	switch action {
	// "start" runs runtime to create container and create unix socket. That address is
	// returned from shim.Start(). Then shim starts itself again with no action args.
	case "start":
		addr, err := shim.Start(ctx, containerId)
		if err != nil {
			return err
		}
		if _, err := os.Stdout.WriteString(addr); err != nil {
			return err
		}
		return nil
	}

	// set self as subreaper

	// serve rpc server and waits for container with pid killed by SIGCHLD sig
	server := &server{
		sigChan: sigChan,
	}
	if err := server.serve(ctx); err != nil {
		return err
	}

	// TODO: clean up
	return nil
}
