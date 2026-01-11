package request

import (
	"bytes"
	"fmt"
	"io"
)

type parseRequestState string

const (
	RequestStateInit parseRequestState = "init"
	RequestStateDone parseRequestState = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Status      parseRequestState
}

func (r *Request) parse(data []byte) (int, error) {
	if r.Status != RequestStateDone {

		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if rl == nil {
			return 0, nil
		}

		r.Status = RequestStateDone

		r.RequestLine = *rl

		return n, nil

	}
	return 0, nil
}

var (
	ERROR_MALFORMED_REQUEST_LINE   = fmt.Errorf("malformed request line")
	ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported HTTP version")
	ERROR_INCOMPLETE_REQUEST       = fmt.Errorf("incomplete request - no separator found")
)

const SEPARATOR = "\r\n"

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, []byte(SEPARATOR))

	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	// restMsg := b[idx+len(SEPARATOR):]

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpv := bytes.Split(parts[2], []byte("/"))

	if len(httpv) != 2 || !bytes.Equal(httpv[0], []byte("HTTP")) || !bytes.Equal(httpv[1], []byte("1.1")) {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VERSION
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpv[1]),
	}

	n := idx + len(SEPARATOR)

	return rl, n, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{Status: RequestStateInit}

	buf := make([]byte, 8)
	acc := []byte{}

	for r.Status != RequestStateDone {

		n, err := reader.Read(buf)

		if n > 0 {
			acc = append(acc, buf[:n]...)
		}

		i, parseErr := r.parse(acc)
		if parseErr != nil {
			return nil, parseErr
		}

		acc = acc[i:]

		if err == io.EOF {

			if r.Status != RequestStateDone {
				return nil, ERROR_INCOMPLETE_REQUEST
			}
			break
		}

		if err != nil {
			return nil, err
		}

	}

	return r, nil
}
