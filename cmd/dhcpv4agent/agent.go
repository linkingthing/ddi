package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ben-han-cn/cement/shell"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dhcp/service"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	kg "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

const (
	StartDHCPv4               = "StartDHCPv4"
	StopDHCPv4                = "StopDHCPv4"
	CreateSubnetv4            = "CreateSubnetv4"
	UpdateSubnetv4            = "UpdateSubnetv4"
	DeleteSubnetv4            = "DeleteSubnetv4"
	CreateSubnetv4Pool        = "CreateSubnetv4Pool"
	UpdateSubnetv4Pool        = "UpdateSubnetv4Pool"
	DeleteSubnetv4Pool        = "DeleteSubnetv4Pool"
	CreateSubnetv4Reservation = "CreateSubnetv4Reservation"
	UpdateSubnetv4Reservation = "UpdateSubnetv4Reservation"
	DeleteSubnetv4Reservation = "DeleteSubnetv4Reservation"
)

var (
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
)

const (
	checkPeriod = 5
)

var dhcpv4Start bool = false
var KafkaOffsetFileDhcpv4 = "/tmp/kafka-offset-dhcpv4.txt" // store kafka offset num into this file
var KafkaOffsetDhcpv4 int64 = 0

func dhcpClient() {

	conn, err := grpc.Dial(dhcp.Dhcpv4AgentAddr, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	cliv4 := pb.NewDhcpv4ManagerClient(conn)

	kafkaReader = kg.NewReader(kg.ReaderConfig{
		Brokers: []string{utils.KafkaServerProm},
		Topic:   dhcp.Dhcpv4Topic,
		//StartOffset: 95,
	})
	var KafkaOffsetDhcpv4 int64
	size, err := ioutil.ReadFile(KafkaOffsetFileDhcpv4)
	if err == nil {
		offset, err2 := strconv.Atoi(string(size))
		if err2 != nil {
			log.Println(err2)
		}
		KafkaOffsetDhcpv4 = int64(offset)
		kafkaReader.SetOffset(KafkaOffsetDhcpv4)
	}
	log.Println("kafka Offset: ", KafkaOffsetDhcpv4)

	var message kg.Message
	ticker := time.NewTicker(checkPeriod * time.Second)
	quit := make(chan int)
	for {
		message, err = kafkaReader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
			return
		}

		l := "message at offset: " + strconv.FormatInt(message.Offset, 10) + " key: " + string(message.Key) +
			" value: " + string(message.Value)
		log.Println(l)

		//store curOffset into KafkaOffsetFile
		curOffset := kafkaReader.Stats().Offset
		if curOffset > KafkaOffsetDhcpv4 {
			KafkaOffsetDhcpv4 = curOffset
			byteOffset := []byte(strconv.Itoa(int(curOffset)))
			err = ioutil.WriteFile(KafkaOffsetFileDhcpv4, byteOffset, 0644)
			if err != nil {
				log.Println(err)
			}
		}
		switch string(message.Key) {
		case StartDHCPv4:
			var target pb.StartDHCPv4Req

			if err := proto.Unmarshal(message.Value, &target); err != nil {
				log.Fatal(err)
			}
			cliv4.StartDHCPv4(context.Background(), &target)
			go KeepDhcpv4Alive(ticker, quit)

		case StopDHCPv4:
			var target pb.StopDHCPv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
				log.Fatal(err)
			}

			cliv4.StopDHCPv4(context.Background(), &target)
			quit <- 1

		case CreateSubnetv4:
			var target pb.CreateSubnetv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.CreateSubnetv4(context.Background(), &target)

		case UpdateSubnetv4:
			var target pb.UpdateSubnetv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.UpdateSubnetv4(context.Background(), &target)

		case DeleteSubnetv4:
			var target pb.DeleteSubnetv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.DeleteSubnetv4(context.Background(), &target)

		case CreateSubnetv4Pool:
			var target pb.CreateSubnetv4PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.CreateSubnetv4Pool(context.Background(), &target)

		case UpdateSubnetv4Pool:
			var target pb.UpdateSubnetv4PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.UpdateSubnetv4Pool(context.Background(), &target)

		case DeleteSubnetv4Pool:
			var target pb.DeleteSubnetv4PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.DeleteSubnetv4Pool(context.Background(), &target)

		case CreateSubnetv4Reservation:
			var target pb.CreateSubnetv4ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.CreateSubnetv4Reservation(context.Background(), &target)

		case UpdateSubnetv4Reservation:
			var target pb.UpdateSubnetv4ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.UpdateSubnetv4Reservation(context.Background(), &target)

		case DeleteSubnetv4Reservation:
			var target pb.DeleteSubnetv4ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv4.DeleteSubnetv4Reservation(context.Background(), &target)

		}
	}
}

func KeepDhcpv4Alive(ticker *time.Ticker, quit chan int) {
	log.Print("into KeepDhcpv4Alive, return")
	return

	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(dhcp.KeaPidPath + dhcp.KeaDhcp4PidFile); err == nil {
				log.Print("dhcp4 pid file exists, continue")
				continue
			}
			param1 := "start"
			param2 := "-s" + dhcp.KEADHCPv4Service
			ret, err := shell.Shell("keactrl", param1, param2)
			if err != nil {
				log.Fatal(err)
			}
			log.Print(ret)
			return

		case <-quit:
			return
		}
	}
}

func main() {
	utils.SetHostIPs(config.YAML_CONFIG_FILE) //set global vars from yaml conf

	//ver string, ConfPath string, addr string
	s, err := server.NewDHCPv4GRPCServer(dhcp.KEADHCPv4Service, dhcp.DhcpConfigPath, dhcp.Dhcpv4AgentAddr)
	if err != nil {

		log.Fatal(err)
		return
	}
	log.Println("begin to call dhcpv4client")
	//go dhcpv4agent.Dhcpv4Client()
	go dhcpClient()

	s.Start()
	defer s.Stop()

}
