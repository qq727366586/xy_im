package comet

import "xy_im/internal/comet/conf"

type Server struct {
	c     *conf.Config
	round *Round
}

func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:     c,
		round: NewRound(c),
	}
	return s
}
