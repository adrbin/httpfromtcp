package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	Address = ":42069"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", Address)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println("Error sending data:", err)
		}
	}
}
