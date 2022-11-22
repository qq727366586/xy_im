package logic

import (
	"xy_im/internal/logic/conf"
)

type Logic struct {
	c *conf.Config
}

func New(c *conf.Config) (logic *Logic) {
	logic = &Logic{
		c: c,
	}
	return
}
