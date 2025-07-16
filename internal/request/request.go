package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	r := Request{}
	lines := strings.Split(string(request), "\r\n")
	r.RequestLine, err = parseRequestLine(lines[0])
	if err != nil {
		return nil, err
	}

	return &r, nil

}

func parseRequestLine(line string) (RequestLine, error) {
	rl := RequestLine{}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return rl, fmt.Errorf("malformed request line")
	}

	// Verify that the "method" part only contains capital alphabetic characters.
	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return rl, fmt.Errorf("method is not uppercase")
		}
	}

	partsHttp := strings.Split(parts[2], "/")
	httpVersion := partsHttp[1]

	// Verify that the http version part is 1.1, extracted from the literal HTTP/1.1 format, as we only support HTTP/1.1 for now.
	if httpVersion != "1.1" {
		return rl, fmt.Errorf("unsupported http version")
	}

	rl.Method = method
	rl.HttpVersion = httpVersion
	rl.RequestTarget = parts[1]

	return rl, nil

}
