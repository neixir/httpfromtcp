package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	// Obrim el fitxer
	f, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for line := range getLinesChannel(f) {
		fmt.Printf("read: %s\n", line)
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
