package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
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

		// Prints a message to the console that a connection has been accepted.
		fmt.Println("Connection accepted")

		// Uses our getLinesChannel function to read lines from the connection.
		for line := range getLinesChannel(conn) {
			// Prints each line to the console with no additional formatting (aside from a terminating newline).
			fmt.Println(line)
		}

		// Prints a message to the console that the connection has been closed when the channel is closed.
		fmt.Println("Connection closed")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	line := make(chan string)

	// Llegim de 8 en 8
	currentLine := ""
	go func() {
		defer f.Close()
		defer close(line)
		for {
			buf := make([]byte, 8)
			_, err := f.Read(buf)

			if err == io.EOF {
				break
			}

			// Ajuntem les parts
			parts := strings.Split(string(buf), "\n")

			// Note that if we only have one "part", we don't need to print, as we have not reached a new line yet.
			if len(parts) == 1 {
				currentLine = fmt.Sprintf("%s%s", currentLine, parts[0])
			} else {
				for i := 0; i < len(parts)-1; i++ {
					currentLine = fmt.Sprintf("%s%s", currentLine, parts[i])
					// fmt.Printf("read: %s\n", currentLine)
					line <- currentLine
					currentLine = ""
				}

				// Add the last "part" to the "current line" variable. Repeat until you reach the end of the file.
				currentLine = fmt.Sprintf("%s%s", currentLine, parts[len(parts)-1])
			}
		}

		// Once you're done reading the file, if there's anything left in the "current line" variable, print it in the same read: LINE format.
		// fmt.Printf("read: %s\n", currentLine)
		line <- currentLine
	}()

	return line
}
