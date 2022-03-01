package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
}

func main() {
	logrus.Info("hello. this is from child")
	//if err := ioutil.WriteFile("./a.txt", []byte("hello from client"), 0755); err != nil {
	//	logrus.Fatal("failed to write file from client")
	//}
}
