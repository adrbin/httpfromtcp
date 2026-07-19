package main

import (
	"fmt"
	"os"
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
		fmt.Printf("read: %s\n", string(bytes[:n]))
	}
}
