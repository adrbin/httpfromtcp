package server

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Err        error
}

func (he HandlerError) Error() string {
	return fmt.Sprintf("StatusCode: %d, Error: %v", he.StatusCode, he.Err)
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
		log.Printf("Error parsing request: %v", err)
		return
	}
	buffer := bytes.Buffer{}
	if s.handler != nil {
		handlerError := s.handler(&buffer, req)
		if handlerError != nil {
			writeErrorResponse(conn, handlerError)
			return
		}
	}
	response.WriteStatusLine(conn, response.Ok)
	response.WriteHeaders(conn, response.GetDefaultHeaders(buffer.Len()))
	_, err = io.Copy(conn, &buffer)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func writeErrorResponse(w io.Writer, handlerError *HandlerError) {
	response.WriteStatusLine(w, handlerError.StatusCode)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(handlerError.Err.Error())))
	w.Write([]byte(handlerError.Err.Error()))
}
