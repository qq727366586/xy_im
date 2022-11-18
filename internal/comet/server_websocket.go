package comet

import (
	"log"
	"net"
	"xy_im/pkg/bytes"
	timer "xy_im/pkg/timer/pile"
)

var (
	maxInt = 1<<31 - 1
)

// 初始化ws
func InitWebsocket(server *Server, addrs []string, accept int) (err error) {
	var (
		bind     string
		addr     *net.TCPAddr
		listener *net.TCPListener
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			return
		}
		log.Printf("ws start listen: %s", bind)
		for i := 0; i < accept; i++ {
			go acceptWebsocket(server, listener)
		}
	}
	return
}

// 接收ws请求
func acceptWebsocket(server *Server, listener *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = listener.AcceptTCP(); err != nil {
			return
		}
		if err = conn.SetKeepAlive(server.c.TCP.KeepAlive); err != nil {
			return
		}
		if err = conn.SetReadBuffer(server.c.TCP.Rcvbuf); err != nil {
			return
		}
		if err = conn.SetWriteBuffer(server.c.TCP.Sndbuf); err != nil {
			return
		}
		go serveWebsocket(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveWebsocket(s *Server, conn net.Conn, r int) {
	var (
		tr = s.round.Timer(r)
		rp = s.round.Reader(r)
		wp = s.round.Writer(r)
	)
	s.ServeWebsocket(conn, rp, wp, tr)
}

func (s *Server) ServeWebsocket(conn net.Conn, rp, wp *bytes.Pool, tr *timer.Timer) {

}
