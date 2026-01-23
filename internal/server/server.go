package server

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/mugiwara999/httpfromtcp/internal/request"
	"github.com/mugiwara999/httpfromtcp/internal/response"
)

type Server struct {
	Listener net.Listener
	Closed   atomic.Bool
	Handler  Handler
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	Status  response.StatusCode
	Message string
}

func (s *Server) runConnection(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		WriteHandlerError(conn, &HandlerError{
			Status:  response.StatusBadRequest,
			Message: "Bad Request",
		})
		return
	}
	if req == nil {
		WriteHandlerError(conn, &HandlerError{
			Status:  response.StatusBadRequest,
			Message: "Bad Request",
		})
		return
	}

	if herr := s.Handler(conn, req); herr != nil {
		WriteHandlerError(conn, herr)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		Listener: listener,
		Handler:  handler,
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

		go s.runConnection(conn)

	}
}
