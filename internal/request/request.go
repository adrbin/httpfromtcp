package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const BufferSize = 8

type ParseState int

const (
	Initialized ParseState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	parseState  ParseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func NewRequest() *Request {
	return &Request{
		parseState: Initialized,
		Headers:    headers.NewHeaders(),
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := NewRequest()

	readToIndex := 0
	buffer := make([]byte, BufferSize)

	for r.parseState != Done {
		n, err := reader.Read(buffer[readToIndex:])
		if n == 0 && errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("unexpected EOF while parsing request")
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
			readToIndex = copy(buffer, buffer[readBytes:readToIndex])
		}
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.parseState != Done {
		bytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed + bytesParsed, err
		}
		if bytesParsed == 0 {
			break
		}
		totalBytesParsed += bytesParsed
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.parseState {
	case Initialized:
		bytesRead, requestLine, err := parseRequestLine(data)
		if err != nil {
			return bytesRead, err
		}
		if bytesRead == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.parseState = ParsingHeaders
		return bytesRead, nil
	case ParsingHeaders:
		bytesRead, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.parseState = ParsingBody
		}
		return bytesRead, nil
	case ParsingBody:
		contentLengthStr := r.Headers.Get("content-length")
		if contentLengthStr == "" {
			r.parseState = Done
			return 0, nil
		}
		bytesLength := len(data)
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return bytesLength, fmt.Errorf("invalid content-length header value: %s", contentLengthStr)
		}
		r.Body = slices.Concat(r.Body, data)
		if len(r.Body) > contentLength {
			return bytesLength, fmt.Errorf("body exceeds content-length")
		}
		if len(r.Body) == contentLength {
			r.parseState = Done
		}
		return bytesLength, nil
	case Done:
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown parse state: %v", r.parseState)
	}
}

func parseRequestLine(requestBytes []byte) (int, *RequestLine, error) {
	lineLength := bytes.Index(requestBytes, []byte("\r\n"))
	if lineLength == -1 {
		return 0, nil, nil
	}
	bytesRead := lineLength + 2
	requestLine := requestBytes[:lineLength]
	parts := bytes.Split(requestLine, []byte(" "))
	if len(parts) != 3 {
		return bytesRead, nil, fmt.Errorf("invalid request line format: %s", string(requestLine))
	}

	method := string(parts[0])

	if !regexp.MustCompile(`^[A-Z]+$`).MatchString(method) {
		return bytesRead, nil, fmt.Errorf("invalid request method")
	}

	httpVersion := strings.TrimPrefix(string(parts[2]), "HTTP/")
	if httpVersion != "1.1" {
		return bytesRead, nil, fmt.Errorf("unsupported HTTP version: %s", httpVersion)
	}

	return bytesRead, &RequestLine{
		Method:        method,
		RequestTarget: string(parts[1]),
		HttpVersion:   httpVersion,
	}, nil
}
