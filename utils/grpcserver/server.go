package grpcserver

import (
	dhcpservice "github.com/linkingthing/ddi/dhcp/service"
	dnsservice "github.com/linkingthing/ddi/dns/service"
	"github.com/linkingthing/ddi/pb"
	//"github.com/linkingthing/ddi/utils"
	"google.golang.org/grpc"
	"net"
)

type GRPCServer struct {
	dnsService   *dnsservice.DNSService
	dhcp4Service *dhcpservice.DHCPv4Service
	server       *grpc.Server
	listener     net.Listener
}

func NewGRPCServer(addr string, ConfPath string, agentPath string, dhcp4Ver string, dhcp4ConfPath string, dhcp4Addr string, isDnsOpen bool, isDhcpOpen bool) (*GRPCServer, error) {
	server := grpc.NewServer()
	var dnsService *dnsservice.DNSService
	var dhcp4Service *dhcpservice.DHCPv4Service
	/*if utils.NodeRole == "dns" {
		dnsService = dnsservice.NewDNSService(ConfPath, agentPath)
	}
	if utils.NodeRole == "dhcp" {
		dhcp4Service = dhcpservice.NewDHCPv4Service(dhcp4Ver, dhcp4Addr, dhcp4ConfPath)
	}*/
	if isDnsOpen {
		dnsService = dnsservice.NewDNSService(ConfPath, agentPath)
	}
	if isDhcpOpen {
		dhcp4Service = dhcpservice.NewDHCPv4Service(dhcp4Ver, dhcp4Addr, dhcp4ConfPath)
	}
	pb.RegisterAgentManagerServer(server, dnsService)
	pb.RegisterDhcpv4ManagerServer(server, dhcp4Service)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &GRPCServer{
		dnsService:   dnsService,
		dhcp4Service: dhcp4Service,
		server:       server,
		listener:     listener,
	}, nil
}

func (s *GRPCServer) Start() error {
	return s.server.Serve(s.listener)
}

func (s *GRPCServer) Stop() error {
	s.server.GracefulStop()
	s.dnsService.Close()
	s.dhcp4Service.Close()
	return nil
}
