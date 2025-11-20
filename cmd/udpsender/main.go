package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	ip   = "127.0.0.1"
	port = "42069"
)

func main() {
	remoteUDPAdr, err := net.ResolveUDPAddr("udp", ip+":"+port)
	if err != nil {
		log.Fatalf("failed to resolve udp addres with ip: %v and port: %v\n error: %v", ip, port, err)
	}

	log.Println("Successfully Resolved UPD Address")

	/*
		localUDPAdr := net.UDPAddr{
			IP:   net.IP("127.0.0.1"),
			Port: 42070,
		}

	*/

	conn, err := net.DialUDP("udp", nil, remoteUDPAdr)
	if err != nil {
		log.Fatalf("failed to Dial: %v", err)
	}
	defer conn.Close()

	log.Println("Successfully Dialed")

	stdin := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		message, err := stdin.ReadString('\n')
		if err != nil {
			log.Printf("error while reading stdin: %v\n", err)
			if message == "" {
				continue
			}
		}

		_, err = conn.Write([]byte(message))
		if err != nil {
			log.Printf("error while writing to connection: %v", err)
		}
	}
}
