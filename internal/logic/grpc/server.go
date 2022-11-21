package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"time"
	pb "xy_im/api/logic"
	"xy_im/internal/logic"
	"xy_im/internal/logic/conf"
)

func New(c *conf.RPCServer, l *logic.Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	pb.RegisterLogicServer(srv, &server{l})
}

type server struct {
}

func (s *server) Connect(c context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	return nil, nil

}
func (s *server) Disconnect(c context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	return nil, nil
}

func (s *server) Heartbeat(c context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	return nil, nil
}

func (s *server) RenewOnline(c context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	return nil, nil
}

func (s *server) Receive(c context.Context, req *pb.ReceiveReq) (*pb.ReceiveReply, error) {
	return nil, nil
}

func (s *server) Nodes(c context.Context, req *pb.NodesReq) (*pb.NodesReply, error) {
	return nil, nil
}

func (s *server) mustEmbedUnimplementedLogicServer() {

}
