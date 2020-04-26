package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/linkingthing/ddi/utils"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
)

var (
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
	cmd         = ""
)

const (
	StartDHCPv6 = "StartDHCPv6"
	StopDHCPv6  = "StopDHCPv6"
)

func init() {
	kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{utils.KafkaServer},
		Topic:   utils.Dhcpv6Topic,
	})
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{utils.KafkaServer},
		Topic:   utils.Dhcpv6Topic,
	})
	flag.StringVar(&cmd, "cmd", "", StartDHCPv6+"\n"+
		StopDHCPv6)
}
func main() {
	flag.Parse()

	TestStartDHCPv6()
}

func TestStartDHCPv6() {

	fmt.Printf("---into TestStartDHCPv6\n")

	dhcpv4Req := pb.StartDHCPv4Req{Config: "StartDHCPv6--test"}
	data, err := proto.Marshal(&dhcpv4Req)
	if err != nil {
		return
	}

	postData := kg.Message{
		Key:   []byte(StartDHCPv6),
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
