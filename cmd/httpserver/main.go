package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mugiwara999/httpfromtcp/internal/request"
	"github.com/mugiwara999/httpfromtcp/internal/response"
	"github.com/mugiwara999/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w io.Writer, req *request.Request) *server.HandlerError {
		body := "Hello from POST!\n"

		status := response.StatusOK
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = ""
		case "/myproblem":
			status = response.StatusInternalServerError
			body = ""
		}

		response.WriteStatusLine(w, status)

		headers := response.GetDefaultHeader(len(body))
		response.WriteHeaders(w, headers)

		w.Write([]byte(body))
		return nil
	}
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	log.Println("Server started on port", port)
	defer server.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
