package main

import (
	"log"
	"os"
	_"io"
	_"fmt"
	"strings"
	"time"
	_"os/exec"
	"sync"
)

const (

	StartForwardCmd = "sudo bash -c 'echo 1 > /proc/sys/net/ipv4/ip_forward'"

	AddPreroutingCmd = "sudo iptables -t nat -A PREROUTING -p &&prot --dport &&port_dest -j DNAT --to-destination &&ip_dest:&&port_dest"
	AddPostroutingCmd = "sudo iptables -t nat -A POSTROUTING -p &&prot -d &&ip_dest -j SNAT --to-source &&ip_this"
	DelPreroutingCmd = "sudo iptables -t nat -D PREROUTING -p &&prot --dport &&port_dest -j DNAT --to-destination &&ip_dest:&&port_dest"
	DelPostroutingCmd = "sudo iptables -t nat -D POSTROUTING -p &&prot -d &&ip_dest -j SNAT --to-source &&ip_this"
)

var (
	LocalIp string
	secretPins []string
	JoiningHosts map[string]string
	UnlockPass string
	PortsFileContent []string
    MuPortFile       sync.Mutex

)

type Ports struct {
	PortNumber string
	HostName string
	HostIp string
	CmdStart []string
	CmdStop []string
}

func BindCmd(command string, ip_dest string, port_dest string, ip_this string, prot string) string{
	command = strings.ReplaceAll(command, "&&ip_dest", ip_dest)
	command = strings.ReplaceAll(command, "&&port_dest", port_dest)
	command = strings.ReplaceAll(command, "&&ip_this", ip_this)
	command = strings.ReplaceAll(command, "&&prot", prot)
	return command
}

func ReadFile(ch chan []string){

	for {
		MuPortFile.Lock()
		
		dat, err := os.ReadFile("PORTS")
		if err != nil {
			panic("could not read PORTS files")
		}
		MuPortFile.Unlock()
		
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

func LaunchForward(ch chan []string, activePorts []Ports){

	for {
		hosts := <- ch
		PortsFileContent = hosts
		
		//activePorts := make([]string, 0)
		newPortsToOpen := make([]Ports, 0)

		for i:=0; i<=len(hosts)-1; i+=2 {

			newPort := hosts[i+1]
			newHost := hosts[i]

			if _, ok := JoiningHosts[newHost]; !ok {
				newHost = ""
				continue
			}

			for j:=0; j<=len(activePorts)-1; j++{

				if (hosts[i+1] == activePorts[j].PortNumber &&
				JoiningHosts[hosts[i]] == activePorts[j].HostIp) {
					newPort = ""
					break
				}
			}
			if newPort != "" {
				cs := []string{ 
					BindCmd(AddPreroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp ,"tcp"),
					BindCmd(AddPostroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp ,"tcp"),
				}
				ce := []string{ 
					BindCmd(DelPreroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp ,"tcp"),
					BindCmd(DelPostroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp ,"tcp"),
				}
				ps := Ports{
					PortNumber: newPort,
					HostName: hosts[i],
					HostIp: JoiningHosts[hosts[i]],
					CmdStart: cs, 
					CmdStop: ce,
				}
				activePorts = append(activePorts, ps)
				newPortsToOpen = append(newPortsToOpen, ps)
			}
		}


		//now delete ports not in the PORTS file or that has changed their ips
		for i:=0; i<=len(activePorts)-1; i++ {
			stillexist := false
			for j:=0; j<=len(hosts)-1; j+=2{
				if activePorts[i].PortNumber == hosts[j+1] &&
				activePorts[i].HostIp == JoiningHosts[hosts[j]]{
					stillexist = true
					break
				}
			}
			if !stillexist {
				//delete Port
				log.Println(activePorts[i].CmdStop[0])
				log.Println(activePorts[i].CmdStop[1])
				log.Println(activePorts[i].PortNumber, activePorts[i].HostIp, "to delete")
				activePorts = append(activePorts[:i], activePorts[i+1:]... )
			}
		}

		/// launch activePorts
		for i:=0; i<=len(newPortsToOpen)-1; i++ {
			//go runCommand(newPortsToOpen[i], activePorts)
			log.Println("adding new")
			log.Println(newPortsToOpen[i].CmdStart[0])
			log.Println(newPortsToOpen[i].CmdStart[1])
		}
	}
}

/*func runCommand(command Ports, activePorts []Ports){

	ip, ok := JoiningHosts[command.HostName];
	if !ok {
		log.Println(command.HostName, "did not show up online yet! ... will keep waiting for it")
		found := false
		for !found {
			ip, found = JoiningHosts[command.HostName];
			time.Sleep(time.Second * 5)
			log.Println("wait", command.HostName)
		}
	}

	log.Println(command.PortNumber, "new port to open for ip", ip)

	cmd := exec.Command(path, fmt.Sprintf("0.0.0.0:%s", command.PortNumber), fmt.Sprintf("%s:%s", ip, command.PortNumber))

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
*/
func ListenForJoiningHosts(ch chan string){

	for {
		
		pin_ip := <- ch

		dd := strings.Split(pin_ip, "&")

		JoiningHosts[dd[0]] = dd[1]

		log.Println("current joined hosts", JoiningHosts)
	}
}

func DividSecretPins(pins string) []string{
	return strings.Split(pins, "-")
}

func main(){
	log.Println("Broadcast v1")

	if len(os.Args)  < 4 {
		log.Println("Usage: main_forwarder localip secret_pin*6chars-secret_pin2-secret_pin3... unlockPass")	
		return
	}

	JoiningHosts = make(map[string]string, 0)
	LocalIp = os.Args[1]
	secretPins = DividSecretPins(os.Args[2])
	UnlockPass = os.Args[3]

	ch := make(chan []string)
	ch1 := make(chan string)

	activePorts := make([]Ports, 0)

	go ReadFile(ch)

	go LaunchForward(ch, activePorts)

	go StartListeningServer(ch1, secretPins)
	
	go ListenForJoiningHosts(ch1)

	select{}
}