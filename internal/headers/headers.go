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
	bytesRead := lineLength + 2
	if lineLength == 0 {
		return bytesRead, true, nil
	}
	line := data[:lineLength]
	nameBytes, valueBytes, found := bytes.Cut(line, []byte(":"))
	if !found {
		return bytesRead, false, fmt.Errorf("invalid header format: %s", string(line))
	}
	trimmedName := bytes.TrimSpace(nameBytes)
	if len(trimmedName) == 0 {
		return bytesRead, false, fmt.Errorf("empty header name")
	}
	if len(trimmedName) != len(nameBytes) {
		return bytesRead, false, fmt.Errorf("invalid whitespace in header name: %s", string(nameBytes))
	}
	if !regexp.MustCompile("^[A-Za-z0-9!#$%&'*+-.^_`|~]+$").Match(trimmedName) {
		return bytesRead, false, fmt.Errorf("invalid characters in header name: %s", string(trimmedName))
	}
	name := string(bytes.ToLower(trimmedName))
	value := string(bytes.TrimSpace(valueBytes))
	if existingValue, exists := h[name]; exists {
		value = existingValue + ", " + value
	}
	h[name] = value
	return bytesRead, false, nil
}
