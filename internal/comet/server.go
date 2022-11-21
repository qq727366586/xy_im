package comet

import (
	"github.com/zhenjl/cityhash"
	"xy_im/internal/comet/conf"
)

type Server struct {
	c           *conf.Config
	round       *Round
	buckets     []*Bucket
	bucketIndex uint32
}

func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:     c,
		round: NewRound(c),
	}
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIndex = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	// todo 房间在线人数
	go s.onlineProc()
	return s
}

func (s *Server) onlineProc() {
	for {
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomId, count := range bucket.RoomsCount() {
				roomCount[roomId] += count
			}
		}

	}
}

func (s *Server) Bucket(subKey string) *Bucket {
	idx := cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIndex
	return s.buckets[idx]
}
