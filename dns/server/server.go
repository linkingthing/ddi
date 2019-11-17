package server

import (
	"github.com/linkingthing/ddi/pb"
	"google.golang.org/grpc"
	"net"
)

type DNSGRPCServer struct {
	service  *DNSService
	server   *grpc.Server
	listener net.Listener
}

func NewDNSGRPCServer(addr string, ConfPath string, agentPath string) (*DNSGRPCServer, error) {
	server := grpc.NewServer()
	service := newDNSService(ConfPath, agentPath)
	pb.RegisterAgentManagerServer(server, service)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &DNSGRPCServer{
		service:  service,
		server:   server,
		listener: listener,
	}, nil
}

func (s *DNSGRPCServer) Start() error {
	return s.server.Serve(s.listener)
}

func (s *DNSGRPCServer) Stop() error {
	s.server.GracefulStop()
	s.service.Close()
	return nil
}
