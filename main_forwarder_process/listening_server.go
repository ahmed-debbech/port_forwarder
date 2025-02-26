package main

import (
	"log"
	"net"
	"fmt"
	"bufio"
	"os"
	"bytes"
	"io/ioutil"
	"net/url"
	"net/http"
)

func StartListeningServer(ch chan string, secretPins []string){
	log.Println("listening on 3150")

	Run(ch, secretPins)
}

func Run(ch chan string, secretPins []string) {
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

		go handleRequest(conn, ch, secretPins)
	}
}

func ProcessHttpRequest(req *http.Request) http.Response{

	var t http.Response

	switch req.URL.Path {
	case "/home":
		dat, _ := os.ReadFile("templates/home.html")
		t = http.Response{
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Body:          ioutil.NopCloser(bytes.NewBufferString(string(dat))),
			ContentLength: int64(len(dat)),
			Request:       req,
			Header:        make(http.Header, 0),
		}
	case "/save":
		log.Println("Eee")
		
	case "/unlock":
		params, _ := url.ParseQuery(req.URL.RawQuery)

		if params.Get("code") != "" {
			if params.Get("code") == UnlockPass {
				dat, _ := os.ReadFile("templates/tool.html")
				t = http.Response{
					Status:        "200 OK",
					StatusCode:    200,
					Proto:         "HTTP/1.1",
					ProtoMajor:    1,
					ProtoMinor:    1,
					Body:          ioutil.NopCloser(bytes.NewBufferString(string(dat))),
					ContentLength: int64(len(dat)),
					Request:       req,
					Header:        make(http.Header, 0),
				}
			}else{
				j := "<div>WRONG CODE</div>"
				t = http.Response{
					Status:        "200 OK",
					StatusCode:    200,
					Proto:         "HTTP/1.1",
					ProtoMajor:    1,
					ProtoMinor:    1,
					Body:          ioutil.NopCloser(bytes.NewBufferString(j)),
					ContentLength: int64(len(j)),
					Request:       req,
					Header:        make(http.Header, 0),
				}
			}
		}
	default: 
		log.Println("404")
	}

	return t
}

func  handleRequest(conn net.Conn, ch chan string, secretPins []string) {
	reader := bufio.NewReader(conn)
	reader1 := bufio.NewReader(conn)

	defer conn.Close()

	req, err := http.ReadRequest(reader); 
	if err == nil {
		ress := ProcessHttpRequest(req)
		ress.Write(conn)
		return
	}

	buf := make([]byte, 6)
	
	_, err = reader1.Read(buf)
	if err != nil {
		return
	}
	for i:=0; i<=len(secretPins)-1; i++{
		
		if string(buf) == secretPins[i] {
			//get ipv4
			//send to channel
			remoteAddr := conn.RemoteAddr()

			// Type assert to TCPAddr to get the IP address
			tcpAddr, ok := remoteAddr.(*net.TCPAddr)
			if !ok {
				return
			}
			ch <- fmt.Sprintf("%s&%s", secretPins[i], tcpAddr.IP.String())
			break
		}
	}
}