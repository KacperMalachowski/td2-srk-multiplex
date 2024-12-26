package testserver

import (
	"fmt"
	"net"
)

type TestTCPServer struct {
	addr             string
	listener         net.Listener
	ReceivedDataChan chan string
}

func New(addr string) *TestTCPServer {
	return &TestTCPServer{
		addr:             addr,
		ReceivedDataChan: make(chan string),
	}
}

func (s *TestTCPServer) Start() (string, error) {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return "", fmt.Errorf("failed to start test server: %s", err)
	}
	s.listener = l

	go func() {
		defer l.Close()

		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()

				buf := make([]byte, 1024)
				n, err := c.Read(buf)
				if err != nil {
					return
				}

				s.ReceivedDataChan <- string(buf[:n])
			}(conn)
		}
	}()

	return l.Addr().String(), nil
}

func (s *TestTCPServer) Stop() {
	s.listener.Close()
	close(s.ReceivedDataChan)
}
