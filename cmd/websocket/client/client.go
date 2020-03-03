package main

import (
	"github.com/linkingthing/ddi/utils"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
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
	ip := utils.GetLocalIP()
	port := utils.WebSocket_Port

	SocketClient(ip, port)

}
