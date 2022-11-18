package comet

import "sync"

type Room struct {
	ID        string
	rLock     sync.RWMutex
	next      *Channel
	drop      bool
	Online    int32
	AllOnline int32
}
