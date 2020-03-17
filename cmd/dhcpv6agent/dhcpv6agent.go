package main

import (
	"context"
	"os"
	"time"

	"fmt"

	"github.com/ben-han-cn/cement/shell"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dhcp/server"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

const (
	StartDHCPv6 = "StartDHCPv6"
	StopDHCPv6  = "StopDHCPv6"
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
	conn, err := grpc.Dial(dhcp.Dhcpv6AgentAddr, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	cli := pb.NewDhcpv6ManagerClient(conn)

	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{dhcp.KafkaServer},
		Topic:   dhcp.Dhcpv6Topic,
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
		fmt.Printf("v6 message at offset %d: key: %s, value: %s\n", message.Offset, string(message.Key), string(message.Value))

		switch string(message.Key) {
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
			quit <- 1

		}
	}
}

func KeepDhcpv6Alive(ticker *time.Ticker, quit chan int) {
	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat("/root/keatest/" + "named.pid"); err == nil {
				continue
			}
			var param string = "-c" + "/root/keatest/" + "named.conf"
			shell.Shell("named", param)

		case <-quit:
			return
		}
	}
}

func main() {
	go Dhcpv6Client()
	s, err := server.NewDHCPv6GRPCServer(dhcp.KEADHCPv6Service, dhcp.DhcpConfigPath, dhcp.Dhcpv6AgentAddr)
	if err != nil {
		return
	}
	s.Start()
	defer s.Stop()

}
