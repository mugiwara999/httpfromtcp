package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Closed   atomic.Bool
}

func runConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte(
		"HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"Hello World!\n",
	))
}

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		Listener: listener,
	}

	go server.Listen()
	return server, nil
}

func (s *Server) Close() error {
	s.Closed.Store(true)

	return s.Listener.Close()
}

func (s *Server) Listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {

			if s.Closed.Load() {
				return
			}
			continue
		}

		go runConnection(conn)

	}
}
