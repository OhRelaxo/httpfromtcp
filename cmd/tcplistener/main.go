package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ohrelaxo/httpfromtcp/internal/request"
)

const (
	ip   = "127.0.0.1"
	port = "42069"
)

func main() {
	listener, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		log.Fatalf("failed to Listen to tcp connection on IP: %v and Port: %v\n error: %v", ip, port, err)
	}
	defer listener.Close()
	for {
		log.Println("starting tcp listener...")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to Accept message: %v", err)
			continue
		}
		log.Println("a connection has been accepted")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("request has failed: %v", err)
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
	}
}
