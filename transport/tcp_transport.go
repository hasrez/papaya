package transport

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"os"
)

const (
	ChunkSize        = 2048
	ChunkSizeForSend = ChunkSize - 1
	UuidLen          = 16
)

type TcpPeer struct {
	conn   net.Conn
	finish chan bool
}

type TcpTransport struct {
	listenAddr string
	listener   net.Listener
}

func NewTcpTransport(listenAddr string) *TcpTransport {
	return &TcpTransport{
		listenAddr: listenAddr,
	}
}

func (t *TcpTransport) Start() error {
	listener, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	t.listener = listener
	fmt.Println("TCP Listening on " + t.listenAddr)

	go t.acceptLoop()

	return nil
}

func (t *TcpTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("accept error from %v\n", conn)
			continue
		}
		peer := &TcpPeer{conn: conn, finish: make(chan bool)}
		go t.readLoop(peer)
		fmt.Printf("Accepted new connection from %v\n", conn)
		go func() {
			<-peer.finish
			err = peer.conn.Close()
			if err != nil {
				fmt.Printf("close connection error: %v\n", err)
			}
			fmt.Printf("Closing connection from %v successfully\n", peer)
		}()
	}
}

func (t *TcpTransport) readLoop(peer *TcpPeer) {
	for {
		buf := make([]byte, ChunkSize)
		n, err := peer.conn.Read(buf)
		if err != nil {
			fmt.Printf("read error from %v\n", peer)
			continue
		}

		id, err := uuid.FromBytes(buf[:UuidLen])
		if err != nil {
			fmt.Printf("uuid error from %v\n", peer)
		}
		content := buf[UuidLen:n]

		writeOnFile(id.String()+".gif", content)

		if n < ChunkSize {
			peer.finish <- true
			break
		}
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func writeOnFile(filename string, content []byte) error {
	if !fileExists(filename) {
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return err
		}
		file.Close()
	}
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}
	return nil
}
