package server

import (
	"io"

	"github.com/mugiwara999/httpfromtcp/internal/response"
)

func WriteHandlerError(w io.Writer, herr *HandlerError) {
	if herr == nil {
		return
	}

	message := herr.Message + "\n"

	response.WriteStatusLine(w, herr.Status)
	response.WriteHeaders(w, response.GetDefaultHeader(len(message)))

	w.Write([]byte(message))
}
