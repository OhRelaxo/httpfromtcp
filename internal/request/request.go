package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/ohrelaxo/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	status requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	initialized requestState = iota
	parsingHeaders
	parsingBody
	done
)

const (
	crlf       = "\r\n"
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize)
	readToIndex := 0
	request := Request{status: initialized, Headers: make(headers.Headers), Body: make([]byte, 0)}
	for request.status != done {
		if readToIndex >= len(buffer) {
			tempBuffer := make([]byte, len(buffer)*2)
			copy(tempBuffer, buffer)
			buffer = tempBuffer
		}

		bytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.status != done {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d, bytes not read: %v", request.status, bytesRead, readToIndex)
				}
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesConsumed, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buffer, buffer[bytesConsumed:])
		readToIndex -= bytesConsumed

	}

	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.status {
	case initialized:
		consumed, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}

		r.status = parsingHeaders
		r.RequestLine = *requestLine
		return consumed, nil
	case parsingHeaders:
		bytesUsed := 0
		for r.status != parsingBody {
			consumed, parseDone, err := r.Headers.Parse(data[bytesUsed:])
			if err != nil {
				return 0, err
			}
			if parseDone {
				if contentLength := r.Headers.Get("Content-Length"); contentLength == "" || contentLength == "0" {
					r.status = done
				} else {
					r.status = parsingBody
				}
			}
			if consumed == 0 {
				break
			}
			bytesUsed += consumed
		}
		return bytesUsed, nil
	case parsingBody:
		length, err := strconv.Atoi(r.Headers.Get("Content-Length"))
		if err != nil {
			return 0, err
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > length {
			fmt.Printf("content length: %v\nlen(data): %v\n", length, len(data))
			return 0, fmt.Errorf("error: content-lenght header is too small for body")
		}
		if length == len(r.Body) {
			r.status = done
		}
		return len(data), nil

	case done:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error:  unkown state")
	}
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, nil, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return 0, nil, err
	}
	return idx + 2, requestLine, nil
}

func requestLineFromString(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) > 3 {
		return nil, fmt.Errorf("the request-line contains too many parts")
	}

	method := parts[0]
	for _, char := range method {
		upper := unicode.IsUpper(char)
		letter := unicode.IsLetter(char)
		if !upper || !letter {
			return nil, fmt.Errorf("the method is not formatted correctly: %s", requestLine)
		}
	}

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", requestLine)
	}
	if httpPart := versionParts[0]; httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	if version := versionParts[1]; version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		HttpVersion:   versionParts[1],
		RequestTarget: parts[1],
		Method:        method,
	}, nil
}
