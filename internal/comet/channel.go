package comet

import (
	"sync"
	"xy_im/api/protocol"
	"xy_im/internal/comet/errors"
	"xy_im/pkg/bufio"
)

type Channel struct {
	mutex    sync.RWMutex
	Mid      int64
	Key      string
	IP       string
	Room     *Room
	CliProto Ring

	Reader bufio.Reader
	Writer bufio.Writer

	Prev *Channel
	Next *Channel

	watchOps map[int32]struct{}
	signal   chan *protocol.Proto
}

func NewChannel(cli, srv int) *Channel {
	c := new(Channel)
	c.CliProto.Init(cli)
	c.signal = make(chan *protocol.Proto, srv)
	c.watchOps = make(map[int32]struct{})
	return c
}

func (c *Channel) Push(p *protocol.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
		err = errors.ErrSignalFullMsgDropped
	}
	return
}

func (c *Channel) Watch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	c.mutex.Unlock()
}

func (c *Channel) Ready() *protocol.Proto {
	return <-c.signal
}

func (c *Channel) Signal() {
	c.signal <- protocol.ProtoReady
}

func (c *Channel) Close() {
	c.signal <- protocol.ProtoFinish
}
