package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/neixir/httpfromtcp/internal/headers"
	"github.com/neixir/httpfromtcp/internal/request"
	"github.com/neixir/httpfromtcp/internal/response"
	"github.com/neixir/httpfromtcp/internal/server"
)

const port = 42069

func Chapter7(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget

	switch target {
	case "/":
		sendHTMLResponse(w, response.StatusOk, "200 OK", "Success!", "Your request was an absolute banger.")

	case "/yourproblem":
		sendHTMLResponse(w, response.StatusBadRequest, "400 Bad Request", "400 Bad Request", "Your request honestly kinda sucked.")

	case "/myproblem":
		sendHTMLResponse(w, response.StatusInternalServerError, "500 Internal Server Error", "Internal Server Error", "Okay, you know what? This one is on me.")

	}
}

// Add a new proxy handler to your server that maps /httpbin/x to https://httpbin.org/x,
// supporting both proxying and chunked responsing.
func Chapter8(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget

	// I used the strings.HasPrefix and strings.TrimPrefix functions to handle routing and route parsing.
	if !strings.HasPrefix(target, "/httpbin") {
		return
	}

	path := strings.TrimPrefix(target, "/httpbin")
	httpbinUrl := fmt.Sprint("https://httpbin.org", path)

	// I used http.Get to make the request to httpbin.org and httpbinResponse.Body.Read to read the response body.
	// I used a buffer size of 1024 bytes, and then printed n on every call to Read so that I could see
	// how much data was being read.
	// Use n as your chunk size and write that chunk data back to the client as soon as you get it from httpbin.org.
	httpbinResponse, err := http.Get(httpbinUrl)
	if err != nil {
		// TODO Send error? Log error?
	}

	defer httpbinResponse.Body.Close()

	// Status Line
	err = w.WriteStatusLine(response.StatusCode(httpbinResponse.StatusCode)) // response.StatusOk)
	if err != nil {
		// TODO Send error? Log error?
	}

	// Be sure to remove the Content-Length header from the response,
	resHeaders := make(headers.Headers)
	for key, values := range httpbinResponse.Header {
		if strings.ToLower(key) != "content-length" {
			resHeaders[key] = strings.Join(values, ", ") // pq values es un array de strings
		}
	}

	// and add the Transfer-Encoding: chunked header.
	resHeaders["Transfer-Encoding"] = "chunked"

	err = w.WriteHeaders(resHeaders)
	if err != nil {
		// TODO Send error? Log error?
	}

	buf := make([]byte, 1024)
	for {
		n, err := httpbinResponse.Body.Read(buf)

		if err == io.EOF {
			w.WriteChunkedBodyDone()
			break
		}

		if err != nil {
			// TODO Send error? Log error?
		}

		if n > 0 {
			// fmt.Printf("Chunk of %d bytes\n", n)
			w.WriteChunkedBody(buf[:n]) // :n pq el buffer tindra coses velles i potser no l'omplim
		}
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
	server, err := server.Serve(port, Chapter8)
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
