package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/neixir/httpfromtcp/internal/headers"
)

// Create a new StatusCode "enum" type (you know, the fake Go enums). Define constants for the only 3 codes we're going to worry about:
type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

// It should map the given status code to the correct reason phrase, if it's one of the 3 that we support:
// 200 should return HTTP/1.1 200 OK
// 400 should return HTTP/1.1 400 Bad Request
// 500 should return HTTP/1.1 500 Internal Server Error
// Any other code should just leave the reason phrase blank.
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error = nil

	switch statusCode {
	case StatusOk:
		_, err = w.Write([]byte("HTTP/1.1 200 OK\r\n"))
	case StatusBadRequest:
		_, err = w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
	case StatusInternalServerError:
		_, err = w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
	default:
		_, err = w.Write([]byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)))
	}

	return err
}

// it should set the following headers that we always want to include in our responses:
// Content-Length (Set to the given size)
// Connection (Set to close because we're not doing keep-alive's yet)
// Content-Type (Set to text/plain)
func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.Headers{}

	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"

	return h
}

// Implement a func WriteHeaders(w io.Writer, headers headers.Headers) error that does exactly what you'd expect.
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(
			fmt.Sprintf("%s: %s\r\n", key, value),
		))
		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
