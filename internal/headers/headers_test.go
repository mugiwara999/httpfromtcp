package headers

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            []byte
	numBytesPerRead int
	pos             int
}

func newChunkReader(data []byte, chunkSize int) *chunkReader {
	return &chunkReader{
		data:            data,
		numBytesPerRead: chunkSize,
		pos:             0,
	}
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}

	toRead := min(cr.numBytesPerRead, len(p), len(cr.data)-cr.pos)
	n = copy(p, cr.data[cr.pos:cr.pos+toRead])
	cr.pos += n

	return n, nil
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func parseHeadersFromChunks(data []byte, chunkSize int) (Headers, int, bool, error) {
	reader := newChunkReader(data, chunkSize)
	headers := NewHeaders()

	buf := make([]byte, 4096)
	accumulated := 0
	totalRead := 0

	for {
		n, readErr := reader.Read(buf[accumulated:])
		if readErr != nil && readErr != io.EOF {
			return nil, totalRead, false, readErr
		}

		accumulated += n

		parsed, done, parseErr := headers.Parse(buf[:accumulated])
		totalRead += parsed

		if parseErr != nil {
			return headers, totalRead, false, parseErr
		}

		if done {
			return headers, totalRead, true, nil
		}

		if parsed > 0 {
			copy(buf, buf[parsed:accumulated])
			accumulated -= parsed
		}

		if readErr == io.EOF {
			return headers, totalRead, false, nil
		}
	}
}

func TestHeadersParse_Basic(t *testing.T) {
	data := []byte("Host: localhost\r\n\r\n")

	headers, n, done, err := parseHeadersFromChunks(data, 3)
	require.NoError(t, err)
	require.True(t, done)

	assert.Equal(t, []string{"localhost"}, headers["host"])
	assert.Equal(t, len(data), n)
}

func TestHeadersParse_MultipleHeaders(t *testing.T) {
	data := []byte(
		"Host: api.example.com\r\n" +
			"Content-Type: application/json\r\n" +
			"User-Agent: curl/8.0\r\n\r\n",
	)

	headers, n, done, err := parseHeadersFromChunks(data, 5)
	require.NoError(t, err)
	require.True(t, done)

	assert.Equal(t, []string{"api.example.com"}, headers["host"])
	assert.Equal(t, []string{"application/json"}, headers["content-type"])
	assert.Equal(t, []string{"curl/8.0"}, headers["user-agent"])
	assert.Equal(t, len(data), n)
}

func TestHeadersParse_RepeatedHeader(t *testing.T) {
	data := []byte(
		"Set-Cookie: a=1\r\n" +
			"Set-Cookie: b=2\r\n\r\n",
	)

	headers, n, done, err := parseHeadersFromChunks(data, 4)
	require.NoError(t, err)
	require.True(t, done)

	assert.Equal(t, []string{"a=1", "b=2"}, headers["set-cookie"])
	assert.Equal(t, len(data), n)
}

func TestHeadersParse_Invalid_NoColon(t *testing.T) {
	data := []byte("Host localhost\r\n\r\n")

	headers, n, done, err := parseHeadersFromChunks(data, 10)
	require.Error(t, err)
	assert.Equal(t, ErrorNoFieldName, err)
	assert.False(t, done)
	assert.Equal(t, 0, n)
	assert.Empty(t, headers)
}

func TestHeadersParse_Invalid_EmptyFieldName(t *testing.T) {
	data := []byte(": value\r\n\r\n")

	headers, n, done, err := parseHeadersFromChunks(data, 10)
	require.Error(t, err)
	assert.Equal(t, ErrorNoFieldName, err)
	assert.False(t, done)
	assert.Equal(t, 0, n)
	assert.Empty(t, headers)
}

func TestHeadersParse_Incomplete(t *testing.T) {
	data := []byte("Host: localhost\r\nUser-Agent: test")

	headers, n, done, err := parseHeadersFromChunks(data, 8)
	require.NoError(t, err)
	assert.False(t, done)

	assert.Equal(t, []string{"localhost"}, headers["host"])
	assert.Equal(t, 17, n) // "Host: localhost\r\n"
}

func TestHeadersParse_ChunkBoundary(t *testing.T) {
	data := []byte("Authorization: Bearer token\r\n\r\n")

	headers, n, done, err := parseHeadersFromChunks(data, 2)
	require.NoError(t, err)
	require.True(t, done)

	assert.Equal(t, []string{"Bearer token"}, headers["authorization"])
	assert.Equal(t, len(data), n)
}

func TestHeadersParse_EmptyHeaders(t *testing.T) {
	data := []byte("\r\n")

	headers, n, done, err := parseHeadersFromChunks(data, 1)
	require.NoError(t, err)
	require.True(t, done)

	assert.Equal(t, 0, len(headers))
	assert.Equal(t, len(data), n)
}
