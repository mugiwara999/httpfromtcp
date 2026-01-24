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
	WriteStateChunkedBody WriterState = "chunkedBody"
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

	if te, ok := h.Get("transfer-encoding"); ok && len(te) > 0 {
		for _, val := range te {
			if val == "chunked" {
				w.State = WriteStateChunkedBody
				return nil
			}
		}
	}

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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.State != WriteStateChunkedBody {
		return 0, ErrorResponeWrite
	}
	
	size := len(p)
	if size == 0 {
		return 0, nil
	}
	
	fmt.Fprintf(&w.Buf, "%x\r\n", size)
	n, err := w.Buf.Write(p)
	if err != nil {
		return 0, err
	}
	_, err = w.Buf.WriteString("\r\n")
	if err != nil {
		return 0, err
	}
	
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.State != WriteStateChunkedBody {
		return 0, ErrorResponeWrite
	}
	
	n, err := w.Buf.WriteString("0\r\n\r\n")
	w.State = WriteStateDone
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.State != WriteStateDone {
		return ErrorResponeWrite
	}
	
	for n, v := range h {
		for _, val := range v {
			fmt.Fprintf(&w.Buf, "%s: %s\r\n", n, val)
		}
	}
	w.Buf.WriteString("\r\n")
	return nil
}
