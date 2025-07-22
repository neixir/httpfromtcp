package response

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/neixir/httpfromtcp/internal/headers"
)

// Create a new StatusCode "enum" type (you know, the fake Go enums). Define constants for the only 3 codes we're going to worry about:
type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type WriterStatus int

const (
	writerStateReadyForStatus WriterStatus = iota
	writerStateReadyForHeaders
	writerStateReadyForBody
)

type Writer struct {
	conn         net.Conn
	writerStatus WriterStatus
	isChunked    bool
}

// In the response package
func NewWriter(conn net.Conn) *Writer {
	return &Writer{
		conn:         conn,
		writerStatus: writerStateReadyForStatus,
	}
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

// CH7 L7
// It should map the given status code to the correct reason phrase, if it's one of the 3 that we support:
// 200 should return HTTP/1.1 200 OK
// 400 should return HTTP/1.1 400 Bad Request
// 500 should return HTTP/1.1 500 Internal Server Error
// Any other code should just leave the reason phrase blank.
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var err error = nil

	if w.writerStatus == writerStateReadyForStatus {
		switch statusCode {
		case StatusOk:
			_, err = w.conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
		case StatusBadRequest:
			_, err = w.conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		case StatusInternalServerError:
			_, err = w.conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		default:
			_, err = w.conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)))
		}

		w.writerStatus = writerStateReadyForHeaders

	} else {
		return fmt.Errorf("response status line already sent")
	}

	return err

}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerStatus == writerStateReadyForHeaders {
		for key, value := range headers {

			_, err := w.conn.Write([]byte(
				fmt.Sprintf("%s: %s\r\n", key, value),
			))
			if err != nil {
				return err
			}

			if strings.ToLower(key) == "transfer-encoding" &&
				strings.ToLower(value) == "chunked" {
				w.isChunked = true
			}

		}

		_, err := w.conn.Write([]byte("\r\n"))
		if err != nil {
			return err
		}
		w.writerStatus = writerStateReadyForBody

	} else {
		return fmt.Errorf("response headers already sent")
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerStatus == writerStateReadyForBody {
		w.conn.Write(p)

		if !w.isChunked {
			w.writerStatus = writerStateReadyForStatus
		}

	} else {
		return 0, fmt.Errorf("response body already sent")
	}

	return len(p), nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	body := fmt.Sprintf("%X\r\n%v\r\n", len(p), string(p))

	n, err := w.WriteBody([]byte(body))
	if err != nil {
		return n, err
	}

	return len(body), nil

}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	body := "0\r\n\r\n"

	n, err := w.WriteBody([]byte(body))
	if err != nil {
		return n, err
	}

	w.isChunked = false

	return len(body), nil
}
