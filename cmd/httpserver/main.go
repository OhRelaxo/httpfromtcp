package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ohrelaxo/httpfromtcp/internal/request"
	"github.com/ohrelaxo/httpfromtcp/internal/response"
	"github.com/ohrelaxo/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	var respMessage string
	var statusCode int
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		respMessage = "<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>"
		statusCode = 400
	case "/myproblem":
		respMessage = "<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>"
		statusCode = 500
	default:
		respMessage = "<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>"
		statusCode = 200
	}

	header := response.GetDefaultHeaders(len(respMessage))
	header.Set("Content-Type", "text/html")
	err := w.WriteStatusLine(response.StatusCode(statusCode))
	if err != nil {
		log.Println(err)
	}
	err = w.WriteHeaders(header)
	if err != nil {
		log.Println(err)
	}
	_, err = w.WriteBody([]byte(respMessage))
	if err != nil {
		log.Println(err)
	}
}
