package main

import (
	"log"
	"net"
	"os"
)

const (
	networkType = "tcp"
	addressPort = ":9090"
	message     = "training:ok "
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	listener, err := net.Listen(networkType, addressPort)
	if err != nil {
		panic(err)
	}
	logger.Printf("Listening on %s", addressPort)
	defer listenerClose(listener, logger)

	for {
		conn, err := listener.Accept()
		if err != nil {
			_ = conn.Close()
			panic(err)
		}

		_, err = conn.Write([]byte(message))
		if err != nil {
			_ = conn.Close()
			panic(err)
		}

		logger.Printf("Message was sucsessfuly sent into conn %s", conn.RemoteAddr())
		_ = conn.Close()
	}
}

func listenerClose(listener net.Listener, logger *log.Logger) {
	err := listener.Close()
	if err != nil {
		logger.Printf("Error closing listener: %s", err)
	}

	logger.Printf("Listener closed")
}
