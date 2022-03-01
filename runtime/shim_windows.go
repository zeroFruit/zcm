// +build windows

package runtime

import (
	"context"
	"os"
)

func setupSignals() (chan os.Signal, error) {
	return nil, nil
}

func handleSignals(ctx context.Context, sigChan chan os.Signal) error {
	return nil
}
