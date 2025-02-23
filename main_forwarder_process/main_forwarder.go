package main

import (
	"log"
	"os"
	_"io"
	"fmt"
	"strings"
	"time"
	"os/exec"
)


var (
	path string
	secretPin string
	joiningHosts []string
)



func ReadFile(ch chan []string){

	for {
		dat, err := os.ReadFile("PORTS")
		if err != nil {
			panic("could not read PORTS files")
		}

		ls := strings.Split(string(dat), "\n")
		
		l := make([]string, 0)
		
		if len(ls) == 0 {ch <- ls; continue;}

		for i := 0; i<=len(ls)-1; i++{
			if strings.Contains(ls[i], ":") {
				m := strings.Split(ls[i], ":")
				l = append(l, m[0])
				l = append(l, m[1])
			}
		}

		ch <- l
		time.Sleep(time.Second * 5)
	}
}
type Ports struct {
	PortNumber string
	HostIp string
	Cmd *exec.Cmd
}

func LaunchForward(ch chan []string, activePorts []Ports){

	for {
		hosts := <- ch
		
		//activePorts := make([]string, 0)
		newPortsToOpen := make([]Ports, 0)

		for i:=0; i<=len(hosts)-1; i+=2 {
			newPort := hosts[i+1]
			for j:=0; j<=len(activePorts)-1; j++{
				if hosts[i+1] == activePorts[j].PortNumber  {
					newPort = ""
					break
				}
			}
			if newPort != "" {
				ps := Ports{
					PortNumber: newPort,
					HostIp: hosts[i],
					Cmd: nil,
				}
				activePorts = append(activePorts, ps)
				newPortsToOpen = append(newPortsToOpen, ps)
			}
		}

		/// launch activePorts

		for i:=0; i<=len(newPortsToOpen)-1; i++ {
			go runCommand(newPortsToOpen[i], activePorts)
		}

		//now delete ports not in the PORTS file
		for i:=0; i<=len(activePorts)-1; i++ {
			stillexist := false
			for j:=0; j<=len(hosts)-1; j+=2{
				if activePorts[i].PortNumber == hosts[j+1] {
					stillexist = true
					break
				}
			}
			if !stillexist {
				//delete Process
				activePorts[i].Cmd.Process.Kill() 
				log.Println(activePorts[i], "to delete")
				activePorts = append(activePorts[:i], activePorts[i+1:]... )
			}
		}


	}
}

func runCommand(command Ports, activePorts []Ports){
	log.Println(command.PortNumber, "new port to open")

	cmd := exec.Command(path, fmt.Sprintf("0.0.0.0:%s", command.PortNumber), fmt.Sprintf("%s:%s", command.HostIp, command.PortNumber))

	for k:=0; k<=len(activePorts)-1; k++{
		if activePorts[k].PortNumber == command.PortNumber {
			activePorts[k].Cmd = cmd
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//cmdOut, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil{
		log.Println("can't run command for port", command.PortNumber, "because:", err)
		return
	}
	//cmdBytes, _ := io.ReadAll(cmdOut)

	err = cmd.Wait()
	if err != nil {
		log.Println("a problem occured in the command for port", command.PortNumber, "because:", err)
		return
	}

	log.Println(fmt.Sprintf("port %s has been shutdown, bellow are its logs:", command.PortNumber))
	//log.Println(string(cmdBytes))	
}

func ListenForJoiningHosts(ch chan string){

	for {
		
		ip := <- ch
		
		found := false
		for i:=0; i<=len(joiningHosts)-1; i++ {
			if joiningHosts[i] == ip {
				found = true
				break
			}
		}
		if !found {
			joiningHosts = append(joiningHosts, ip)  
		}
		
		log.Println("current joined hosts", joiningHosts)
	}
}

func main(){
	log.Println("Broadcast v1")

	if len(os.Args)  < 3 {
		log.Println("Usage: main_forwarder path/to/single_forwarder secret_pin")	
		return
	}

	path = os.Args[1]
	secretPin = os.Args[2]

	ch := make(chan []string)
	ch1 := make(chan string)

	activePorts := make([]Ports, 0)

	go ReadFile(ch)

	go LaunchForward(ch, activePorts)

	go StartListeningServer(ch1, secretPin)
	
	go ListenForJoiningHosts(ch1)

	select{}
}