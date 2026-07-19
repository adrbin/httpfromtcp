package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	BufferSize = 8
	Address    = ":42069"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string)
	go func() {
		defer close(c)
		defer f.Close()
		currentLine := ""
		bytes := make([]byte, BufferSize)
		for {
			n, err := f.Read(bytes)
			if n == 0 {
				break
			}
			if err != nil {
				return
			}
			chunk := string(bytes[:n])
			parts := strings.Split(chunk, "\n")
			lastPartIndex := len(parts) - 1
			for _, part := range parts[:lastPartIndex] {
				currentLine += part
				c <- string(currentLine)
				currentLine = ""
			}
			currentLine += parts[lastPartIndex]
		}
		if currentLine != "" {
			c <- string(currentLine)
		}
	}()
	return c
}

func main() {
	f, err := net.Listen("tcp", Address)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()
	for {
		conn, err := f.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Printf("Connection accepted from %s\n", conn.RemoteAddr())
		c := getLinesChannel(conn)
		for line := range c {
			fmt.Printf("%s\n", line)
		}
		fmt.Printf("Connection closed from %s\n", conn.RemoteAddr())
	}

}
