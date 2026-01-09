package main

import (
	"fmt"
	"io"
	"log"
	"os"
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
	f, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal("error", err)
	}

	l := getLinesChannel(f)

	for s := range l {
		fmt.Println(s)
	}
}
