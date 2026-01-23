package main

import (
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
	handler := func(w *response.Writer, req *request.Request) *server.HandlerError {
		body := "<html><head>    <title>200 OK</title>  </head>  <body>    <h1>Success!</h1>    <p>Your request was an absolute banger.</p>  </body></html>"

		status := response.StatusOK
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = "<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>"
		case "/myproblem":
			status = response.StatusInternalServerError
			body = "<html><head><title>500 Internal Server Error</title>  </head>  <body>    <h1>Internal Server Error</h1>    <p>Okay, you know what? This one is on me.</p>  </body></html>"
		}

		w.WriteStatusLine(status)
		h := response.GetDefaultHeader(len(body))
		h.Set("content-type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody([]byte(body))

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
