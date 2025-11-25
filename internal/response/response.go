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
	statusTrailer
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
		return fmt.Errorf("error: response: %v is getting written in wrong order, current status: %v", statusLine, w.status)
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
	defer func() { w.status = statusHeaders }()
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()
	header.Set("Content-Length", strconv.Itoa(contentLen))
	header.Set("Connection", "close")
	header.Set("Content-Type", "text/plain")
	return header
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.status != statusHeaders {
		return fmt.Errorf("error: response: %v is getting written in wrong order, current status: %v", statusHeaders, w.status)
	}
	if h == nil {
		h = GetDefaultHeaders(0)
	}

	defer func() { w.status = statusBody }()
	return w.processHeadersOrTrailers(h)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.status != statusBody {
		return 0, fmt.Errorf("error: response: %v is getting written in wrong order, current status: %v", statusBody, w.status)
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.status != statusBody {
		return 0, fmt.Errorf("error: response: %v is getting written in wrong order, current status: %v", statusBody, w.status)
	}

	lenData := len(p)
	hex := strconv.FormatInt(int64(lenData), 16)
	body := hex + "\r\n" + string(p) + "\r\n"
	return w.writer.Write([]byte(body))
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.status != statusBody {
		return 0, fmt.Errorf("error: response: %v is getting written in wrong order, current status: %v", statusBody, w.status)
	}
	defer func() { w.status = statusTrailer }()
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.status != statusTrailer {
		return fmt.Errorf("error: response: %v (Trailers) is getting written in wrong order, current status: %v", statusTrailer, w.status)
	}

	/*
		_, err := w.writer.Write([]byte("0\r\n"))
		if err != nil {
			return err
		}

	*/

	return w.processHeadersOrTrailers(h)
}

func (w *Writer) processHeadersOrTrailers(h headers.Headers) error {
	for k, v := range h {
		_, err := w.writer.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
