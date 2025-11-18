package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("Failed to Open file: %v", err)
	}
	defer file.Close()

	message := getLinesChannel(file)

	for msg := range message {
		fmt.Printf("read: %v\n", msg)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	message := make(chan string)

	go func() {
		var strBuff string
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if strBuff != "" {
						message <- strBuff
					}
					f.Close()
					close(message)
					break
				}
				log.Fatalf("Failed to Read file: %v", err)
			}

			str := string(buffer[:n])
			parts := strings.Split(str, "\n")

			if len(parts) == 1 {
				strBuff += parts[0]
				continue
			}

			output := strBuff
			strBuff = parts[len(parts)-1]

			for i := 0; i < len(parts)-1; i++ {
				output += parts[i]
			}

			message <- output
		}
	}()

	return message
}
