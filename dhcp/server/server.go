package server

import (
	"net"

	"github.com/linkingthing/ddi/pb"
	"google.golang.org/grpc"
)

type DHCPGRPCServer struct {
	service  *DHCPService
	server   *grpc.Server
	listener net.Listener
}

func NewDHCPGRPCServer(addr string, ConfPath string, agentPath string) (*DHCPGRPCServer, error) {
	server := grpc.NewServer()
	service := newDHCPService(addr, ConfPath, agentPath)
	pb.RegisterDhcpManagerServer(server, service)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &DHCPGRPCServer{
		service:  service,
		server:   server,
		listener: listener,
	}, nil
}

func (s *DHCPGRPCServer) Start() error {
	return s.server.Serve(s.listener)
}

func (s *DHCPGRPCServer) Stop() error {
	s.server.GracefulStop()
	s.service.Close()
	return nil
}
