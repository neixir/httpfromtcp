package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/neixir/httpfromtcp/internal/request"
	"github.com/neixir/httpfromtcp/internal/server"
)

const port = 42069

// type HandlerFunc func(w io.Writer, req *request.Request) *HandlerError
func Chapter7(w io.Writer, req *request.Request) *server.HandlerError {
	target := req.RequestLine.RequestTarget

	switch target {
	// If the request target (path) is /yourproblem return a 400 and the message "Your problem is not my problem\n"
	case "/yourproblem":
		message := "Your problem is not my problem\n"
		w.Write([]byte(message))
		return &server.HandlerError{
			StatusCode: 400,
			Message:    message,
		}

	// If the request target (path) is /myproblem return a 500 and the message "Woopsie, my bad\n"
	case "/myproblem":
		message := "Woopsie, my bad\n"
		w.Write([]byte(message))
		return &server.HandlerError{
			StatusCode: 500,
			Message:    message,
		}

	// Otherwise, it should just write the string "All good, frfr\n" to the response body.
	default:
		w.Write([]byte("All good, frfr\n"))

	}

	return nil

}

func main() {
	server, err := server.Serve(port, Chapter7)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
