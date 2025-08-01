package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
		log.Fatalf("error getting httpbin: %v", err)
	}

	defer httpbinResponse.Body.Close()

	// Status Line
	err = w.WriteStatusLine(response.StatusCode(httpbinResponse.StatusCode)) // response.StatusOk)
	if err != nil {
		log.Fatalf("error writing status line: %v", err)
	}

	// Be sure to remove the Content-Length header from the response,
	resHeaders := make(headers.Headers)
	for key, values := range httpbinResponse.Header {
		if strings.ToLower(key) != "content-length" {
			resHeaders[key] = strings.Join(values, ", ") // pq values es un array de strings
		}
	}

	// and add the Transfer-Encoding: chunked header
	resHeaders["Transfer-Encoding"] = "chunked"

	// Announce X-Content-SHA256 and X-Content-Length as trailers in the Trailer header.
	resHeaders["Trailer"] = "X-Content-SHA256, X-Content-Length"

	err = w.WriteHeaders(resHeaders)
	if err != nil {
		log.Fatalf("error writing headers: %v", err)
	}

	buf := make([]byte, 1024)
	fullbody := []byte{}
	// fullbodyLength := 0

	for {
		n, err := httpbinResponse.Body.Read(buf)

		if err == io.EOF {
			hash := sha256.Sum256([]byte(fullbody))
			trailers := headers.Headers{ //map[string]string{
				"X-Content-SHA256": fmt.Sprintf("%x", hash),
				"X-Content-Length": strconv.Itoa(len(fullbody)),
			}

			w.WriteChunkedBodyDone(trailers)
			break
		}

		if err != nil {
			log.Fatalf("error reading chunk: %v", err)
		}

		if n > 0 {
			// Keep track of the full response body as you read it in chunks from the httpbin server
			fullbody = append(fullbody, buf[:n]...)
			// fullbodyLength += n

			// fmt.Printf("Chunk of %d bytes\n", n)
			w.WriteChunkedBody(buf[:n]) // :n pq el buffer tindra coses velles i potser no l'omplim
		}
	}

}

func Chapter9(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget

	if target != "/video" {
		return
	}

	// Llegim el fitxer
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	// Status Line
	err = w.WriteStatusLine(response.StatusOk)
	if err != nil {
		log.Fatalf("error writing status line: %v", err)
	}

	// Be sure to remove the Content-Length header from the response,
	resHeaders := response.GetDefaultHeaders(len(data))

	headers.OverwriteHeader(resHeaders, "Content-Type", "video/mp4")

	err = w.WriteHeaders(resHeaders)
	if err != nil {
		log.Fatalf("error writing headers: %v", err)
	}

	w.WriteBody([]byte(data))

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
	server, err := server.Serve(port, Chapter9)
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
