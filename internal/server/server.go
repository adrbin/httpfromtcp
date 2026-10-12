package server

import (
	"errors"
	"fmt"
	"log"
	"net"
)

type Server struct {
	conn *net.Listener
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{conn: &listener}
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
	conn.Write([]byte(`HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13

Hello World!"))`))
}
