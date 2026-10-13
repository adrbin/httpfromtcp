package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	Ok StatusCode = 200
	BadRequest StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case Ok:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d OK\r\n", statusCode)
		return err
	case BadRequest:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d Bad Request\r\n", statusCode)
		return err
	case InternalServerError:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d Internal Server Error\r\n", statusCode)
		return err
	default:
		return fmt.Errorf("unknown status code: %d", statusCode)
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = fmt.Sprintf("%d", contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for name, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", name, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	return err
}

type WriterState int

const (
	WritingStatusLine WriterState = iota
	WritingHeaders
	WritingBody
)

type Writer struct {
	w io.Writer
	state WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w, state: WritingStatusLine}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != WritingStatusLine {
		return fmt.Errorf("invalid state: expected WritingStatusLine, got %v", w.state)
	}
	err := WriteStatusLine(w.w, statusCode)
	if err == nil {
		w.state = WritingHeaders
	}
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != WritingHeaders {
		return fmt.Errorf("invalid state: expected WritingHeaders, got %v", w.state)
	}
	err := WriteHeaders(w.w, headers)
	if err == nil {
		w.state = WritingBody
	}
	return err
}

func (w *Writer) WriteBody(p []byte) (error) {
	_, err := w.w.Write(p)
	return err
}