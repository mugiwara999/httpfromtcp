package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/mugiwara999/httpfromtcp/internal/headers"
)

type StatusCode uint

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var statusText = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, code StatusCode) error {
	text, ok := statusText[code]
	if !ok {
		text = "Unknown"
	}

	line := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, text)

	_, err := w.Write([]byte(line))
	return err
}

func GetDefaultHeader(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h["content-length"] = []string{strconv.Itoa(contentLen)}
	h["connection"] = []string{"close"}
	h["content-type"] = []string{"text/plain"}

	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for n, v := range h {
		if len(v) > 1 {
			b := []byte(fmt.Sprintf("%s: %s", n, v[0]))
			for i := 1; i < len(v); i++ {
				b = append(b, fmt.Appendf(nil, ", %s", v[i])...)
			}
			_, err := w.Write(b)
		} else {
			mess := fmt.Sprintf("%s: %s", n, v[0])
			_, err := w.Write([]byte(mess))
		}
		w.Write([]byte("\r\n"))
	}

	_, err := w.Write([]byte("\r\n"))

	return err
}
