package main

import (
	"context"
	"fmt"
	"log"
	"net/rpc"
	"simpleconman/playground/rpc/common"
	"simpleconman/runtime"
)

func main() {
	ctx := context.Background()

	addr, err := runtime.SocketAddr(ctx, "./sock", "rpc_test")
	if err != nil {
		log.Fatal(err)
	}
	cli, err := rpc.Dial("unix", runtime.SocketPath(addr))
	if err != nil {
		log.Fatal(err)
	}
	resp := &common.Response{}
	if err := cli.Call("Handler.Execute", common.Request{Id: "user"}, resp); err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}
