package logic

import (
	"xy_im/internal/logic/conf"
	"xy_im/internal/logic/dao"
)

type Logic struct {
	c         *conf.Config
	dao       *dao.Base
	roomCount map[string]int32
}

func New(c *conf.Config) (logic *Logic) {
	logic = &Logic{
		c:   c,
		dao: dao.New(c),
	}
	return
}
