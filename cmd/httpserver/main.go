package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := "http://httpbin.org" + target
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Proxing to: %v", url)

	w.WriteStatusLine(response.Ok)
	header := response.GetDefaultHeaders(0)
	header.Delete("Content-Length")
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(header)

	var respBody []byte
	buff := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buff)
		log.Printf("Read %d Bytes", n)
		if n > 0 {
			_, err = w.WriteChunkedBody(buff[:n])
			if err != nil {
				log.Printf("error writing Chunked Body: %v\n", err)
				break
			}
			respBody = append(respBody, buff[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error reading response body: %v\n", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("error writing chunked body done: %v\n", err)
	}

	sha := sha256.Sum256(respBody)
	log.Printf("%x\n", sha)

	bodyLength := len(respBody)
	log.Println(bodyLength)

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sha))
	trailers.Set("X-Content-Length", strconv.Itoa(bodyLength))
	err = w.WriteTrailers(trailers)
	if err != nil {
		log.Printf("error while writing Trailers: %v", err)
	}
}
