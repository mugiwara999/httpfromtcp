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

		err := response.WriteStatusLine(w, response.StatusOK)
		if err != nil {
			log.Println("Failed to write status")
		}

		headers := response.GetDefaultHeader(len(body))
		err = response.WriteHeaders(w, headers)
		if err != nil {
			log.Println("Failed to write headers")
		}

		_, err = w.Write([]byte(body))
		if err != nil {
			log.Println("Failed to write body")
		}
		return nil
	}
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
