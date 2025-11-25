package server

import (
	"bytes"
	"fmt"
	"io"
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

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type serverState int

const (
	listening serverState = iota
	closed
)

func (h *HandlerError) WriteHandlerError(w io.Writer) error {
	err := response.WriteStatusLine(w, h.StatusCode)
	if err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(h.Message))
	err = response.WriteHeaders(w, headers)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(h.Message))
	if err != nil {
		return err
	}
	return nil
}

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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("request failed: %v\n", err)
		hErr := &HandlerError{
			StatusCode: response.BadRequest,
			Message:    err.Error(),
		}
		hErr.WriteHandlerError(conn)
		return
	}

	buff := bytes.Buffer{}
	handlerErr := s.handler(&buff, req)
	if handlerErr != nil {
		handlerErr.WriteHandlerError(conn)
		return
	}

	err = response.WriteStatusLine(conn, response.Ok)
	if err != nil {
		log.Printf("failed to write status line: %v", err)
		return
	}
	headers := response.GetDefaultHeaders(len(buff.Bytes()))
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("failed to wirte headers: %v", err)
		return
	}
	conn.Write(buff.Bytes())
}

func (s *Server) Close() error {
	s.state = closed
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
