package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"xy_im/internal/logic"
	"xy_im/internal/logic/conf"
	"xy_im/internal/logic/grpc"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	srv := logic.New(conf.Conf)
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	// 关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			rpcSrv.GracefulStop()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
