package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.BadRequest,
			Err:        fmt.Errorf("Your problem is not my problem\n"),
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.InternalServerError,
			Err:        fmt.Errorf("Woopsie, my bad\n"),
		}
	default:
		_, err := fmt.Fprintf(w, "All good, frfr\n")
		if err != nil {
			return &server.HandlerError{
				StatusCode: response.InternalServerError,
				Err:        fmt.Errorf("Error writing response: %v", err),
			}
		}
		return nil
	}
}
