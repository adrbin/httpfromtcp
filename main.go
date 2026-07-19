package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	BufferSize = 8
)

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()
	currentLine := ""
	bytes := make([]byte, BufferSize)

	for {
		n, err := f.Read(bytes)
		if n == 0 {
			break
		}
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		chunk := string(bytes[:n])
		parts := strings.Split(chunk, "\n")
		lastPartIndex := len(parts) - 1
		for _, part := range parts[:lastPartIndex] {
			currentLine += part
			fmt.Printf("read: %s\n", string(currentLine))
			currentLine = ""
		}
		currentLine += parts[lastPartIndex]
	}

	fmt.Printf("read: %s\n", currentLine)
}
