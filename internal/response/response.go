package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/ohrelaxo/httpfromtcp/internal/headers"
)

type Writer struct {
	writer io.Writer
	status writerStatus
}

type writerStatus int

const (
	statusLine writerStatus = iota
	statusHeaders
	statusBody
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
		status: statusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.status != statusLine {
		return fmt.Errorf("error: response is getting written in wrong order, current status: %v", w.status)
	}
	codes := map[StatusCode]string{
		Ok:                  "OK",
		BadRequest:          "Bad Request",
		InternalServerError: "Internal Server Error",
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %v %v\r\n", statusCode, codes[statusCode])
	_, err := w.writer.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	w.status = statusHeaders
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()
	header.Set("Content-Length", strconv.Itoa(contentLen))
	header.Set("Connection", "close")
	header.Set("Content-Type", "text/plain")
	return header
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.status != statusHeaders {
		return fmt.Errorf("error: response is getting written in wrong order, current status: %v", w.status)
	}
	if headers == nil {
		headers = GetDefaultHeaders(0)
	}

	for k, v := range headers {
		_, err := w.writer.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.status = statusBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.status != statusBody {
		return 0, fmt.Errorf("error: response is getting written in wrong order, current status: %v", w.status)
	}
	return w.writer.Write(p)
}
