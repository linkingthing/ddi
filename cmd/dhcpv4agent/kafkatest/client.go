package main

import (
	"context"
	"flag"

	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
)

var (
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
	cmd         = ""
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

func init() {
	kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{dhcp.KafkaServer},
		Topic:   dhcp.Dhcpv4Topic,
	})
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{dhcp.KafkaServer},
		Topic:   dhcp.Dhcpv4Topic,
	})
	flag.StringVar(&cmd, "cmd", "", StartDHCPv4+"\n"+
		StopDHCPv4)
}
func main() {
	flag.Parse()

	fmt.Printf("cmd: %s\n", cmd)
	switch cmd {
	case StartDHCPv4:
		TestStartDHCPv4()
	case StopDHCPv4:
		TestStopDHCPv4()
	}

}

func TestStartDHCPv4() {

	fmt.Printf("---into TestStartDHCPv4\n")

	dhcpv4Req := pb.StartDHCPv4Req{Config: "StartDHCPv4"}
	data, err := proto.Marshal(&dhcpv4Req)
	if err != nil {
		fmt.Printf("---err in TestStartDHCPv4 \n")
		return
	}

	postData := kg.Message{
		Key:   []byte(StartDHCPv4),
		Value: data,
	}
	fmt.Print(postData)
	err = kafkaWriter.WriteMessages(context.Background(), postData)
	if err != nil {
		fmt.Printf("---err in TestStartDHCPv4 writemessage \n")
		return
	}
}

func TestStopDHCPv4() {
	fmt.Printf("---into TestStopDHCPv4\n")

	stopDHCPv4 := pb.StopDHCPv4Req{Config: "StopDHCPv4"}
	data, err := proto.Marshal(&stopDHCPv4)
	if err != nil {
		fmt.Print(err)
		return
	}
	postData := kg.Message{
		Key:   []byte(StopDHCPv4),
		Value: data,
	}
	err = kafkaWriter.WriteMessages(context.Background(), postData)
	if err != nil {
		return
	}

}
