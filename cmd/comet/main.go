package main

import (
	"flag"
	"math/rand"
	"runtime"
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
	comet.NewServer(conf.Conf)

	// 初始化ws

}
