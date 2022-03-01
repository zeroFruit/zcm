package main

import (
	"context"
	"log"
	"net/rpc"
	"simpleconman/playground/rpc/common"
	"simpleconman/runtime"
)

func main() {
	ctx := context.Background()
	if err := rpc.Register(&common.Handler{}); err != nil {
		log.Fatal(err)
	}
	addr, err := runtime.SocketAddr(ctx, "./sock", "rpc_test")
	if err != nil {
		log.Fatal(err)
	}
	if err := runtime.RemoveSocket(addr); err != nil {
		log.Fatal(err)
	}
	log.Println("addr", addr)
	sock, err := runtime.NewSocket(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer sock.Close()

	rpc.Accept(sock)
}
