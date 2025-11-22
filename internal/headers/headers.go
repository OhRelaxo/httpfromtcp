package headers

import (
	"bytes"
	"fmt"
	"strings"
)

func NewHeaders() Headers {
	return map[string]string{}
}

type Headers map[string]string

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}
	header := string(data[:idx])

	trim := strings.Trim(header, " ")
	parts := strings.SplitN(trim, ":", 2)
	key := parts[0]
	if ok := strings.HasSuffix(key, " "); ok {
		return 0, false, fmt.Errorf("header: %s is malformated", header)
	}
	value, _ := strings.CutPrefix(parts[1], " ")
	h.set(key, value)

	return idx + 2, false, nil
}

func (h Headers) set(key, value string) {
	h[key] = value
}
