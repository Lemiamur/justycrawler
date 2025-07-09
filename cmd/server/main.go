package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":1935")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 	Клиент -> Сервер:
	// c0: [version]
	// c1: [time][zero][random data]
	c0 := readBytes(conn, 1)
	if c0 == nil {
		return
	}

	c1 := readBytes(conn, 1536)
	if c1 == nil {
		return
	}

	// Сервер -> Клиент:
	// s0: [version]
	// s1: [time][zero][random data]
	s0 := []byte{0x01}
	s1 := make([]byte, 1536)

	err := writeBytes(conn, s0)
	if err != nil {
		return
	}

	err = writeBytes(conn, s1)
	if err != nil {
		return
	}

	// Клиент -> Сервер:
	// c2: [echo time from s1][time from c1][s1's random data]
	c2 := readBytes(conn, 1536)
	if c2 == nil {
		return
	}

	// Сервер -> Клиент:
	// s2: [echo time from с2][time from с2][с1's random data]
	s2 := make([]byte, 1536)

	err = writeBytes(conn, s2)
	if err != nil {
		return
	}
}

func readBytes(conn net.Conn, totalBytesToRead int) []byte {
	buffer := make([]byte, totalBytesToRead)

	totalBytesRead := 0

	for totalBytesRead < totalBytesToRead {
		bytesRead, err := conn.Read(buffer[totalBytesRead:])
		if err != nil {
			if err == io.EOF {
				fmt.Println("Соединение закрыто.")
				break
			}
			log.Println(err)
			return nil
		}

		totalBytesRead += bytesRead
	}

	return buffer
}

func writeBytes(conn net.Conn, data []byte) error {
	totalBytesWritten := 0

	for totalBytesWritten < len(data) {
		bytesWritten, err := conn.Write(data[totalBytesWritten:])
		if err != nil {
			return err
		}

		totalBytesWritten += bytesWritten
	}

	return nil
}
