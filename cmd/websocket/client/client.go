package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
)

const (
	message       = "IP_"
	StopCharacter = "\r\n\r\n"
)

func SocketClient(ip string, port int, role string) {
	addr := strings.Join([]string{ip, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Println("err != nil")
		log.Fatalln(err)
		os.Exit(1)
	}

	defer conn.Close()

	msg := ip + "_" + role
	conn.Write([]byte(msg))
	conn.Write([]byte(StopCharacter))
	log.Printf("Send: %s", msg)

	buff := make([]byte, 128)
	n, _ := conn.Read(buff)
	log.Printf("Receive: %s", buff[:n])

}

func main() {

	//get promServer from yaml config file
	yamlConfig := config.GetConfig("/etc/vanguard/vanguard.conf")
	ip := yamlConfig.Localhost.IP
	role := yamlConfig.Localhost.Role
	port := utils.WebSocket_Port

	for {
		SocketClient(ip, port, role)
		time.Sleep(5 * time.Second)
	}

}
