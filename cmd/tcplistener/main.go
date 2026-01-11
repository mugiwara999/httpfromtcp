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
			log.Fatal("error", err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", err)
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", r.RequestLine.Method)
		fmt.Println("- Target:", r.RequestLine.RequestTarget)
		fmt.Println("- Version:", r.RequestLine.HttpVersion)

	}
}
