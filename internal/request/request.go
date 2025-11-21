package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to ReadAll the given reader error: %v", err)
	}

	requestLine, err := parseRequestLine(data)
	if err != nil {
		return nil, err
	}

	return &Request{RequestLine: *requestLine}, nil

}

func parseRequestLine(data []byte) (*RequestLine, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, fmt.Errorf("could not find CRLF in request-line")
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, err
	}
	return requestLine, nil
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
			return nil, fmt.Errorf("the method is not formatted correctly")
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
