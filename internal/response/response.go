package response

import (
	"bytes"
	"fmt"
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

type WriterState string

const (
	WriteStateStatusLine WriterState = "statusLine"
	WriteStateHeaders    WriterState = "headers"
	WriteStateBody       WriterState = "body"
	WriteStateDone       WriterState = "done"
)

var ErrorResponeWrite = fmt.Errorf("invalid order of writing response")

type Writer struct {
	Buf   bytes.Buffer
	State WriterState
}

func NewWriter() *Writer {
	return &Writer{State: WriteStateStatusLine}
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.Buf.Write(p)
}

func (w *Writer) WriteStatusLine(code StatusCode) error {
	if w.State != WriteStateStatusLine {
		return ErrorResponeWrite
	}
	text := statusText[code]
	line := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, text)
	_, err := w.Buf.WriteString(line)
	w.State = WriteStateHeaders
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.State != WriteStateHeaders {
		return ErrorResponeWrite
	}
	for n, v := range h {
		for _, val := range v {
			fmt.Fprintf(&w.Buf, "%s: %s\r\n", n, val)
		}
	}
	w.Buf.WriteString("\r\n")

	if n, ok := h.Get("content-length"); ok && len(n) > 0 && len(n[0]) > 0 {
		w.State = WriteStateBody
	} else {
		w.State = WriteStateDone
	}
	return nil
}

func (w *Writer) WriteBody(b []byte) error {
	if w.State != WriteStateBody {
		return ErrorResponeWrite
	}
	w.Buf.Write(b)
	w.State = WriteStateDone
	return nil
}

func GetDefaultHeader(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h["content-length"] = []string{strconv.Itoa(contentLen)}
	h["connection"] = []string{"close"}
	h["content-type"] = []string{"text/plain"}

	return h
}

func (w *Writer) Read(p []byte) (int, error) {
	return w.Buf.Read(p)
}

func (w *Writer) WriteTrailer(h headers.Headers) error {
	for n, v := range h {
		for _, val := range v {
			fmt.Fprintf(&w.Buf, "%s: %s\r\n", n, val)
		}
	}
	w.Buf.WriteString("\r\n")
	return nil
}
