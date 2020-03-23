package main

import (
	"context"
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
)

var (
	kafkaServer = "localhost:9092"
	dhcpv4Topic = "dhcpv4"
	dhcpv6Topic = "dhcpv6"
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
	cmd         = ""
)

const (
	STARTDNS   = "StartDNS"
	STOPDNS    = "StopDNS"
	CREATEACL  = "CreateACL"
	DELETEACL  = "DeleteACL"
	CREATEVIEW = "CreateView"
	UPDATEVIEW = "UpdateView"
	DELETEVIEW = "DeleteView"
	CREATEZONE = "CreateZone"
	DELETEZONE = "DeleteZone"
	CREATERR   = "CreateRR"
	UPDATERR   = "UpdateRR"
	DELETERR   = "DeleteRR"
)

func init() {
	kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dhcpv4Topic,
	})
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic:   dhcpv4Topic,
	})
	flag.StringVar(&cmd, "cmd", "", STARTDNS+"\n"+
		STOPDNS+"\n"+
		CREATEACL+"\n"+
		DELETEACL+"\n"+
		CREATEVIEW+"\n"+
		UPDATEVIEW+"\n"+
		DELETEVIEW+"\n"+
		CREATEZONE+"\n"+
		DELETEZONE+"\n"+
		CREATERR+"\n"+
		UPDATERR+"\n"+
		DELETERR)
}
func main() {
	flag.Parse()
	switch cmd {
	case STARTDNS:
		TestStartDNS()
	case STOPDNS:
		TestStopDNS()
	}
}

func TestStartDNS() {
	var config string = ""
	dnsStartReq := pb.DNSStartReq{Config: config}
	data, err := proto.Marshal(&dnsStartReq)
	if err != nil {
		return
	}
	postData := kg.Message{
		Key:   []byte(STARTDNS),
		Value: data,
	}
	err = kafkaWriter.WriteMessages(context.Background(), postData)
	if err != nil {
		return
	}
}

func TestStopDNS() {
	dnsStopReq := pb.DNSStopReq{}
	data, err := proto.Marshal(&dnsStopReq)
	if err != nil {
		return
	}
	postData := kg.Message{
		Key:   []byte(STOPDNS),
		Value: data,
	}
	err = kafkaWriter.WriteMessages(context.Background(), postData)
	if err != nil {
		return
	}

}
