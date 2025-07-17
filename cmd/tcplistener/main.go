package main

import (
	"fmt"
	"net"
	"os"

	"github.com/neixir/httpfromtcp/internal/request"
)

func main() {
	// Use net.Listen to set up a TCP listener on port :42069.
	l, err := net.Listen("tcp", ":42069") // 127.0.0.1
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Don't forget to .Close() the listener when the program exits.
	defer l.Close()

	// With the listener created, start an infinite loop that:
	for {
		// Starts by waiting to .Accept a connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range r.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

	}
}
