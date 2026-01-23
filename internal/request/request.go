package request

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/mugiwara999/httpfromtcp/internal/headers"
)

type parseRequestState string

const (
	RequestStateInit parseRequestState = "init"
	HeadersState     parseRequestState = "headers"
	BodyState        parseRequestState = "body"
	RequestStateDone parseRequestState = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	Status      parseRequestState
}

var (
	ERROR_MALFORMED_REQUEST_LINE   = fmt.Errorf("malformed request line")
	ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported HTTP version")
	ERROR_INCOMPLETE_REQUEST       = fmt.Errorf("incomplete request")
)

const SEPARATOR = "\r\n"

func (r *Request) parse(data []byte) (int, error) {
	totalConsumed := 0

	switch r.Status {

	case RequestStateInit:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if rl == nil {
			return 0, nil // need more data
		}

		r.RequestLine = *rl
		r.Status = HeadersState
		totalConsumed += n

		i, err := r.parse(data[n:])

		return i + totalConsumed, err

	case HeadersState:
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}

		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return totalConsumed, err
		}

		totalConsumed += n

		if done {
			r.Status = RequestStateDone
			if cl, ok := r.Headers["content-length"]; ok && len(cl) > 0 {

				l, _ := strconv.Atoi(cl[0])

				if l > 0 {

					r.Status = BodyState
					// Continue parsing body if we have more data after headers
					remainingData := data[n:]
					if len(remainingData) > 0 {
						i, err := r.parse(remainingData)
						return i + totalConsumed, err
					}
				}

			}
		}

		return totalConsumed, nil

	case BodyState:

		n := r.Headers["content-length"]
		l, _ := strconv.Atoi(n[0])

		if len(r.Body) > l {
			return len(data), ERROR_INCOMPLETE_REQUEST
		}

		toConsume := min(len(data), l-len(r.Body))
		r.Body = append(r.Body, data[:toConsume]...)

		if len(r.Body) == l {
			r.Status = RequestStateDone
			return toConsume, nil
		}

		return toConsume, nil

	case RequestStateDone:
		return 0, nil
	}

	return 0, nil
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, []byte(SEPARATOR))
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
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
	r := &Request{
		Status:  RequestStateInit,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}

	buf := make([]byte, 4096)
	acc := []byte{}

	for r.Status != RequestStateDone {
		if len(acc) > 0 {
			consumed, parseErr := r.parse(acc)
			if parseErr != nil {
				return nil, parseErr
			}
			if consumed > 0 {
				acc = acc[consumed:]
			}
		}

		if r.Status == RequestStateDone {
			break
		}

		n, err := reader.Read(buf)
		if n > 0 {
			acc = append(acc, buf[:n]...)
		}

		if err == io.EOF {
			if r.Status != RequestStateDone {
				log.Println(r.Status)
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
