package comet

import (
	"sync"
	"xy_im/internal/comet/conf"
)

type Bucket struct {
	c           *conf.Config
	cLock       sync.RWMutex
	rooms       map[string]*Room
	ipCnts      map[string]int32
	routinesNum uint64
}
