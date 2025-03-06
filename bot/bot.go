package main

import (
	"os"
	"log"
	"time"
	"net"
	"fmt"
	"strconv"
)

func main(){
	log.Println("Bot started!")

	if len(os.Args) < 4 {
		log.Println("Usage: go run bot.go yourserverip secretPin refreshRateInSec")
		return
	}

	refreshRate, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return
	}

	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:3150", os.Args[1]))
		if err != nil {
			log.Println("could not connect to server because:", err)
			time.Sleep(time.Second * time.Duration(refreshRate))
			continue
		}
	
		s := fmt.Sprintf("%s", os.Args[2])
		_, err = conn.Write([]byte(s))
		if err != nil {
			log.Println("could not send data to server because:", err)
			time.Sleep(time.Second * time.Duration(refreshRate))
			continue
		}

		log.Println("sent to server!")
		conn.Close()
		time.Sleep(time.Second * time.Duration(refreshRate))
	}
}