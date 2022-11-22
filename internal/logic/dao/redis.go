package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/zhenjl/cityhash"
	"strconv"
	"time"
	"xy_im/internal/logic/conf"
	"xy_im/internal/logic/model"
)

const (
	_prefixMidServer    = "mid_%d"
	_prefixKeyServer    = "key_%s"
	_prefixServerOnline = "online_%s"
)

func keyMidServer(mid int64) string {
	return fmt.Sprintf(_prefixMidServer, mid)
}

func keyKeyServer(key string) string {
	return fmt.Sprintf(_prefixKeyServer, key)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(_prefixServerOnline, key)
}

func newRedis(c *conf.Redis) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Password:     c.Auth,
		DialTimeout:  time.Duration(c.DialTimeout) * time.Second,
		IdleTimeout:  time.Duration(c.IdleTimeout) * time.Second,
		ReadTimeout:  time.Duration(c.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.WriteTimeout) * time.Second,
		PoolSize:     c.PoolSize,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(err)
	}
	return client
}

func (b *Base) HSet(c context.Context, mid int64, key, server string) (err error) {
	expire := time.Duration(b.c.Redis.Expire) * time.Second
	if mid > 0 {
		if _, err = b.redis.HSet(c, keyMidServer(mid), key, server).Result(); err != nil {
			return
		}
		if _, err = b.redis.Expire(c, keyMidServer(mid), expire).Result(); err != nil {
			return
		}
	}
	if _, err = b.redis.Set(c, keyKeyServer(key), server, expire).Result(); err != nil {
		return
	}
	return
}

func (b *Base) HDel(c context.Context, mid int64, key, server string) (has bool, err error) {
	if mid > 0 {
		if _, err = b.redis.HDel(c, keyMidServer(mid), key).Result(); err != nil {
			return
		}
	}
	if _, err = b.redis.Del(c, keyKeyServer(key)).Result(); err != nil {
		return
	}
	return true, nil
}

func (b *Base) Expire(c context.Context, mid int64, key string) (has bool, err error) {
	expire := time.Duration(b.c.Redis.Expire) * time.Second
	if mid > 0 {
		if _, err = b.redis.Expire(c, keyMidServer(mid), expire).Result(); err != nil {
			return
		}
	}
	if _, err = b.redis.Expire(c, keyKeyServer(key), expire).Result(); err != nil {
		return
	}
	return true, nil
}

func (b *Base) AddServerOnline(c context.Context, server string, online *model.Online) (err error) {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64] = rMap
		}
		rMap[room] = count
	}
	key := keyServerOnline(server)
	for hashKey, value := range roomsMap {
		err = b.addServerOnline(c, key, strconv.FormatInt(int64(hashKey), 10), &model.Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

func (b *Base) addServerOnline(c context.Context, key string, hashKey string, online *model.Online) (err error) {
	bt, _ := json.Marshal(online)
	expire := time.Duration(b.c.Redis.Expire) * time.Second
	if _, err = b.redis.HSet(c, key, hashKey, bt).Result(); err != nil {
		return
	}
	if _, err = b.redis.Expire(c, key, expire).Result(); err != nil {
		return
	}
	return
}
