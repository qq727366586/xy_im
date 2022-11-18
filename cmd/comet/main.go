package main

import (
	"flag"
	"math/rand"
	"runtime"
	"time"
	"xy_im/internal/comet/conf"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

}
