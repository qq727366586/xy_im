package comet

import (
	"sync"
	"xy_im/api/comet"
	"xy_im/internal/comet/conf"
)

type Bucket struct {
	c           *conf.Bucket
	cLock       sync.RWMutex
	rooms       map[string]*Room
	chs         map[string]*Channel
	ipCnts      map[string]int32
	routines    []chan *comet.BroadcastRoomReq
	routinesNum uint64
}

func NewBucket(c *conf.Bucket) (b *Bucket) {
	b = new(Bucket)
	b.chs = make(map[string]*Channel, c.Channel)
	b.ipCnts = make(map[string]int32)
	b.c = c
	b.rooms = make(map[string]*Room, c.Room)
	b.routines = make([]chan *comet.BroadcastRoomReq, c.RoutineAmount)
	for i := uint64(0); i < c.RoutineAmount; i++ {
		chs := make(chan *comet.BroadcastRoomReq, c.RoutineSize)
		b.routines[i] = chs
		go b.roomProc(chs)
	}
	return
}

func (b *Bucket) roomProc(c chan *comet.BroadcastRoomReq) {
	for {
		arg := <-c
		if room := b.Room(arg.RoomID); room != nil {
			room.Push(arg.Proto)
		}
	}
}

func (b *Bucket) Room(rid string) (room *Room) {
	b.cLock.RLock()
	room = b.rooms[rid]
	b.cLock.RUnlock()
	return
}

func (b *Bucket) RoomsCount() (res map[string]int32) {
	b.cLock.RLock()
	res = make(map[string]int32)
	for roomId, room := range b.rooms {
		if room.Online > 0 {
			res[roomId] = room.Online
		}
	}
	b.cLock.RUnlock()
	return
}

func (b *Bucket) Put(rid string, ch *Channel) (err error) {
	var (
		room *Room
		ok   bool
	)
	b.cLock.Lock()
	// 关闭旧的channel
	if dch := b.chs[ch.Key]; dch != nil {
		dch.Close()
	}
	// 重新赋值
	b.chs[ch.Key] = ch
	if rid != "" {
		if room, ok = b.rooms[rid]; !ok {
			room = NewRoom(rid)
			b.rooms[rid] = room
		}
		ch.Room = room
	}
	b.ipCnts[ch.IP]++
	b.cLock.Unlock()
	if room != nil {
		err = room.Put(ch)
	}
	return
}

func (b *Bucket) Del(dch *Channel) {
	room := dch.Room
	b.cLock.Lock()
	if ch, ok := b.chs[dch.Key]; ok {
		if ch == dch {
			delete(b.chs, ch.Key)
		}
		if b.ipCnts[ch.IP] > 1 {
			b.ipCnts[ch.IP]--
		} else {
			delete(b.ipCnts, ch.IP)
		}
	}
	b.cLock.Unlock()
	if room != nil && room.Del(dch) {
		// 如果是个空房间, 必须移除
		b.DelRoom(room)
	}
}

func (b *Bucket) DelRoom(room *Room) {
	b.cLock.Lock()
	delete(b.rooms, room.ID)
	b.cLock.Unlock()
	room.Close()
}
