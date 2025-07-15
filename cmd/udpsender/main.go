package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// The program should start by using net.ResolveUDPAddr to resolve the address localhost:42069
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}

	// Use net.DialUDP to prepare a UDP connection, and defer the closing of the connection.
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println((err.Error()))
		os.Exit(1)
	}
	defer conn.Close()

	// Create a new bufio.Reader that reads from os.Stdin
	// Stdin := NewFile(uintptr(syscall.Stdin), "/dev/stdin")
	reader := bufio.NewReader(os.Stdin)

	// Start an infinite loop that:
	for {
		// Prints a > character to the console (to indicate that the program is ready for user input)
		fmt.Print("> ")

		// Reads a line from the bufio.Reader using reader.ReadString, and log any errors
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println((err.Error()))
		}

		// Writes the line to the UDP connection using conn.Write, and log any errors
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println((err.Error()))
		}
	}

}
