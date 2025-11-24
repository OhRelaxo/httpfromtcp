package server

import (
	"fmt"
	"log"
	"net"

	"github.com/ohrelaxo/httpfromtcp/internal/request"
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
	resp := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\n\r\nHello World!\n")
	_, err = conn.Write(resp)
	if err != nil {
		log.Printf("failed to write to connection: %v\n", err)
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
