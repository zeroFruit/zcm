package runtime

import (
	"context"
	"net"
	"net/rpc"
	"os"

	"github.com/sirupsen/logrus"
)

type server struct {
	sigChan chan os.Signal
}

func (s *server) serve(ctx context.Context) error {
	// fd 3 is set by shim start bin
	l, err := net.FileListener(os.NewFile(3, "socket"))
	path := "[inherited from parent]"
	if err != nil {
		return err
	}
	logrus.WithField("socket", path).Info("serving api on socket")

	// for now, it is empty service rpc server
	go func() {
		defer l.Close()
		rpc.Accept(l)
	}()

	return handleSignals(ctx, s.sigChan)
}
