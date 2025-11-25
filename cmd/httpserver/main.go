package main

import (
	"crypto/sha256"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ohrelaxo/httpfromtcp/internal/headers"
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
	if ok := strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"); ok {
		proxyHandler(w, req)
		return
	}

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

func proxyHandler(w *response.Writer, req *request.Request) {
	targetRequestTarget := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	newTarget := "http://httpbin.org" + targetRequestTarget
	resp, err := http.Get(newTarget)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Proxing to: %v", newTarget)

	w.WriteStatusLine(200)
	headers := headerWithOutContentLength()
	headers.Set("Transfer-Encoding", "chunked")
	headers.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(headers)

	var respBody []byte
	buff := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buff)
		if n > 0 {
			_, err = w.WriteChunkedBody(buff[:n])
			if err != nil {
				log.Printf("error writing Chunked Body: %v\n", err)
			}
			respBody = append(respBody, buff[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Printf("error while reading from httpbin.org: %v\n", err)
			continue
		}
		w.WriteChunkedBodyDone()
		respBody = append(respBody, []byte("0\r\n\r\n")...)
	}

	trailers := headerWithOutContentLength()
	sha := sha256.Sum256(respBody)
	
	bodyLength := len(respBody)
	trailers.Set("X-Content-SHA256", string(sha))
	trailers.Set("X-Content-Length", string(bodyLength))
}

func headerWithOutContentLength() headers.Headers {
	headers := response.GetDefaultHeaders(0)
	headers.Delete("Content-Length")
	return headers
}
