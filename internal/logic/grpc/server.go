package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"net"
	"time"
	pb "xy_im/api/logic"
	"xy_im/internal/logic"
	"xy_im/internal/logic/conf"
)

func New(c *conf.RPCServer, logic *logic.Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	pb.RegisterLogicServer(srv, &server{srv: logic})
	listen, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err = srv.Serve(listen); err != nil {
			panic(err)
		}
	}()
	return srv
}

type server struct {
	srv *logic.Logic
	pb.UnimplementedLogicServer
}

func (s *server) Connect(c context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	mid, key, room, accepts, hb, err := s.srv.Connect(c, req.Server, req.Cookie, req.Token)
	if err != nil {
		return &pb.ConnectReply{}, err
	}
	return &pb.ConnectReply{
		Mid:       mid,
		Key:       key,
		RoomID:    room,
		Accepts:   accepts,
		Heartbeat: hb,
	}, nil
}

func (s *server) Disconnect(c context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.srv.Disconnect(c, req.Mid, req.Key, req.Server)
	if err != nil {
		return &pb.DisconnectReply{}, err
	}
	return &pb.DisconnectReply{Has: has}, nil
}

func (s *server) Heartbeat(c context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.srv.Heartbeat(c, req.Mid, req.Key, req.Server); err != nil {
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

func (s *server) RenewOnline(c context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.srv.RenewOnline(c, req.Server, req.RoomCount)
	if err != nil {
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}

func (s *server) Receive(c context.Context, req *pb.ReceiveReq) (*pb.ReceiveReply, error) {
	return &pb.ReceiveReply{}, nil
}

func (s *server) Nodes(c context.Context, req *pb.NodesReq) (*pb.NodesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Nodes not implemented")
}
