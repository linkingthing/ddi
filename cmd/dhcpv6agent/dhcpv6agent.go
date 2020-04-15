package main

import (
	"context"

	"github.com/linkingthing/ddi/utils/config"

	"github.com/linkingthing/ddi/utils"

	"log"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	server "github.com/linkingthing/ddi/dhcp/service"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

const (
	StartDHCPv6               = "StartDHCPv6"
	StopDHCPv6                = "StopDHCPv6"
	CreateSubnetv6            = "CreateSubnetv6"
	UpdateSubnetv6            = "UpdateSubnetv6"
	DeleteSubnetv6            = "DeleteSubnetv6"
	CreateSubnetv6Pool        = "CreateSubnetv6Pool"
	UpdateSubnetv6Pool        = "UpdateSubnetv6Pool"
	DeleteSubnetv6Pool        = "DeleteSubnetv6Pool"
	CreateSubnetv6Reservation = "CreateSubnetv6Reservation"
	UpdateSubnetv6Reservation = "UpdateSubnetv6Reservation"
	DeleteSubnetv6Reservation = "DeleteSubnetv6Reservation"
)

var (
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
)

const (
	checkPeriod = 5
)

var dhcpv6Start bool = false

func Dhcpv6Client() {
	log.Println("in dhcpv6agent/agent.go, utils.KafkaServerProm: ", utils.KafkaServerProm)
	conn, err := grpc.Dial(utils.Dhcpv6AgentAddr, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	cli := pb.NewDhcpv6ManagerClient(conn)

	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{utils.KafkaServerProm},
		Topic:   dhcp.Dhcpv6Topic,
	})
	var message kg.Message
	//ticker := time.NewTicker(checkPeriod * time.Second)
	quit := make(chan int)
	for {
		message, err = kafkaReader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
			return
		}
		log.Println("v6 message at offset %d: key: %s, value: %s\n", message.Offset, string(message.Key), string(message.Value))

		switch string(message.Key) {
		case StartDHCPv6:
			var target pb.StartDHCPv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StartDHCPv6(context.Background(), &target)
			//go KeepDhcpv6Alive(ticker, quit)

		case StopDHCPv6:
			var target pb.StopDHCPv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StopDHCPv6(context.Background(), &target)
			quit <- 1
		case CreateSubnetv6:
			var target pb.CreateSubnetv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateSubnetv6(context.Background(), &target)
		}
	}
}

//func KeepDhcpv6Alive(ticker *time.Ticker, quit chan int) {
//	for {
//		select {
//		case <-ticker.C:
//			if _, err := os.Stat("/root/keatest/" + "named.pid"); err == nil {
//				continue
//			}
//			var param string = "-c" + "/root/keatest/" + "named.conf"
//			shell.Shell("named", param)
//
//		case <-quit:
//			return
//		}
//	}
//}

func main() {
	utils.SetHostIPs(config.YAML_CONFIG_FILE) //set global vars from yaml conf
	//yamlConfig := config.GetConfig("/etc/vanguard/vanguard.conf")
	//if yamlConfig.Localhost.IsDHCP {
	//	go node.RegisterNode("/etc/vanguard/vanguard.conf", "dhcp")
	//}
	//log.Println("yamlConfig iscontroller: ", yamlConfig.Localhost.IsController)
	//if !yamlConfig.Localhost.IsController {
	//	log.Println("begin to call node exporter")
	//	go physicalMetrics.NodeExporter()
	//}
	s, err := server.NewDHCPv6GRPCServer(dhcp.KEADHCPv6Service, dhcp.DhcpConfigPath, utils.Dhcpv6AgentAddr)
	if err != nil {

		log.Fatal(err)
		return
	}
	log.Println("begin to call dhcpv6client")

	go Dhcpv6Client()

	s.Start()
	defer s.Stop()

}
