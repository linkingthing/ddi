package main

import (
	"context"
	"os"
	"time"

	"github.com/ben-han-cn/cement/shell"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp/server"
	"github.com/linkingthing/ddi/pb"
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
	StartDHCPv6               = "StartDHCPv6"
	StopDHCPv6                = "StopDHCPv6"
)

var (
	kafkaServer = "localhost:9092"
	dhcpTopic   = "test"
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
	address     = "localhost:8888"
)

const (
	checkPeriod = 5
)

var dhcpv4Start bool = false
var dhcpv6Start bool = false

func dhcpClient() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	cli := pb.NewDhcpManagerClient(conn)
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})
	var message kg.Message
	ticker := time.NewTicker(checkPeriod * time.Second)
	quit := make(chan int)
	for {
		message, err = kafkaReader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
			return
		}

		switch string(message.Key) {
		case StartDHCPv4:
			var target pb.StartDHCPv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StartDHCPv4(context.Background(), &target)
			go KeepDhcpv4Alive(ticker, quit)

		case StopDHCPv4:
			var target pb.StopDHCPv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StopDHCPv4(context.Background(), &target)
			quit <- 1

		case CreateSubnetv4:
			var target pb.CreateSubnetv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateSubnetv4(context.Background(), &target)

		case UpdateSubnetv4:
			var target pb.UpdateSubnetv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateSubnetv4(context.Background(), &target)

		case DeleteSubnetv4:
			var target pb.DeleteSubnetv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteSubnetv4(context.Background(), &target)

		case CreateSubnetv4Pool:
			var target pb.CreateSubnetv4PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateSubnetv4Pool(context.Background(), &target)

		case UpdateSubnetv4Pool:
			var target pb.UpdateSubnetv4PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateSubnetv4Pool(context.Background(), &target)

		case DeleteSubnetv4Pool:
			var target pb.DeleteSubnetv4PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteSubnetv4Pool(context.Background(), &target)

		case CreateSubnetv4Reservation:
			var target pb.CreateSubnetv4ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateSubnetv4Reservation(context.Background(), &target)

		case UpdateSubnetv4Reservation:
			var target pb.UpdateSubnetv4ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateSubnetv4Reservation(context.Background(), &target)

		case DeleteSubnetv4Reservation:
			var target pb.DeleteSubnetv4ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteSubnetv4Reservation(context.Background(), &target)

		case StartDHCPv6:
			var target pb.StartDHCPv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StartDHCPv6(context.Background(), &target)
			go KeepDhcpv6Alive(ticker, quit)

		case StopDHCPv6:
			var target pb.StopDHCPv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StopDHCPv6(context.Background(), &target)
			quit <- 2
		}
	}
}

func KeepDhcpv4Alive(ticker *time.Ticker, quit chan int) {
	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat("/root/bindtest/" + "named.pid"); err == nil {
				continue
			}
			var param string = "-c" + "/root/bindtest/" + "named.conf"
			shell.Shell("named", param)

		case <-quit:
			return
		}
	}
}

func KeepDhcpv6Alive(ticker *time.Ticker, quit chan int) {
	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat("/root/bindtest/" + "named.pid"); err == nil {
				continue
			}
			var param string = "-c" + "/root/bindtest/" + "named.conf"
			shell.Shell("named", param)

		case <-quit:
			return
		}
	}
}

func main() {
	go dhcpClient()
	s, err := server.NewDHCPGRPCServer("localhost:8888", "/root/bindtest/", "/root/bindtest/")
	if err != nil {
		return
	}
	s.Start()
	defer s.Stop()
}
