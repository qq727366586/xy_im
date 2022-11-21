package comet

import (
	"context"
	"github.com/zhenjl/cityhash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"time"
	"xy_im/api/logic"
	"xy_im/internal/comet/conf"
)

const (
	minServerHeartbeat = time.Minute * 10
	maxServerHeartbeat = time.Minute * 30

	// grpc
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
	grpcBackoffMaxDelay       = time.Second * 3
)

type Server struct {
	c           *conf.Config
	serverID    string
	round       *Round
	buckets     []*Bucket
	bucketIndex uint32
	rpcClient   logic.LogicClient
}

func newLogicClient(c *conf.RPCClient) logic.LogicClient {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()
	conn, err := grpc.DialContext(ctx, c.Bind,
		[]grpc.DialOption{
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),
		}...)
	if err != nil {
		panic(err)
	}
	return logic.NewLogicClient(conn)

}

func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:         c,
		round:     NewRound(c),
		rpcClient: newLogicClient(c.RPCClient),
	}
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIndex = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	// todo 房间在线人数
	s.serverID = c.Env.Host
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
