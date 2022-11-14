package protocol

import (
	"fmt"
	"net"
	"testing"
	"time"
	"xy_im/pkg/bufio"
	"xy_im/pkg/websocket"
)

func TestWs(t *testing.T) {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.FailNow()
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				t.Error(err)
			}
			rd := bufio.NewReader(conn)
			wr := bufio.NewWriter(conn)
			req, err := websocket.ReadRequest(rd)
			if err != nil {
				t.Error(err)
			}
			if req.RequestURI != "/sub" {
				t.Error(err)
			}
			ws, err := websocket.Upgrade(conn, rd, wr, req)
			if err != nil {
				t.Error(err)
			}
			go func() {
				n := 1
				for {
					op, p, err := ws.ReadMessage()
					fmt.Println(n)
					if err != nil {
						fmt.Println(err)
						t.FailNow()
					}
					fmt.Println(op, string(p), err)
					if err = ws.WriteMessage(websocket.TextMessage, []byte("你好"+fmt.Sprintf("%d", n))); err != nil {
						t.FailNow()
					}
					n++
					ws.Flush()
				}
			}()
		}
	}()
	time.Sleep(1000 * time.Minute)
}
