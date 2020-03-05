package server

import (
	"bufio"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
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

	log.Printf("SocketServer(), Begin listen port: %d", port)
	go getKafkaMsg()
	//for host, v := range utils.OnlinePromHosts {
	//	log.Println("+++ host")
	//	log.Println(host)
	//	log.Println("--- host")
	//
	//	log.Println("+++ v")
	//	log.Println(v)
	//	log.Println("--- v")
	//}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go handler(conn)
	}

}

//handle ping pong heartbeat msg
func handler(conn net.Conn) {

	defer conn.Close()

	var (
		buf = make([]byte, 128)
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

			for ip, host := range utils.OnlinePromHosts {
				log.Println("-- ip: ", ip, ", -- data: ", data)
				if ip == data {
					now := time.Now().Unix()
					log.Println("+++ hbtime: ", host.HbTime, " +++ now: ", now)
					host.HbTime = time.Now().Unix()
				}
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

func test() {

	//get yaml config file, update global variable PromServer and localhost
	var conf *config.VanguardConf
	conf = config.GetConfig()
	log.Println("in agent.go, cur utils.promServer ip: ", utils.PromServer)
	utils.PromServer = conf.Server.Prometheus.IP
	if conf.Localhost.IP != utils.PromServer {
		utils.PromLocalInstance = conf.Localhost.IP + ":" + utils.PromLocalPort
	}

	port := utils.WebSocket_Port
	go SocketServer(port)

	mux := http.NewServeMux()
	mux.Handle("/", &myHandler{})

	mux.HandleFunc("/apis/linkingthing/node/v1/servers", list_server)
	mux.HandleFunc("/apis/linkingthing/node/v1/nodes", Query)
	mux.HandleFunc("/apis/linkingthing/node/v1/hists", Query_range)       //history
	mux.HandleFunc("/apis/linkingthing/dashboard/v1/dashdns", GetDashDns) //dns log info

	log.Println("Starting v2 httpserver")
	log.Fatal(http.ListenAndServe(":1210", mux))
	log.Println("end of main, should not come here")
}
