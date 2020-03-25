package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	physicalMetrics "github.com/linkingthing/ddi/cmd/metrics"
	"github.com/linkingthing/ddi/cmd/node"
	"github.com/linkingthing/ddi/dhcp"
	businessMetrics "github.com/linkingthing/ddi/dns/metrics"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	"github.com/linkingthing/ddi/utils/grpcserver"
	kg "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

const (
	STARTDNS                  = "StartDNS"
	STOPDNS                   = "StopDNS"
	CREATEACL                 = "CreateACL"
	UPDATEACL                 = "UpdateACL"
	DELETEACL                 = "DeleteACL"
	CREATEVIEW                = "CreateView"
	UPDATEVIEW                = "UpdateView"
	DELETEVIEW                = "DeleteView"
	CREATEZONE                = "CreateZone"
	DELETEZONE                = "DeleteZone"
	CREATERR                  = "CreateRR"
	UPDATERR                  = "UpdateRR"
	DELETERR                  = "DeleteRR"
	UPDATEDEFAULTFORWARD      = "UpdateDefaultForward"
	DELETEDEFAULTFORWARD      = "DeleteDefaultForward"
	UPDATEFORWARD             = "UpdateForward"
	DELETEFORWARD             = "DeleteForward"
	CREATEREDIRECTION         = "CreateRedirection"
	UPDATEREDIRECTION         = "UpdateRedirection"
	DELETEREDIRECTION         = "DeleteRedirection"
	CREATEDEFAULTDNS64        = "CreateDefaultDNS64"
	UPDATEDEFAULTDNS64        = "UpdateDefaultDNS64"
	DELETEDEFAULTDNS64        = "DeleteDefaultDNS64"
	CREATEDNS64               = "CreateDNS64"
	UPDATEDNS64               = "UpdateDNS64"
	DELETEDNS64               = "DeleteDNS64"
	CREATEIPBLACKHOLE         = "CreateIPBlackHole"
	UPDATEIPBLACKHOLE         = "UpdateIPBlackHole"
	DELETEIPBLACKHOLE         = "DeleteIPBlackHole"
	UPDATERECURSIVECONCURRENT = "UpdateRecursiveConcurrent"
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
	kafkaServer     = "localhost:9092"
	dnsTopic        = "dns"
	kafkaWriter     *kg.Writer
	kafkaReader     *kg.Reader
	address         = "localhost:8888"
	dnsExporterPort = "8001"
)

func main() {
	utils.SetHostIPs(config.YAML_CONFIG_FILE) //set global vars from yaml conf

	handler := businessMetrics.NewMetricsHandler("/root/bindtest", 10, 10, "/root/bindtest/")
	go handler.Statics()
	go handler.DNSExporter(dnsExporterPort, "/metrics", "dns")
	yamlConfig := config.GetConfig("/etc/vanguard/vanguard.conf")
	if yamlConfig.Localhost.IsDHCP {
		go node.RegisterNode("/etc/vanguard/vanguard.conf", "dhcp")
	}
	if yamlConfig.Localhost.IsDNS {
		go node.RegisterNode("/etc/vanguard/vanguard.conf", "dns")
	}
	if !yamlConfig.Localhost.IsController {
		go physicalMetrics.NodeExporter()
	}
	s, err := grpcserver.NewGRPCServer("localhost:8888", "/root/bindtest/", "/root/bindtest/", dhcp.KEADHCPv4Service, dhcp.DhcpConfigPath, dhcp.Dhcpv4AgentAddr, yamlConfig.Localhost.IsDNS, yamlConfig.Localhost.IsDHCP)
	if err != nil {
		return
	}
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()
	if yamlConfig.Localhost.IsDNS {
		go dnsClient(conn, yamlConfig.Server.Kafka.Host+":"+yamlConfig.Server.Kafka.Port)
	}
	if yamlConfig.Localhost.IsDHCP {
		go dhcpClient(conn, yamlConfig.Server.Kafka.Host+":"+yamlConfig.Server.Kafka.Port)
	}
	s.Start()
	defer s.Stop()
}

func dnsClient(conn *grpc.ClientConn, kafkaServer string) {
	cli := pb.NewAgentManagerClient(conn)
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic:   dnsTopic,
	})
	var message kg.Message
	var err error
	for {
		message, err = kafkaReader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
			return
		}

		fmt.Println(string(message.Key))
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
		case UPDATEDEFAULTFORWARD:
			var target pb.UpdateDefaultForwardReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateDefaultForward(context.Background(), &target)
		case DELETEDEFAULTFORWARD:
			var target pb.DeleteDefaultForwardReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteDefaultForward(context.Background(), &target)
		case UPDATEFORWARD:
			var target pb.UpdateForwardReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateForward(context.Background(), &target)
		case DELETEFORWARD:
			var target pb.DeleteForwardReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteForward(context.Background(), &target)
		case CREATEREDIRECTION:
			var target pb.CreateRedirectionReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateRedirection(context.Background(), &target)
		case UPDATEREDIRECTION:
			var target pb.UpdateRedirectionReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateRedirection(context.Background(), &target)
		case DELETEREDIRECTION:
			var target pb.DeleteRedirectionReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteRedirection(context.Background(), &target)
		case CREATEDEFAULTDNS64:
			var target pb.CreateDefaultDNS64Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateDefaultDNS64(context.Background(), &target)
		case UPDATEDEFAULTDNS64:
			var target pb.UpdateDefaultDNS64Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateDefaultDNS64(context.Background(), &target)
		case DELETEDEFAULTDNS64:
			var target pb.DeleteDefaultDNS64Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteDefaultDNS64(context.Background(), &target)
		case CREATEDNS64:
			var target pb.CreateDNS64Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateDNS64(context.Background(), &target)
		case UPDATEDNS64:
			var target pb.UpdateDNS64Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateDNS64(context.Background(), &target)
		case DELETEDNS64:
			var target pb.DeleteDNS64Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteDNS64(context.Background(), &target)
		case CREATEIPBLACKHOLE:
			var target pb.CreateIPBlackHoleReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.CreateIPBlackHole(context.Background(), &target)
		case UPDATEIPBLACKHOLE:
			var target pb.UpdateIPBlackHoleReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateIPBlackHole(context.Background(), &target)
		case DELETEIPBLACKHOLE:
			var target pb.DeleteIPBlackHoleReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.DeleteIPBlackHole(context.Background(), &target)
		case UPDATERECURSIVECONCURRENT:
			var target pb.UpdateRecurConcuReq
			if err := proto.Unmarshal(message.Value, &target); err != nil {
			}
			cli.UpdateRecursiveConcurrent(context.Background(), &target)
		}
	}
}

func dhcpClient(conn *grpc.ClientConn, kafkaServer string) {
	cliv4 := pb.NewDhcpv4ManagerClient(conn)

	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers:     []string{kafkaServer},
		Topic:       dhcp.Dhcpv4Topic,
		StartOffset: 95,
	})
	var message kg.Message
	var err error
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
				log.Fatal(err)
			}
			cliv4.StartDHCPv4(context.Background(), &target)

		case StopDHCPv4:
			var target pb.StopDHCPv4Req
			if err := proto.Unmarshal(message.Value, &target); err != nil {
				log.Fatal(err)
			}

			cliv4.StopDHCPv4(context.Background(), &target)

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
