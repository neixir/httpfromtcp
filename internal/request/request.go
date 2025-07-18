package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/neixir/httpfromtcp/internal/headers"
)

// type ParserState int
const bufferSize = 8

const (
	requestStateInitialized int = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	// Instead of reading all the bytes, and then parsing the request line,
	// it should use a loop to continually read from the reader
	// and parse new chunks using the parse method.

	// The loop should continue until the parser is in the "done" state.
	buf := make([]byte, bufferSize)

	// This will keep track of how much data we've read from the io.Reader into the buffer.
	readToIndex := 0

	// Create a new Request struct and set the state to "initialized".
	r := Request{
		State:   requestStateInitialized,
		Headers: headers.Headers{},
	}

	// While the state of the parser is not "done":
	for r.State != requestStateDone {

		// If the buffer is full (we've read data into the entire buffer), grow it.
		// Create a new slice that's twice the size and copy the old data into the new slice.
		if len(buf) == readToIndex {
			newbuf := make([]byte, len(buf)*2)
			copy(newbuf, buf)
			buf = newbuf
		}

		// Read from the io.Reader into the buffer starting at readToIndex.
		n, err := reader.Read(buf[readToIndex:])

		// If you hit the end of the reader (io.EOF) set the state to "done" and break out of the loop.
		// No ha resultat ser tan facil...
		if err == io.EOF {
			// First let the parser process anything left in the buffer
			if readToIndex > 0 {
				_, err = r.parse(buf[:readToIndex])
				if err != nil {
					return nil, err
				}
			}

			// FINAL call, with truly empty data and truly at end
			_, err = r.parse([]byte{})
			if err != nil {
				return nil, err
			}

			// Only now do we check for an incomplete body!
			contentLength := r.Headers.Get("Content-Length")
			if contentLength != "" {
				length, _ := strconv.Atoi(contentLength)
				if len(r.Body) < length {
					return nil, fmt.Errorf("actual length is smaller than Content-Length (%d < %d)\nBody: [%s]",
						len(r.Body), length, string(r.Body))
				}
			}

			r.State = requestStateDone
			break
		}

		// Update readToIndex with the number of bytes you actually read
		readToIndex += n

		// Call r.parse passing the slice of the buffer that has data that you've actually read so far
		parsedBytes, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// Remove the data that was parsed successfully from the buffer
		// (this keeps our buffer small and memory efficient).
		copy(buf, buf[parsedBytes:])

		// Decrement the readToIndex by the number of bytes that were parsed
		// so that it matches the new length of the buffer.
		readToIndex -= parsedBytes
	}

	return &r, nil

}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	// If it can't find an \r\n (this is important!) it should return 0 and no error.
	// This just means that it needs more data before it can parse the request line.
	if !strings.Contains(string(data), "\r\n") {
		return RequestLine{}, 0, nil
	}

	// Agafem fins el primer CRLF que trobem
	i := strings.Index(string(data), "\r\n")
	line := string(data[:i])

	rl := RequestLine{}
	numBytes := len(line)
	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		return rl, numBytes, fmt.Errorf("malformed request line")
	}

	// Verify that the "method" part only contains capital alphabetic characters.
	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return rl, numBytes, fmt.Errorf("method is not uppercase")
		}
	}

	partsHttp := strings.Split(parts[2], "/")
	httpVersion := partsHttp[1]

	// Verify that the http version part is 1.1, extracted from the literal HTTP/1.1 format, as we only support HTTP/1.1 for now.
	// TODO Tambe podriem comprovar que partsHttp[0] sigui HTTP
	if httpVersion != "1.1" {
		return rl, numBytes, fmt.Errorf("unsupported http version")
	}

	rl.Method = method
	rl.HttpVersion = httpVersion
	rl.RequestTarget = parts[1]

	return rl, numBytes, nil

}

func (r *Request) parse(data []byte) (int, error) {
	// It accepts the next slice of bytes that needs to be parsed into the Request struct
	// It updates the "state" of the parser, and the parsed RequestLine field.
	// It returns the number of bytes it consumed (meaning successfully parsed) and an error if it encountered one.
	totalBytesParsed := 0
	for r.State != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])

		if err != nil {
			return n, err
		}

		if n == 0 {
			return totalBytesParsed, nil
		}

		totalBytesParsed += n
	}

	return totalBytesParsed, nil

}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case requestStateInitialized:
		// If the state of the parser is "initialized", it should call qquestLine.
		rl, n, err := parseRequestLine(data)

		// If there is an error, it should just return the error.
		if err != nil {
			return n, err
		}

		// If zero bytes are parsed, but no error is returned, it should return 0 and nil: it needs more data.
		if n == 0 {
			return 0, nil
		}

		// If bytes are consumed successfully, it should update the .RequestLine field
		r.RequestLine = rl
		r.State = requestStateParsingHeaders

		return n + 2, nil // +2 per CRLF

	case requestStateParsingHeaders:
		totalParsed := 0
		for {
			n, done, err := r.Headers.Parse(data[totalParsed:])
			totalParsed += n

			if err != nil {
				return totalParsed, err
			}

			if done {
				r.State = requestStateParsingBody
				return totalParsed, nil
			}

			if n == 0 {
				return totalParsed, nil
			}
		}

	case requestStateParsingBody:
		// If there isn't a Content-Length header, move to the done state, nothing to parse
		contentLength := r.Headers.Get("Content-Length")
		if contentLength == "" {
			r.State = requestStateDone
			return 0, nil
		}

		length, err := strconv.Atoi(contentLength)
		if err != nil {
			// r.State = requestStateDone
			return 0, fmt.Errorf("invalid Content-Length")
		}

		// Figure out how many bytes you still need
		remaining := length - len(r.Body)

		// If there's more data than you need, only take as much as you need to hit Content-Length.
		toCopy := data
		if len(data) > remaining {
			toCopy = data[:remaining]
		}

		// Append the right number of bytes from data onto r.Body.
		r.Body = append(r.Body, toCopy...)

		// If you grabbed extra bytes from data, remember to return the right number (so the rest can be processed next).

		// After appending, if len(r.Body) is equal to length, you’re done!
		if len(r.Body) == length {
			r.State = requestStateDone
			return len(toCopy), nil
		}

		// If len(r.Body) is more than length, that’s an error.
		if len(r.Body) > length {
			return len(toCopy), fmt.Errorf("actual length is greater than Content-Length (%d > %d)\nBody: [%s]",
				len(r.Body), length, string(r.Body))
		}

		// If less, you need to wait for more data.
		return len(toCopy), nil

	case requestStateDone:
		// If the state of the parser is "done", it should return an error that says something like "error: trying to read data in a done state"
		return 0, fmt.Errorf("trying to read data in a done state")

	default:
		// If the state is anything else, it should return an error that says something like "error: unknown state"
		return 0, fmt.Errorf("unknown state")
	}

	return 0, nil
}
