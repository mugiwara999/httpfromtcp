package server

import (
	"github.com/mugiwara999/httpfromtcp/internal/response"
)

func WriteHandlerError(w *response.Writer, herr *HandlerError) {
	if herr == nil {
		return
	}

	message := herr.Message + "\n"

	w.WriteStatusLine(herr.Status)
	w.WriteHeaders(response.GetDefaultHeader(len(message)))
	w.WriteBody([]byte(message))
}
