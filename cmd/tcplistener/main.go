package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer close(out)
		defer f.Close()

		b := make([]byte, 8)

		var line []byte

		for {
			n, err := f.Read(b)

			if n > 0 {
				for i := 0; i < n; i++ {
					ch := b[i]

					if ch == '\n' {
						if line != nil {
							out <- fmt.Sprintf("read %s\n", string(line))
						}
						line = nil
					} else {
						line = append(line, ch)
					}

				}
			}

			if err == io.EOF {
				out <- string(line)
				break
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	return out
}

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

		for line := range getLinesChannel(conn) {
			fmt.Printf("read %s", line)
		}
	}
}
