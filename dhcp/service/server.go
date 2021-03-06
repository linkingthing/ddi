package server

import (
	"log"
	"net"

	"fmt"

	"github.com/linkingthing/ddi/pb"
	"google.golang.org/grpc"
)

type DHCPv4GRPCServer struct {
	service  *DHCPv4Service
	server   *grpc.Server
	listener net.Listener
}

func NewDHCPv4GRPCServer(ver string, ConfPath string, addr string) (*DHCPv4GRPCServer, error) {
	log.Println("into NewDHCPv4GRPCServer, ver: ", ver)
	log.Println("into NewDHCPv4GRPCServer, ConfPath: ", ConfPath)
	log.Println("into NewDHCPv4GRPCServer, addr: ", addr)

	server := grpc.NewServer()
	servicev4 := NewDHCPv4Service(ver, addr, ConfPath)
	log.Println("after NewDHCPv4Service")
	pb.RegisterDhcpv4ManagerServer(server, servicev4)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	return &DHCPv4GRPCServer{
		service:  servicev4,
		server:   server,
		listener: listener,
	}, nil
}

func (s *DHCPv4GRPCServer) Start() error {
	return s.server.Serve(s.listener)
}

func (s *DHCPv4GRPCServer) Stop() error {
	s.server.GracefulStop()
	s.service.Close()
	return nil
}

// dhcpv6 begins
type DHCPv6GRPCServer struct {
	service  *DHCPv6Service
	server   *grpc.Server
	listener net.Listener
}

func NewDHCPv6GRPCServer(ver string, ConfPath string, addr string) (*DHCPv6GRPCServer, error) {
	server := grpc.NewServer()
	servicev6 := newDHCPv6Service(ver, addr, ConfPath)
	pb.RegisterDhcpv6ManagerServer(server, servicev6)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &DHCPv6GRPCServer{
		service:  servicev6,
		server:   server,
		listener: listener,
	}, nil
}

func (s *DHCPv6GRPCServer) Start() error {
	return s.server.Serve(s.listener)
}

func (s *DHCPv6GRPCServer) Stop() error {
	s.server.GracefulStop()
	s.service.Close()
	return nil
}
