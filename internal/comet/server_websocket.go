package comet

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
	"xy_im/api/protocol"
	"xy_im/pkg/bytes"
	timer "xy_im/pkg/timer/pile"
	"xy_im/pkg/websocket"
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

func (s *Server) RandServerHeartBeat() time.Duration {
	return minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat)))
}

func (s *Server) ServeWebsocket(conn net.Conn, rp, wp *bytes.Pool, tr *timer.Timer) {
	lastHB := time.Now()
	// 初始化通信通道
	ch := NewChannel(s.c.Protocol.CliProto, s.c.Protocol.SvrProto)
	c, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1.handshake
	step := 1
	trd := tr.Add(time.Duration(s.c.Protocol.HandshakeTimeout), func() {
		// tls close block
		_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		_ = conn.Close()
	})

	// 2.request
	step = 2
	var (
		req *websocket.Request
		err error
	)
	rr := &ch.Reader
	rb := rp.Get()
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.IP, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	if req, err = websocket.ReadRequest(rr); err != nil || req.RequestURI != "/sub" {
		conn.Close()
		tr.Del(trd)
		rp.Put(rb)
		if err != io.EOF {
			fmt.Println("2. http.Request(rr) err: ", err)
		}
		return
	}

	// 3.upgrade
	step = 3
	var (
		ws *websocket.Conn
	)
	wr := &ch.Writer
	wb := wp.Get()
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	if ws, err = websocket.Upgrade(conn, rr, wr, req); err != nil {
		conn.Close()
		tr.Del(trd)
		rp.Put(rb)
		wp.Put(wb)
		if err != io.EOF {
			fmt.Println("3. websocket.Upgrade(conn, rr, wr, req) err: ", err)
		}
		return
	}

	// 4.auth
	step = 4
	var (
		rid     string
		accepts []int32
		b       *Bucket
		hb      time.Duration
		p       *protocol.Proto
	)
	if p, err = ch.CliProto.Set(); err == nil {
		if ch.Mid, ch.Key, rid, accepts, hb, err = s.authWebsocket(c, ws, p, req.Header.Get("Cookie")); err != nil {
			ch.Watch(accepts...)
			b = s.Bucket(ch.Key)
			err = b.Put(rid, ch)
		}
	}
	if err != nil {
		ws.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		if err != io.EOF && err != websocket.ErrMessageClose {
			fmt.Println("4. handshake err: ", err)
		}
	}

	// 5.dispatch
	trd.Key = ch.Key
	tr.Set(trd, hb)
	go s.dispatchWebsocket(ws, wp, wb, ch)
	serverHeartBeat := s.RandServerHeartBeat()
	// 监听输入
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if err = p.ReadWebSocket(ws); err != nil {
			break
		}
		if p.Op == protocol.OpHeartbeat {
			// 重新计时
			tr.Set(trd, hb)
			p.Op = protocol.OpHeartbeatReply
			p.Body = nil
			if now := time.Now(); now.Sub(lastHB) > serverHeartBeat {
				if er := s.Heartbeat(c, ch.Mid, ch.Key); er == nil {
					lastHB = now
				}
			}
			step++
		} else {
			// todo
		}
		ch.CliProto.SetAdv()
		ch.Signal()
	}
	if err != nil && err != io.EOF && err != websocket.ErrMessageClose && !strings.Contains(err.Error(), "closed") {
		fmt.Println("5. server ws err: ", err)
	}
	b.Del(ch)
	tr.Del(trd)
	ws.Close()
	ch.Close()
	rp.Put(rb)
	if err = s.Disconnect(c, ch.Mid, ch.Key); err != nil {
		fmt.Println("5. s.Disconnect err: ", err)
	}
}

// auth授权
func (s *Server) authWebsocket(c context.Context, ws *websocket.Conn, p *protocol.Proto, cookie string) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) {
	for {
		if err = p.ReadWebSocket(ws); err != nil {
			return
		}
		if p.Op == protocol.OpAuth {
			break
		}
	}
	if mid, key, rid, accepts, hb, err = s.Connect(c, p, cookie); err != nil {
		return
	}
	p.Op = protocol.OpAuthReply
	p.Body = nil
	if err = p.WriteWebsocket(ws); err != nil {
		return
	}
	err = ws.Flush()
	return
}

// dispatch
func (s *Server) dispatchWebsocket(ws *websocket.Conn, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	// 接收数据
	p := ch.Ready()
	var (
		finish bool
		online int32
		err    error
	)
	switch p {
	// 结束
	case protocol.ProtoFinish:
		finish = true
		goto failed
	// 输出参数
	case protocol.ProtoReady:
		for {
			if p, err = ch.CliProto.Get(); err != nil {
				break
			}
			// 如果是心跳reply
			if p.Op == protocol.OpHeartbeatReply {
				if ch.Room != nil {
					online = ch.Room.OnlineNum()
				}
				if err = p.WriteWebsocketHeart(ws, online); err != nil {
					goto failed
				}
			} else {
				if err = p.WriteWebsocket(ws); err != nil {
					goto failed
				}
			}
		}
		p.Body = nil // 避免内存泄漏
		ch.CliProto.GetAdv()
	default:
		if err = p.WriteWebsocket(ws); err != nil {
			goto failed
		}
		// 饥饿发送
		if err = ws.Flush(); err != nil {
			break
		}
	}
failed:
	if err != nil && err != io.EOF && err != websocket.ErrMessageClose {
		fmt.Println("dispatch err: ", err)
	}
	ws.Close()
	wp.Put(wb)
	for !finish {
		finish = ch.Ready() == protocol.ProtoFinish
	}
}
