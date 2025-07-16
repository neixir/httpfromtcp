package request

import (
	"fmt"
	"io"
	"strings"
)

// type ParserState int
const bufferSize = 8

const (
	StateInitialized int = iota
	StateDone
)

type Request struct {
	RequestLine RequestLine
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
		State: StateInitialized,
	}

	// While the state of the parser is not "done":
	for {
		if r.State == StateDone {
			break
		}

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
		if err == io.EOF {
			r.State = StateDone
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

	lines := strings.Split(string(data), "\r\n")
	line := lines[0]

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

	switch r.State {
	case StateInitialized:
		// If the state of the parser is "initialized", it should call parseRequestLine.
		rl, n, err := parseRequestLine(data)

		// If there is an error, it should just return the error.
		if err != nil {
			return n, err
		}

		// If zero bytes are parsed, but no error is returned, it should return 0 and nil: it needs more data.
		if n == 0 {
			return 0, nil
		}

		// If bytes are consumed successfully, it should update the .RequestLine field and change the state to "done".
		r.RequestLine = rl
		r.State = StateDone

		return n, nil

	case StateDone:
		// If the state of the parser is "done", it should return an error that says something like "error: trying to read data in a done state"
		return 0, fmt.Errorf("trying to read data in a done state")

	default:
		// If the state is anything else, it should return an error that says something like "error: unknown state"
		return 0, fmt.Errorf("unknown state")
	}
}
