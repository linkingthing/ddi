package main

import (
	"bufio"
	"github.com/linkingthing/ddi/utils"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Message       = "Pong"
	StopCharacter = "\r\n\r\n"
	//checkDuration = 24 * time.Hour
	checkDuration = 24 * time.Second
)

func SocketServer(port int) {

	listen, err := net.Listen("tcp4", ":"+strconv.Itoa(port))

	if err != nil {
		log.Fatalf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}

	defer listen.Close()

	log.Printf("Begin listen port: %d", port)
	go getKafkaMsg()
	for host, v := range utils.OnlinePromHosts {
		log.Println("+++ host")
		log.Println(host)
		log.Println("--- host")

		log.Println("+++ v")
		log.Println(v)
		log.Println("--- v")
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go handler(conn)
	}

}

func handler(conn net.Conn) {

	defer conn.Close()

	var (
		buf = make([]byte, 1024)
		r   = bufio.NewReader(conn)
		w   = bufio.NewWriter(conn)
	)

ILOOP:
	for {
		n, err := r.Read(buf)
		data := string(buf[:n])

		switch err {
		case io.EOF:
			break ILOOP
		case nil:
			log.Println("Receive:", data)
			if isTransportOver(data) {
				break ILOOP
			}

		default:
			log.Fatalf("Receive data failed:%s", err)
			return
		}

	}
	w.Write([]byte(Message))
	w.Flush()
	log.Printf("Send: %s", Message)

}

func isTransportOver(data string) (over bool) {
	over = strings.HasSuffix(data, "\r\n\r\n")
	return
}

func main() {

	port := 3333
	go SocketServer(port)

	mux := http.NewServeMux()
	mux.Handle("/", &myHandler{})
	mux.HandleFunc("/apis/linkingthing/node/v1/nodes", query)
	mux.HandleFunc("/apis/linkingthing/node/v1/hists", query_range) //history

	log.Println("Starting v2 httpserver")
	log.Fatal(http.ListenAndServe(":1210", mux))

	log.Println("end of main, should not come here")
}