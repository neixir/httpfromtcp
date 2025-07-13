package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	// Obrim el fitxer
	f, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Llegim de 8 en 8
	for {
		buf := make([]byte, 8)
		_, err := f.Read(buf)

		if err == io.EOF {
			break
		}

		fmt.Printf("read: %s\n", buf)
	}

}
