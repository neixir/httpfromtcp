package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// "Valid single header with extra whitespace"
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// "Valid 2 headers with existing headers"
	// Test that when you call Parse multiple times, it adds new headers to the existing map without overwriting previous ones.
	headers = NewHeaders()
	data1 := []byte("Host: localhost:42069\r\n")
	headers.Parse(data1)
	data2 := []byte("Accept: */*\r\n\r\n")
	_, done, err = headers.Parse(data2)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "*/*", headers["accept"])
	assert.False(t, done)

	// "Valid done"
	headers = NewHeaders()
	data = []byte("\r\n\r\nHola bona tarda")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Capital letters in key
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\n")
	_, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.False(t, done)

	// Invalid character in header key
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n")
	_, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 2, n)
	assert.False(t, done)

}
