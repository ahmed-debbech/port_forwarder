package main

import (
	_ "fmt"
	_ "io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	AddPreroutingCmd  = "iptables -t nat -A PREROUTING -p &&prot --dport &&port_dest -j DNAT --to-destination &&ip_dest:&&port_dest"
	AddPostroutingCmd = "iptables -t nat -A POSTROUTING -p &&prot -d &&ip_dest -j SNAT --to-source &&ip_this"
	DelPreroutingCmd  = "iptables -t nat -D PREROUTING -p &&prot --dport &&port_dest -j DNAT --to-destination &&ip_dest:&&port_dest"
	DelPostroutingCmd = "iptables -t nat -D POSTROUTING -p &&prot -d &&ip_dest -j SNAT --to-source &&ip_this"
)

var (
	LocalIp          string
	secretPins       []string
	JoiningHosts     map[string]string
	UnlockPass       string
	PortsFileContent []string
	MuPortFile       sync.Mutex
	FilePath = "cmd_log"
)

type PortToCmd struct {
	Port string 
	Cmd string
}

type Ports struct {
	PortNumber string
	HostName   string
	HostIp     string
	CmdStart   []string
	CmdStop    []string
}

func BindCmd(command string, ip_dest string, port_dest string, ip_this string, prot string) string {
	command = strings.ReplaceAll(command, "&&ip_dest", ip_dest)
	command = strings.ReplaceAll(command, "&&port_dest", port_dest)
	command = strings.ReplaceAll(command, "&&ip_this", ip_this)
	command = strings.ReplaceAll(command, "&&prot", prot)
	return command
}

func ReadFile(ch chan []string) {

	for {
		MuPortFile.Lock()

		dat, err := os.ReadFile("PORTS")
		if err != nil {
			panic("could not read PORTS files")
		}
		MuPortFile.Unlock()

		ls := strings.Split(string(dat), "\n")

		l := make([]string, 0)

		if len(ls) == 0 {
			ch <- ls
			continue
		}

		for i := 0; i <= len(ls)-1; i++ {
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

func LaunchForward(ch chan []string, activePorts []Ports) {

	enableForwarding()
	UndoCachedCommands()

	for {
		hosts := <-ch
		PortsFileContent = hosts

		newPortsToOpen := make([]Ports, 0)

		for i := 0; i <= len(hosts)-1; i += 2 {

			newPort := hosts[i+1]
			newHost := hosts[i]

			if _, ok := JoiningHosts[newHost]; !ok {
				newHost = ""
				continue
			}

			for j := 0; j <= len(activePorts)-1; j++ {

				if hosts[i+1] == activePorts[j].PortNumber &&
					JoiningHosts[hosts[i]] == activePorts[j].HostIp {
					newPort = ""
					break
				}
			}
			if newPort != "" {
				cs := []string{
					BindCmd(AddPreroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp, "tcp"),
					BindCmd(AddPostroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp, "tcp"),
				}
				ce := []string{
					BindCmd(DelPreroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp, "tcp"),
					BindCmd(DelPostroutingCmd, JoiningHosts[hosts[i]], hosts[i+1], LocalIp, "tcp"),
				}
				ps := Ports{
					PortNumber: newPort,
					HostName:   hosts[i],
					HostIp:     JoiningHosts[hosts[i]],
					CmdStart:   cs,
					CmdStop:    ce,
				}
				activePorts = append(activePorts, ps)
				newPortsToOpen = append(newPortsToOpen, ps)
			}
		}

		//now delete ports not in the PORTS file or that has changed their ips
		for i := 0; i <= len(activePorts)-1; i++ {
			stillexist := false
			for j := 0; j <= len(hosts)-1; j += 2 {
				if activePorts[i].PortNumber == hosts[j+1] &&
					activePorts[i].HostIp == JoiningHosts[hosts[j]] {
					stillexist = true
					break
				}
			}
			if !stillexist {
				//delete Port
				fireIptablesCommands(activePorts[i].CmdStop[0])
				RemoveFromLoggedCmds(activePorts[i].PortNumber+"PRE")
				fireIptablesCommands(activePorts[i].CmdStop[1])
				RemoveFromLoggedCmds(activePorts[i].PortNumber+"POST")
				log.Println("[DELETING]", activePorts[i].HostIp, ":", activePorts[i].PortNumber)
				activePorts = append(activePorts[:i], activePorts[i+1:]...)
			}
		}

		/// launch activePorts
		for i := 0; i <= len(newPortsToOpen)-1; i++ {
			success := fireIptablesCommands(newPortsToOpen[i].CmdStart[0])
			if success {
				AppendToLoggedCmds(newPortsToOpen[i].PortNumber+"PRE", newPortsToOpen[i].CmdStart[0])
			}
			success = fireIptablesCommands(newPortsToOpen[i].CmdStart[1])
			if success {
				AppendToLoggedCmds(newPortsToOpen[i].PortNumber+"POST", newPortsToOpen[i].CmdStart[1])
			}

			log.Println("[ADDING]", newPortsToOpen[i].HostIp, ":", newPortsToOpen[i].PortNumber)
		}
	}
}

func UndoCachedCommands(){
	file := createOrOpenLoggedCmdsFile()
	if file != nil {
		return
	}
	defer file.Close()

	data, err := os.ReadFile(FilePath)
	if err != nil {
		log.Println("[WARNING] could not read LoggedCmds file to remove commands")
		return
	}

	[]PortToCmd structureReadData(data)
	
}

func enableForwarding() {
	cmd := exec.Command("sh", "-c", "echo 1 > /proc/sys/net/ipv4/ip_forward")
	if err := cmd.Start(); err != nil {
		log.Println("could not enable forwarding in kernel")
		os.Exit(1)
	}
}
func fireIptablesCommands(command string) bool {
	cmd := exec.Command("iptables", strings.Split(command, " ")[1:]...)
	if err := cmd.Start(); err != nil {
		log.Println("command", command, "is not executing")
		return false
	}
	return true
}

func createOrOpenLoggedCmdsFile() *os.File{

	file, err := os.OpenFile(FilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend);
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(FilePath)
			if err != nil {
				log.Printf("Error creating Logged Commands file beacuse: %v\n", err)
				return nil
			} else {
				log.Printf("File of Logged commands created.\n")
			}
		} else {
			log.Printf("Error opening file of Logged commands '%s': %v\n", FilePath, err)
			return nil
		}
	}
	return file
}

func structureReadData(data []byte) []PortToCmd{

	portToCmds := make([]PortToCmd, 0)

	lines := strings.Split(string(data), "\n")
	for _, v := range lines {
		if v == "" {continue}
		ptc := PortToCmd{
			Port: strings.Split(v, ":")[0],
			Cmd: strings.Join(strings.Split(v, ":")[1:], ","),
		}
		portToCmds = append(portToCmds, ptc)
	}

	return portToCmds
}

func structureWriteData(portsToCmdArr []PortToCmd) string{

	finalFileOutput := ""
	for _, v := range portsToCmdArr {
		finalFileOutput += v.Port + ":" + v.Cmd +"\n"
	}
	return finalFileOutput
}


func RemoveFromLoggedCmds(port string) {
	file := createOrOpenLoggedCmdsFile()
	if file == nil {
		return
	}

	data, err := os.ReadFile(FilePath)
	if err != nil {
		log.Println("[WARNING] could not read LoggedCmds file to remove commands")
		return
	}
	file.Close() 

	updatedPortToCmds := structureReadData(data)
	for _, v := range updatedPortToCmds {
		if !strings.HasPrefix(v.Port, port) {
			updatedPortToCmds = append(updatedPortToCmds, v)
		}
	}

	finalFileOutput := structureWriteData(updatedPortToCmds)

	err = os.WriteFile(FilePath, []byte(finalFileOutput), 0644)
	if err != nil {
		log.Println("[WARNING] could not write LoggedCmds file")
		return
	}
	
}

func AppendToLoggedCmds(port string, command string) {

	file := createOrOpenLoggedCmdsFile()
	if file == nil { 
		log.Println("[WARNING] command started and added to iptables but NOT logged!")
		return
	}
	defer file.Close() 

	_, err := file.WriteString(port + ":" +command + "\n")
	if err != nil {
		log.Printf("Error opening file of Logged commands '%s': %v\n", FilePath, err)
		log.Println("[WARNING] command started and added to iptables but NOT logged!")
	}
}

func ListenForJoiningHosts(ch chan string) {

	mapSize := 0
	for {

		pin_ip := <-ch

		dd := strings.Split(pin_ip, "&")

		JoiningHosts[dd[0]] = dd[1]
		
		if mapSize != len(JoiningHosts) {
			mapSize = len(JoiningHosts)
			log.Println("current joined hosts", JoiningHosts)
		}
	}
}

func DividSecretPins(pins string) []string {
	return strings.Split(pins, "-")
}

func main() {
	log.Println("Broadcast v1")

	if len(os.Args) < 4 {
		log.Println("Usage: main_forwarder localipWithoutPort secret_pin*6chars-secret_pin2-secret_pin3... unlockPass")
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

	select {}
}
