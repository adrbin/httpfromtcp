package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	httpheaders "httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069
const bufferSize = 1024

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

func handler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		message := `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
		writeResponse(w, response.BadRequest, message)
		return
	case "/myproblem":
		message := `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
		writeResponse(w, response.InternalServerError, message)
		return
	default:
		if path, ok := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin"); ok {
			httpbinHandler(w, path)
			return
		}
		message := `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
		writeResponse(w, response.Ok, message)
	}
}

func writeResponse(w *response.Writer, statusCode response.StatusCode, message string) {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	headers := response.GetDefaultHeaders(len(message))
	headers.Set("content-type", "text/html")
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	err = w.WriteBody([]byte(message))
	if err != nil {
		log.Printf("Error writing body: %v", err)
	}
}

func httpbinHandler(w *response.Writer, path string) {
	resp, err := http.Get("https://httpbin.org" + path)
	if err != nil {
		log.Printf("Error fetching httpbin: %v", err)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCode(resp.StatusCode))
	headers := response.GetDefaultHeaders(0)
	headers.Set("content-type", resp.Header.Get("Content-Type"))
	headers.Set("transfer-encoding", "chunked")
	headers.Set("trailer", "X-Content-SHA256, X-Content-Length")
	headers.Delete("content-length")
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	hasher := sha256.New()
	var bodyLength int64
	buffer := make([]byte, bufferSize)
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]
			_, _ = hasher.Write(chunk)
			bodyLength += int64(n)

			_, err = w.WriteChunkedBody(chunk)
			if err != nil {
				log.Printf("Error writing chunked body: %v", err)
				return
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				log.Printf("Error reading from httpbin response: %v", readErr)
				return
			}
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("Error writing chunked body done: %v", err)
		return
	}
	trailers := httpheaders.NewHeaders()
	trailers.Set("X-Content-SHA256", hex.EncodeToString(hasher.Sum(nil)))
	trailers.Set("X-Content-Length", strconv.FormatInt(bodyLength, 10))
	if err := w.WriteTrailers(trailers); err != nil {
		log.Printf("Error writing trailers: %v", err)
	}
}
