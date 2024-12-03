package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hasrez/papaya/transport"
	"io"
	"net"
	"os"
	"time"
)

var addr = ":3000"

func main() {
	tr := transport.NewTcpTransport(addr)
	_ = tr.Start()
	time.Sleep(time.Second)
	go sendStream("bin1.bin")
	go sendStream("bin2.bin")
	select {}
}

func sendStream(name string) {
	binaryFile, err := os.Open(name)
	if err != nil {
		fmt.Println("Error opening binary file:", err)
		return
	}
	defer binaryFile.Close()

	uuid := uuid.New()

	buffer := make([]byte, transport.ChunkSizeForSend-transport.UuidLen)

	for {

		n, err := binaryFile.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading binary file:", err)
			return
		}

		if n == 0 {
			break
		}

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			panic(err)
		}

		_, err = conn.Write(append(uuid[:], buffer[:n]...))
		if err != nil {
			fmt.Println("Error writing binary file:", err)
		}
		time.Sleep(time.Second)
	}
}
