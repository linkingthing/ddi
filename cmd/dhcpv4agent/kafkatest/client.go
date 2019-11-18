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
	dhcpTopic   = "test"
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
	StartDHCPv6               = "StartDHCPv6"
	StopDHCPv6                = "StopDHCPv6"
)

func init() {
	kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{dhcp.KafkaServer},
		Topic:   dhcpTopic,
	})
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})
	flag.StringVar(&cmd, "cmd", "", StartDHCPv4+"\n"+
		StopDHCPv4+"\n"+
		StartDHCPv6+"\n"+
		StopDHCPv6)
}
func main() {
	flag.Parse()

	TestStartDHCPv4()
}

func TestStartDHCPv4() {

	fmt.Printf("---into TestStartDHCPv4\n")

	dhcpv4Req := pb.StartDHCPv4Req{Config: "test"}
	data, err := proto.Marshal(&dhcpv4Req)
	if err != nil {
		return
	}

	postData := kg.Message{
		Key:   []byte(StartDHCPv4),
		Value: data,
	}
	err = kafkaWriter.WriteMessages(context.Background(), postData)
	if err != nil {
		return
	}
}

//func TestStopDNS() {
//	dnsStopReq := pb.DNSStopReq{}
//	data, err := proto.Marshal(&dnsStopReq)
//	if err != nil {
//		return
//	}
//	postData := kg.Message{
//		Key:   []byte(STOPDNS),
//		Value: data,
//	}
//	err = kafkaWriter.WriteMessages(context.Background(), postData)
//	if err != nil {
//		return
//	}
//
//}
