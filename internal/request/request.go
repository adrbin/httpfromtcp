package request

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	requestBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	requestText := string(requestBytes)
	requestLineText, _, found := strings.Cut(requestText, "\r\n")
	if !found {
		return nil, fmt.Errorf("invalid request format")
	}

	requestLine, err := parseRequestLine(requestLineText)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line format")
	}

	method := parts[0]

	if !regexp.MustCompile(`^[A-Z]+$`).MatchString(method) {
		return nil, fmt.Errorf("invalid request method")
	}

	httpVersion := strings.TrimPrefix(parts[2], "HTTP/")
	if httpVersion != "1.1" {
		return nil, fmt.Errorf("unsupported HTTP version: %s", httpVersion)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: parts[1],
		HttpVersion:   httpVersion,
	}, nil
}
