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
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}
	value, _ := strings.CutPrefix(parts[1], " ")

	if h.Get(key) == "" {
		err = h.Set(key, value)
		if err != nil {
			return 0, false, err
		}
	} else {
		h.Put(key, value)
	}

	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) error {
	lower := strings.ToLower(key)
	for _, char := range lower {
		ok := isValidHeaderChar(char)
		if !ok {
			return fmt.Errorf("invalid header token found: %s", key)
		}
	}
	h[lower] = value
	return nil
}

func isValidHeaderChar(c rune) bool {
	if (c >= 'a' && c <= 'z') || c >= 'A' && c <= 'Z' || (c >= '0' && c <= '9') {
		return true
	}
	specialChars := "!#$%&'*+-.^_`|~"
	return strings.ContainsRune(specialChars, c)
}

func (h Headers) Get(key string) (value string) {
	return h[strings.ToLower(key)]
}

func (h Headers) Put(key, value string) {
	h[strings.ToLower(key)] = h[strings.ToLower(key)] + ", " + value
}

func (h Headers) Delete(key string) {
	delete(h, strings.ToLower(key))
}
