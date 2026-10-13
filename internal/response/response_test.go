package response

import (
	"bytes"
	"testing"

	"httpfromtcp/internal/headers"
)

func TestWriteChunkedBodyDoneAndTrailers(t *testing.T) {
	var output bytes.Buffer
	writer := NewWriter(&output)

	if err := writer.WriteStatusLine(Ok); err != nil {
		t.Fatalf("WriteStatusLine() error = %v", err)
	}
	if err := writer.WriteHeaders(headers.NewHeaders()); err != nil {
		t.Fatalf("WriteHeaders() error = %v", err)
	}
	if _, err := writer.WriteChunkedBody([]byte("body")); err != nil {
		t.Fatalf("WriteChunkedBody() error = %v", err)
	}
	if _, err := writer.WriteChunkedBodyDone(); err != nil {
		t.Fatalf("WriteChunkedBodyDone() error = %v", err)
	}

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", "hash")
	if err := writer.WriteTrailers(trailers); err != nil {
		t.Fatalf("WriteTrailers() error = %v", err)
	}

	want := "HTTP/1.1 200 OK\r\n\r\n4\r\nbody\r\n0\r\nx-content-sha256: hash\r\n\r\n"
	if got := output.String(); got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}
