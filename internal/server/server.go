package server

import (
	"errors"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Err        error
}

func (he HandlerError) Error() string {
	return fmt.Sprintf("StatusCode: %d, Error: %v", he.StatusCode, he.Err)
}

func (he *HandlerError) writeErrorResponse(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(he.Err.Error())))
	w.Write([]byte(he.Err.Error()))
}

type Server struct {
	conn    *net.Listener
	handler Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	if handler == nil {
		return nil, errors.New("handler cannot be nil")
	}

	server := &Server{conn: &listener, handler: handler}
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	if s.conn != nil {
		return (*s.conn).Close()
	}
	return nil
}

func (s *Server) listen() {
	for s.conn != nil {
		conn, err := (*s.conn).Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := &HandlerError{
			StatusCode: response.BadRequest,
			Err:        fmt.Errorf("Error parsing request: %v", err),
		}
		handlerError.writeErrorResponse(conn)
		return
	}
	responseWriter := response.NewWriter(conn)
	s.handler(responseWriter, req)
}
