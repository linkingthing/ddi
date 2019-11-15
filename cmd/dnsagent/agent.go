package main

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dns"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
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

var (
	kafkaServer = "localhost:9092"
	dhcpTopic   = "test"
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
)

func main() {
	var handler dns.DNSHandler
	p := dns.NewBindHandler("/root/bindtest/", "/root/bindtest/")
	handler = p
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})
	var err error
	var message kg.Message
	for {
		message, err = kafkaReader.ReadMessage(context.Background())
		if err != nil {
			return
		}
		switch string(message.Key) {
		case STARTDNS:
			var target pb.DNSStartReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.StartDNS(target)
		case STOPDNS:
			handler.StopDNS()
		case CREATEACL:
			var target pb.CreateACLReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.CreateACL(target)
		case DELETEACL:
			var target pb.DeleteACLReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.DeleteACL(target)
		case CREATEVIEW:
			var target pb.CreateViewReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.CreateView(target)
		case UPDATEVIEW:
			var target pb.UpdateViewReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.UpdateView(target)
		case DELETEVIEW:
			var target pb.DeleteViewReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.DeleteView(target)
		case CREATEZONE:
			var target pb.CreateZoneReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.CreateZone(target)
		case DELETEZONE:
			var target pb.DeleteZoneReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.DeleteZone(target)
		case CREATERR:
			var target pb.CreateRRReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.CreateRR(target)
		case UPDATERR:
			var target pb.UpdateRRReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.UpdateRR(target)
		case DELETERR:
			var target pb.DeleteRRReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			handler.DeleteRR(target)
		}
	}
}
