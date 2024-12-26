package tcp

import (
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	listener net.Listener
	outChan  chan []byte
	inChan   chan []byte
	exit     chan struct{}
	clients  []net.Conn
}

func NewServer(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: listener,
		outChan:  make(chan []byte),
		inChan:   make(chan []byte),
		exit:     make(chan struct{}, 1),
	}, nil
}

func NewServerFromPortRange(ip string, minPort, maxPort int) (*Server, error) {
	for port := minPort; port <= maxPort; port++ {
		server, err := NewServer(fmt.Sprintf("%s:%d", ip, port))
		if err == nil {
			return server, nil
		}
	}
	return nil, fmt.Errorf("no available port in range")
}

func (s *Server) Start() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			break
		}
		s.clients = append(s.clients, conn)
		go s.handleConn(conn)
	}
}

func (s *Server) Stop() {
	s.exit <- struct{}{}
	s.listener.Close()
}

func (s *Server) IsRunning() bool {
	return s.listener != nil
}

func (s *Server) HasClients() bool {
	return len(s.clients) > 0
}

func (s *Server) Send(msg []byte) error {
	var errs []error
	for _, conn := range s.clients {
		_, err := conn.Write(msg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error sending message to clients: %v", errs)
	}

	return nil
}

func (s *Server) Receive() []byte {
	return <-s.inChan
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("client disconnected: %v", conn.RemoteAddr())
				s.removeClient(conn)
			} else {
				log.Printf("error reading from connection: %v", err)
			}
			return
		}
		s.inChan <- buf[:n]

		select {
		case <-s.exit:
			return
		default:
		}
	}
}

func (s *Server) removeClient(conn net.Conn) {
	for i, c := range s.clients {
		if c == conn {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}
