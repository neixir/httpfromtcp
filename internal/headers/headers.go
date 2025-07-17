package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

// Mutate the Headers by adding newly parsed key-value pairs
// Return n (the number of bytes consumed), done (whether or not it has finished parsing headers), and err (if it encountered an error)
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Look for a CRLF, if it doesn't find one, assume you haven't been given enough data yet.
	// Consume no data, return false for done, and nil for err.
	if !strings.Contains(string(data), "\r\n") {
		return 0, false, nil
	}

	// If you do find a CRLF, but it's at the start of the data, you've found the end of the headers,
	// so return the proper values immediately.
	// Note: The Parse function should only return done=true when the data starts with a CRLF,
	// which can't happen when it finds a new key/value pair.
	parts := strings.Split(string(data), "\r\n")
	if parts[0] == "" {
		return 2, true, nil // n=2 per CRLF
	}

	// En aquest punt ja tenim una linia sencera
	fieldLine := parts[0]

	// Busquem el primer ":" per dividir
	i := strings.Index(fieldLine, ":")
	if i < 0 {
		return 0, false, fmt.Errorf("malformed header line")
	}

	key := fieldLine[:i]     // field-name
	value := fieldLine[i+1:] // field-value

	// L'ultim caracter de la primera part (la clau) no pot ser espai
	// ("ensure there are no spaces between the colon and the key")
	lastChar := key[len(key)-1:]
	if lastChar == " " {
		return 0, false, fmt.Errorf("malformed header line")
	}

	// Remove any extra whitespace from the key and value
	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.TrimSpace(value)

	// Return an error if the key contains an invalid character.
	// Valid: A-Z, a-z, 0-9 i "!, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~"
	validChars := "!#$%&'*+-.^_`|~"
	for _, c := range key {
		isUpper := c >= 'A' && c <= 'Z'
		isLower := c >= 'a' && c <= 'z'
		isDigit := c >= '0' && c <= '9'
		isSpecial := strings.Contains(validChars, string(c))

		if !isUpper && !isLower && !isDigit && !isSpecial {
			return 0, false, fmt.Errorf("invalid character in field name")
		}
	}

	// Assuming the format was valid (if it isn't return an error),
	// add the key/value pair to the Headers map
	// If a header key already exists in the map before inserting one,
	// append the new value to the existing value, separated by a comma.
	current, ok := h[key]
	if ok {
		h[key] = fmt.Sprintf("%s, %s", current, value)
	} else {
		h[key] = value
	}

	// Return the number of bytes consumed
	consumed := len(fieldLine) + 2 // +2 per CRLF

	// It's important to understand that this function will be called over and over
	// until all the headers are parsed, and it can only parse one key/value pair at a time.

	return consumed, false, nil
}

// Aquesta funcio s'utilitza als test pero no explica com ha de ser. A veure...
func NewHeaders() Headers {
	return make(Headers)
}
