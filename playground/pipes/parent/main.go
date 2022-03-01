package main

import (
	"bufio"
	"flag"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

var (
	childBin string
)

var logger = logrus.WithFields(logrus.Fields{"loc": "parent"})

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
}

func main() {
	flag.StringVar(&childBin, "child", "", "child binary path")
	flag.Parse()

	logger.Info("childBin: " + childBin)

	cmd := exec.Command(childBin)

	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.Fatal(err)
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		logrus.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		logger.Fatal(err)
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Fatal(err)
		}
	}()

	linesCh := make(chan string)
	go func() {
		sc := bufio.NewScanner(cmdStdout)
		for sc.Scan() {
			linesCh <- sc.Text()
		}
	}()

	//timeout := time.After(5 * time.Second)
	select {
	case line := <-linesCh:
		logger.Info("message from buf: " + line)
	}
}
