package dao

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"xy_im/internal/logic/conf"
)

type Base struct {
	c     *conf.Config
	redis *redis.Client
}

func New(c *conf.Config) *Base {
	fmt.Println(c.Redis)
	return &Base{
		c:     c,
		redis: newRedis(c.Redis),
	}
}
