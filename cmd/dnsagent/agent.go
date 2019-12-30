package main

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dns/server"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

const (
	STARTDNS   = "StartDNS"
	STOPDNS    = "StopDNS"
	CREATEACL  = "CreateACL"
	UPDATEACL  = "UpdateACL"
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
	address     = "localhost:8888"
)

func main() {
	go dnsClient()
	s, err := server.NewDNSGRPCServer("localhost:8888", "/root/bindtest/", "/root/bindtest/")
	if err != nil {
		return
	}
	s.Start()
	defer s.Stop()
}

func dnsClient() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	cli := pb.NewAgentManagerClient(conn)
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})
	var message kg.Message
	for {
		message, err = kafkaReader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
			return
		}

		switch string(message.Key) {
		case STARTDNS:
			var target pb.DNSStartReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StartDNS(context.Background(), &target)
		case STOPDNS:
			var target pb.DNSStopReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.StopDNS(context.Background(), &target)
		case CREATEACL:
			var target pb.CreateACLReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateACL(context.Background(), &target)
		case UPDATEACL:
			var target pb.UpdateACLReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateACL(context.Background(), &target)
		case DELETEACL:
			var target pb.DeleteACLReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteACL(context.Background(), &target)
		case CREATEVIEW:
			var target pb.CreateViewReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateView(context.Background(), &target)
		case UPDATEVIEW:
			var target pb.UpdateViewReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateView(context.Background(), &target)
		case DELETEVIEW:
			var target pb.DeleteViewReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteView(context.Background(), &target)
		case CREATEZONE:
			var target pb.CreateZoneReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateZone(context.Background(), &target)
		case DELETEZONE:
			var target pb.DeleteZoneReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteZone(context.Background(), &target)
		case CREATERR:
			var target pb.CreateRRReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateRR(context.Background(), &target)
		case UPDATERR:
			var target pb.UpdateRRReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateRR(context.Background(), &target)
		case DELETERR:
			var target pb.DeleteRRReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteRR(context.Background(), &target)
		}
	}
}
