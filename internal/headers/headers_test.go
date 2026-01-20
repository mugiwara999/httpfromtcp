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
	assert.Equal(t, []string{"localhost:42069"}, headers["host"])
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Multiple headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nContent-Type: application/json\r\nUser-Agent: curl/8.17.0\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost:42069"}, headers["host"])
	assert.Equal(t, []string{"application/json"}, headers["content-type"])
	assert.Equal(t, []string{"curl/8.17.0"}, headers["user-agent"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Multiple values for same header
	headers = NewHeaders()
	data = []byte("Set-Cookie: sessionid=abc123\r\nSet-Cookie: userid=xyz789\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, []string{"sessionid=abc123", "userid=xyz789"}, headers["set-cookie"])
	assert.True(t, done)

	// Test: Headers with extra whitespace in values (valid - trimming is OK)
	headers = NewHeaders()
	data = []byte("Host:    localhost:42069   \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost:42069"}, headers["host"])
	assert.True(t, done)

	// Test: Invalid - no colon
	headers = NewHeaders()
	data = []byte("Host localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, ErrorNoFieldName, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid - empty field name
	headers = NewHeaders()
	data = []byte(": value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, ErrorNoFieldName, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid - whitespace-only field name
	headers = NewHeaders()
	data = []byte("   : value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, ErrorNoFieldName, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Incomplete headers (no empty line yet)
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nContent-Type: application/json\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost:42069"}, headers["host"])
	assert.Equal(t, []string{"application/json"}, headers["content-type"])
	assert.False(t, done)

	// Test: Empty data
	headers = NewHeaders()
	data = []byte("")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Just empty line (immediate end of headers)
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)
	assert.Equal(t, 0, len(headers))

	// Test: Header with empty value
	headers = NewHeaders()
	data = []byte("X-Empty:\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, []string{""}, headers["x-empty"])
	assert.True(t, done)

	// Test: Case insensitivity
	headers = NewHeaders()
	data = []byte("Content-Type: text/html\r\ncontent-type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, []string{"text/html", "application/json"}, headers["content-type"])
	assert.True(t, done)
}
