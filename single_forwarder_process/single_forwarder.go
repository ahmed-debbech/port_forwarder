package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

// Proxy listens on localAddr and forwards connections to remoteAddr.
func startProxy(localAddr, remoteAddr string) error {
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Proxy listening on %s, forwarding to %s\n", localAddr, remoteAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go handleConnection(clientConn, remoteAddr)
	}
}

// Handles the client connection by forwarding it to the remote address.
func handleConnection(clientConn net.Conn, remoteAddr string) {
	fmt.Println("dealing with client")
	defer clientConn.Close()

	serverConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		fmt.Printf("Failed to connect to remote server: %v\n", err)
		return
	}
	defer serverConn.Close()

	fmt.Printf("Connected: %s <-> %s\n", clientConn.RemoteAddr(), remoteAddr)

	// Copy data between client and server
	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run proxy.go <localAddr> <remoteAddr>")
		os.Exit(1)
	}

	localAddr := os.Args[1]
	remoteAddr := os.Args[2]

	err := startProxy(localAddr, remoteAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
