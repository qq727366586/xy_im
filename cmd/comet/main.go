package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"xy_im/internal/comet"
	"xy_im/internal/comet/conf"
)

func main() {
	// 初始化参数
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	// 随机数种子
	rand.Seed(time.Now().UTC().UnixNano())
	// cpu
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 初始化server
	server := comet.NewServer(conf.Conf)
	// 初始化ws
	if err := comet.InitWebsocket(server, conf.Conf.Websocket.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			// todo close

			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
