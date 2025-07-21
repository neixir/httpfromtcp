package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/neixir/httpfromtcp/internal/headers"
	"github.com/neixir/httpfromtcp/internal/request"
	"github.com/neixir/httpfromtcp/internal/response"
	"github.com/neixir/httpfromtcp/internal/server"
)

const port = 42069

// type HandlerFunc func(w *response.Writer, req *request.Request)
func Chapter7(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget

	switch target {
	case "/":
		sendHTMLResponse(w, response.StatusOk, "200 OK", "Success!", "Your request was an absolute banger.")

	case "/yourproblem":
		sendHTMLResponse(w, response.StatusBadRequest, "400 Bad Request", "400 Bad Request", "Your request honestly kinda sucked.")

	case "/myproblem":
		sendHTMLResponse(w, response.StatusInternalServerError, "500 Internal Server Error", "Internal Server Error", "Okay, you know what? This one is on me.")

		// Otherwise...?
		// default:
		// 	w.Write([]byte("All good, frfr\n"))

	}

}

func sendHTMLResponse(w *response.Writer, status response.StatusCode, title, h1, msg string) {
	htmlTemplate := `<html>
  <head>
    <title>%s</title>
  </head>
  <body>
    <h1>%s</h1>
    <p>%s</p>
  </body>
</html>`

	htmlBody := fmt.Sprintf(htmlTemplate, title, h1, msg)
	w.WriteStatusLine(status)

	resHeaders := response.GetDefaultHeaders(len(htmlBody))
	headers.OverwriteHeader(resHeaders, "Content-Type", "text/html")
	w.WriteHeaders(resHeaders)

	w.WriteBody([]byte(htmlBody))
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
