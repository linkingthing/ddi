package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
)

const (
	message       = "IP_"
	StopCharacter = "\r\n\r\n"
)

func SocketClient(ip string, port int) {
	addr := strings.Join([]string{ip, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	defer conn.Close()

	conn.Write([]byte(ip))
	conn.Write([]byte(StopCharacter))
	log.Printf("Send: %s", message)

	buff := make([]byte, 128)
	n, _ := conn.Read(buff)
	log.Printf("Receive: %s", buff[:n])

}

func main() {

	//get promServer from yaml config file
	ip := config.GetLocalIP("/etc/vanguard/vanguard.conf")
	port := utils.WebSocket_Port

	SocketClient(ip, port)

}
