package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	Ok                  StatusCode = 200
	PartialContent      StatusCode = 206
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case Ok:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d OK\r\n", statusCode)
		return err
	case PartialContent:
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d Partial Content\r\n", statusCode)
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
	WritingTrailers
)

type Writer struct {
	w     io.Writer
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

func (w *Writer) WriteBody(p []byte) error {
	_, err := w.w.Write(p)
	return err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != WritingBody {
		return 0, fmt.Errorf("invalid state: expected WritingBody, got %v", w.state)
	}
	chunkSize := len(p)
	if chunkSize == 0 {
		n, err := fmt.Fprint(w.w, "0\r\n\r\n")
		return n, err
	}
	n, err := fmt.Fprintf(w.w, "%x\r\n", chunkSize)
	written := n
	if err != nil {
		return written, err
	}
	n, err = w.w.Write(p)
	written += n
	if err != nil {
		return written, err
	}
	n, err = fmt.Fprint(w.w, "\r\n")
	written += n
	return written, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != WritingBody {
		return 0, fmt.Errorf("invalid state: expected WritingBody, got %v", w.state)
	}
	n, err := fmt.Fprint(w.w, "0\r\n")
	if err == nil {
		w.state = WritingTrailers
	}
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != WritingTrailers {
		return fmt.Errorf("invalid state: expected WritingTrailers, got %v", w.state)
	}
	return WriteHeaders(w.w, h)
}
