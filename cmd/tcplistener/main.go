package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const (
	ip   = "127.0.0.1"
	port = "42069"
)

func main() {
	listener, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		log.Fatalf("failed to Listen to tcp connection on IP: %v and Port: %v\n error: %v", ip, port, err)
	}
	defer listener.Close()
	for {
		log.Println("starting tcp listener...")
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to Accept message: %v", err)
			continue
		}
		log.Println("a connection has been accepted")
		message := getLinesChannel(conn)
		for msg := range message {
			fmt.Printf("%v\n", msg)
		}
		log.Println("the chanel has been closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	message := make(chan string)

	go func() {
		defer f.Close()
		defer close(message)
		var strBuff string
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if strBuff != "" {
						message <- strBuff
					}
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
