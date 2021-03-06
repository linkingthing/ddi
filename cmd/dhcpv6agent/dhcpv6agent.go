package main

import (
	"context"
	"io/ioutil"
	"strconv"

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
var KafkaOffsetFileDhcpv6 = "/tmp/kafka-offset-dhcpv6.txt" // store kafka offset num into this file
var KafkaOffsetDhcpv6 int64 = 0

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

	var KafkaOffsetDhcpv6 int64
	size, err := ioutil.ReadFile(KafkaOffsetFileDhcpv6)
	if err == nil {
		offset, err2 := strconv.Atoi(string(size))
		if err2 != nil {
			log.Println(err2)
		}
		KafkaOffsetDhcpv6 = int64(offset)
		kafkaReader.SetOffset(KafkaOffsetDhcpv6)
	}
	log.Println("kafka Offset: ", KafkaOffsetDhcpv6)

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

		//store curOffset into KafkaOffsetFile
		curOffset := kafkaReader.Stats().Offset
		if curOffset > KafkaOffsetDhcpv6 {
			KafkaOffsetDhcpv6 = curOffset
			byteOffset := []byte(strconv.Itoa(int(curOffset)))
			err = ioutil.WriteFile(KafkaOffsetFileDhcpv6, byteOffset, 0644)
			if err != nil {
				log.Println(err)
			}
		}
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

		case UpdateSubnetv6:
			var target pb.UpdateSubnetv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateSubnetv6(context.Background(), &target)

		case DeleteSubnetv6:
			var target pb.DeleteSubnetv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteSubnetv6(context.Background(), &target)
		case CreateSubnetv6Pool:
			var target pb.CreateSubnetv6PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateSubnetv6Pool(context.Background(), &target)

		case UpdateSubnetv6Pool:
			var target pb.UpdateSubnetv6PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateSubnetv6Pool(context.Background(), &target)

		case DeleteSubnetv6Pool:
			var target pb.DeleteSubnetv6PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteSubnetv6Pool(context.Background(), &target)
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
