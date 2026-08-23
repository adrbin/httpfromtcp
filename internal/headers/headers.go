package headers

import (
	"bytes"
	"fmt"
	"regexp"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	lineLength := bytes.Index(data, []byte("\r\n"))
	if lineLength == -1 {
		return 0, false, nil
	}
	if lineLength == 0 {
		return 2, true, nil
	}
	line := data[:lineLength]
	nameBytes, valueBytes, found := bytes.Cut(line, []byte(":"))
	if !found {
		return 0, false, fmt.Errorf("invalid header format: %s", string(line))
	}
	trimmedName := bytes.TrimSpace(nameBytes)
	if len(trimmedName) == 0 {
		return 0, false, fmt.Errorf("empty header name")
	}
	if len(trimmedName) != len(nameBytes) {
		return 0, false, fmt.Errorf("invalid whitespace in header name: %s", string(nameBytes))
	}
	if !regexp.MustCompile("^[A-Za-z0-9!#$%&'*+-.^_`|~]+$").Match(trimmedName) {
		return 0, false, fmt.Errorf("invalid characters in header name: %s", string(trimmedName))
	}
	name := string(bytes.ToLower(trimmedName))
	value := string(bytes.TrimSpace(valueBytes))
	h[name] = value
	return lineLength + 2, false, nil
}
