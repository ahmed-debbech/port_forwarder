package main

import (
	"os"
	"log"
	"time"
	"net"
	"fmt"
)

func main(){
	log.Println("Bot started!")

	if len(os.Args) < 3 {
		log.Println("Usage: go run bot.go yourserverip secretPin")
		return
	}

	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:3150", os.Args[1]))
		if err != nil {
			log.Println("could not connect to server because:", err)
			time.Sleep(time.Second * 10)
			continue
		}
	
		s := fmt.Sprintf("%s", os.Args[2])
		_, err = conn.Write([]byte(s))
		if err != nil {
			log.Println("could not send data to server because:", err)
			time.Sleep(time.Second * 10)
			continue
		}

		log.Println("sent to server!")
		conn.Close()
		time.Sleep(time.Second * 10)
	}
}