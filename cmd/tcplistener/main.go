package main

import (
	"fmt"
	"log"
	"net"

	"github.com/mugiwara999/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			r, err := request.RequestFromReader(c)
			if err != nil {
				log.Println("request parse error:", err)
				return
			}

			fmt.Println("Request line:")
			fmt.Println("- Method:", r.RequestLine.Method)
			fmt.Println("- Target:", r.RequestLine.RequestTarget)
			fmt.Println("- Version:", r.RequestLine.HttpVersion)

			fmt.Println("Headers:")

			r.Headers.ForEach(func(s string, v []string) {
				fmt.Printf("- %s: %v\n", s, v)
			})

			_, ok := r.Headers.Get("Content-Length")
			if ok {
				fmt.Println("Body:")
				fmt.Println(string(r.Body))
			}

			c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nOK"))
		}(conn)
	}
}
