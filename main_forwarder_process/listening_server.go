package main

import (
	"log"
	"net"
	"fmt"
	"bufio"
	"net/http"
)

func StartListeningServer(ch chan string, secretPin string){
	log.Println("listening on 3150")

	Run(ch, secretPin)
}

func Run(ch chan string, secretPin string) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", "0.0.0.0", "3150"))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleRequest(conn, ch, secretPin)
	}
}

func ProcessHttpRequest(req *http.Request){

	switch req.URL.Path {
	case "/adjust":
		log.Println("Eee")
	case "/save":
		log.Println("Eee")
	case "/unlock":
		log.Println("Eee")
	}
}

func  handleRequest(conn net.Conn, ch chan string, secretPin string) {
	reader := bufio.NewReader(conn)
	defer conn.Close()

	req, err := http.ReadRequest(reader); 
	if err == nil {
		ProcessHttpRequest(req)
		return
	}

	buf := make([]byte, len(secretPin))
	
	_, err = reader.Read(buf)
	if err != nil {
		return
	}
	if string(buf) == secretPin {
		//get ipv4
		//send to channel
		remoteAddr := conn.RemoteAddr()

		// Type assert to TCPAddr to get the IP address
		tcpAddr, ok := remoteAddr.(*net.TCPAddr)
		if !ok {
			return
		}
		ch <- tcpAddr.IP.String()
	}
}