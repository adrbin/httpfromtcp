package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	BufferSize = 8
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

		c <- string(currentLine)
	}()
	return c
}

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	c := getLinesChannel(f)
	for line := range c {
		fmt.Printf("read: %s\n", line)
	}
}
