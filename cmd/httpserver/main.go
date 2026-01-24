package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mugiwara999/httpfromtcp/internal/headers"
	"github.com/mugiwara999/httpfromtcp/internal/request"
	"github.com/mugiwara999/httpfromtcp/internal/response"
	"github.com/mugiwara999/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	handler := func(w *response.Writer, req *request.Request) *server.HandlerError {
		body := "<html><head>    <title>200 OK</title>  </head>  <body>    <h1>Success!</h1>    <p>Your request was an absolute banger.</p>  </body></html>"

		status := response.StatusOK

		if req.RequestLine.RequestTarget == "/yourproblem" {
			status = response.StatusBadRequest
			body = "<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>"

		} else if req.RequestLine.RequestTarget == "/myproblem" {
			status = response.StatusInternalServerError
			body = "<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>"

		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {

			target := req.RequestLine.RequestTarget

			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])

			if err != nil {

				status = response.StatusInternalServerError
				body = "<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>"
				log.Println(err)
			} else {

				w.WriteStatusLine(response.StatusOK)
				h := response.GetDefaultHeader(len(body))
				h.Delete("content-length")
				h.Set("Transfer-encoding", "chunked")
				h.Set("content-type", "text/plain")
				h.Set("Trialer", "X-content-sha256")
				h.Set("Trialer", "X-Content-Length")
				w.WriteHeaders(h)

				fullBody := []byte{}

				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					defer res.Body.Close()
					if err != nil {
						break
					}

					fullBody = append(fullBody, data[:n]...)
					w.Write(fmt.Appendf(nil, "%x\r\n", n))
					w.Write(data[:n])
					w.Write([]byte("\r\n"))
				}
				w.Write([]byte("0\r\n"))

				hashval := sha256.Sum256(fullBody)

				trailer := headers.NewHeaders()
				trailer.Set("X-content-sha256", fmt.Sprintf("%x", hashval))
				trailer.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteTrailer(trailer)
				return nil
			}

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
