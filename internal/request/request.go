package request

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const BufferSize = 8

type ParseState int

const (
	Initialized ParseState = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	parseState  ParseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	if r.parseState == Done {
		return 0, nil
	}
	bytesRead, requestLine, err := parseRequestLine(data)
	if err != nil {
		return bytesRead, err
	}
	if bytesRead == 0 {
		return 0, nil
	}
	r.RequestLine = *requestLine
	r.parseState = Done
	return bytesRead, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		parseState: Initialized,
	}

	readToIndex := 0
	buffer := make([]byte, BufferSize)

	for {
		n, err := reader.Read(buffer[readToIndex:])
		if n == 0 && err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if n == 0 {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
			continue
		}
		readToIndex += n
		readBytes, err := r.parse(buffer[:readToIndex])
		if err != nil {
			return nil, err
		}
		if readBytes > 0 {
			readToIndex = copy(buffer, buffer[readBytes:])
		}
	}

	return r, nil
}

func parseRequestLine(requestBytes []byte) (int, *RequestLine, error) {
	lineLength := bytes.Index(requestBytes, []byte("\r\n"))
	if lineLength == -1 {
		return 0, nil, nil
	}
	requestLine := requestBytes[:lineLength]
	parts := bytes.Split(requestLine, []byte(" "))
	if len(parts) != 3 {
		return lineLength, nil, fmt.Errorf("invalid request line format: %s", string(requestLine))
	}

	method := string(parts[0])

	if !regexp.MustCompile(`^[A-Z]+$`).MatchString(method) {
		return lineLength, nil, fmt.Errorf("invalid request method")
	}

	httpVersion := strings.TrimPrefix(string(parts[2]), "HTTP/")
	if httpVersion != "1.1" {
		return lineLength, nil, fmt.Errorf("unsupported HTTP version: %s", httpVersion)
	}

	return lineLength, &RequestLine{
		Method:        method,
		RequestTarget: string(parts[1]),
		HttpVersion:   httpVersion,
	}, nil
}
