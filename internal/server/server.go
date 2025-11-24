package server

import (
	"fmt"
	"log"
	"net"

	"github.com/ohrelaxo/httpfromtcp/internal/request"
	"github.com/ohrelaxo/httpfromtcp/internal/response"
)

type Server struct {
	state    serverState
	listener net.Listener
}

type serverState int

const (
	listening serverState = iota
	closed
)

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to Listen on port: %v, error: %v", port, err)
		return nil, err
	}
	s := &Server{
		state:    listening,
		listener: listener,
	}
	go s.listen()
	return s, err
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.state == closed {
				return
			}
			log.Printf("failed to accept connection: %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	_, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("request failed: %v\n", err)
		return
	}
	err = response.WriteStatusLine(conn, response.Ok)
	if err != nil {
		log.Printf("failed to write status line: %v", err)
		return
	}
	headers := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("failed to wirte headers: %v", err)
		return
	}
	return
}

func (s *Server) Close() error {
	s.state = closed
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
