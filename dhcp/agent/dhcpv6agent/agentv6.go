package dhcpv6agent

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/linkingthing/ddi/utils"

	"github.com/ben-han-cn/cement/shell"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
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

var dhcpv4Start bool = false

func KeepDhcpv6Alive(ticker *time.Ticker, quit chan int) {
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

func Dhcpv6Client() {

	log.Println("into Dhcpv6Client()")
	conn, err := grpc.Dial(utils.Dhcpv4AgentAddr, grpc.WithInsecure())
	if err != nil {
		log.Println("dhcp连接grpc服务错误 ", err)
		return
	}
	defer conn.Close()
	cliv6 := pb.NewDhcpv6ManagerClient(conn)

	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{utils.KafkaServer},
		Topic:   utils.Dhcpv6Topic,
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
		log.Println("Dhcpv6Client v6 message at offset %d: key: %s, value: %s\n", message.Offset,
			string(message.Key), string(message.Value))

		switch string(message.Key) {
		case StartDHCPv6:
			var target pb.StartDHCPv6Req

			if err := proto.Unmarshal(message.Value, &target); err != nil {
				log.Fatal(err)
			}
			cliv6.StartDHCPv6(context.Background(), &target)
			go KeepDhcpv6Alive(ticker, quit)

		case StopDHCPv6:
			var target pb.StopDHCPv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
				log.Fatal(err)
			}

			cliv6.StopDHCPv6(context.Background(), &target)
			quit <- 1

		case CreateSubnetv6:
			var target pb.CreateSubnetv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.CreateSubnetv6(context.Background(), &target)

		case UpdateSubnetv6:
			var target pb.UpdateSubnetv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.UpdateSubnetv6(context.Background(), &target)

		case DeleteSubnetv6:
			var target pb.DeleteSubnetv6Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.DeleteSubnetv6(context.Background(), &target)

		case CreateSubnetv6Pool:
			var target pb.CreateSubnetv6PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.CreateSubnetv6Pool(context.Background(), &target)

		case UpdateSubnetv6Pool:
			var target pb.UpdateSubnetv6PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.UpdateSubnetv6Pool(context.Background(), &target)

		case DeleteSubnetv6Pool:
			var target pb.DeleteSubnetv6PoolReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.DeleteSubnetv6Pool(context.Background(), &target)

		case CreateSubnetv6Reservation:
			var target pb.CreateSubnetv6ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.CreateSubnetv6Reservation(context.Background(), &target)

		case UpdateSubnetv6Reservation:
			var target pb.UpdateSubnetv6ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.UpdateSubnetv6Reservation(context.Background(), &target)

		case DeleteSubnetv6Reservation:
			var target pb.DeleteSubnetv6ReservationReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cliv6.DeleteSubnetv6Reservation(context.Background(), &target)

		}
	}
}
