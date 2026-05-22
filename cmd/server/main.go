package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("Incoming LOG: ", scanner.Text())
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, _ := listener.Accept()
		go handleConnection(conn)
	}
}
