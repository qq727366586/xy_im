package logic

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"time"
	"xy_im/internal/logic/model"
)

type params struct {
	Mid      int64   `json:"mid"`
	Key      string  `json:"key"`
	RoomId   string  `json:"room_id"`
	Platform string  `json:"platform"`
	Accepts  []int32 `json:"accepts"`
}

func (l *Logic) Connect(c context.Context, server, cookie string, token []byte) (mid int64, key, roomId string, accepts []int32, hb int64, err error) {
	p := new(params)
	if err = json.Unmarshal(token, p); err != nil {
		return
	}
	mid = p.Mid
	roomId = p.RoomId
	accepts = p.Accepts
	hb = int64(l.c.Node.Heartbeat * l.c.Node.HeartbeatMax * 60)
	if key = p.Key; key == "" {
		key = uuid.New().String()
	}
	if err = l.dao.HSet(c, mid, key, server); err != nil {
		return
	}
	return
}

func (l *Logic) Disconnect(c context.Context, mid int64, key, server string) (has bool, err error) {
	return l.dao.HDel(c, mid, key, server)
}

func (l *Logic) Heartbeat(c context.Context, mid int64, key, server string) (err error) {
	has, err := l.dao.Expire(c, mid, key)
	if err != nil {
		return
	}
	if !has {
		if err = l.dao.HSet(c, mid, key, server); err != nil {
			return
		}
	}
	return
}

func (l *Logic) RenewOnline(c context.Context, server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &model.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.dao.AddServerOnline(c, server, online); err != nil {
		return nil, err
	}
	return l.roomCount, nil
}
