package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test: Valid single header
func TestHeadersParse(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")

	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

// Test: Valid single header with extra whitespace
func TestHeadersParseWithWhitespace(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host:    localhost:42069   \r\n\r\n")

	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 29, n)
	assert.False(t, done)
}

// Test: Invalid spacing header
func TestHeadersParseInvalidSpacing(t *testing.T) {
	headers := NewHeaders()
	data := []byte("       Host: localhost:42069\r\n\r\n")

	n, done, err := headers.Parse(data)

	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

// Test: Valid 2 headers with existing headers
func TestHeadersParseFirstOfTwoHeaders(t *testing.T) {
	headers := NewHeaders()
	headers["User-Agent"] = "curl/7.81.0"
	data := []byte("Host: localhost:42069\r\nAccept: */*\r\n\r\n")

	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, "curl/7.81.0", headers["User-Agent"])
	assert.Equal(t, "", headers["Accept"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
}

// Test: Valid done
func TestHeadersParseDone(t *testing.T) {
	headers := NewHeaders()
	data := []byte("\r\n")

	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}
