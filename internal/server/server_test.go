package server

import (
	"bytes"
	"errors"
	"httpfromtcp/internal/request"
	"io"
	"log"
	"net"
	"strings"
	"testing"
)

func TestListenHandlesAcceptErrors(t *testing.T) {
	tests := []struct {
		name    string
		errors  []error
		wantLog string
	}{
		{
			name:   "closed listener",
			errors: []error{net.ErrClosed},
		},
		{
			name:    "unexpected accept error",
			errors:  []error{errors.New("accept failed"), net.ErrClosed},
			wantLog: "Error accepting connection: accept failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			originalOutput := log.Writer()
			log.SetOutput(&output)
			defer log.SetOutput(originalOutput)

			listener := net.Listener(&errorListener{errors: tt.errors})
			server := &Server{conn: &listener}
			server.listen(func(w io.Writer, req *request.Request) *HandlerError {
				return nil
			})

			got := output.String()
			if tt.wantLog == "" && got != "" {
				t.Errorf("log output = %q, want no output", got)
			} else if tt.wantLog != "" && !strings.Contains(got, tt.wantLog) {
				t.Errorf("log output = %q, want it to contain %q", got, tt.wantLog)
			}
		})
	}
}

type errorListener struct {
	errors []error
}

func (l *errorListener) Accept() (net.Conn, error) {
	err := l.errors[0]
	l.errors = l.errors[1:]
	return nil, err
}

func (*errorListener) Close() error {
	return nil
}

func (*errorListener) Addr() net.Addr {
	return &net.TCPAddr{}
}
