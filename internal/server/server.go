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
	handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

type serverState int

const (
	listening serverState = iota
	closed
)

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to Listen on port: %v, error: %v", port, err)
		return nil, err
	}
	s := &Server{
		state:    listening,
		listener: listener,
		handler:  handler,
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
	writer := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("request failed: %v\n", err)
		writer.WriteStatusLine(response.BadRequest)
		body := fmt.Appendf(nil, "Error parsing request: %v", err)
		writer.WriteHeaders(response.GetDefaultHeaders(len(body)))
		writer.WriteBody(body)
		return
	}

	s.handler(writer, req)
	return
}

func (s *Server) Close() error {
	s.state = closed
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
