package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/ohrelaxo/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	codes := map[StatusCode]string{
		Ok:                  "OK",
		BadRequest:          "Bad Request",
		InternalServerError: "Internal Server Error",
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %v %v\r\n", statusCode, codes[statusCode])
	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()
	header.Set("Content-Length", strconv.Itoa(contentLen))
	header.Set("Connection", "close")
	header.Set("Content-Type", "text/plain")
	return header
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := w.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
