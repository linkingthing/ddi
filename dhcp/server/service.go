package server

import (
	"context"
	"fmt"

	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/pb"
)

const (
	opSuccess = 0
	opFail    = 1
)

type DHCPService struct {
	handler *dhcp.KEAHandler
}

func newDHCPService(ver string, dhcpConfPath string, agentPath string) *DHCPService {

	return &DHCPService{dhcp.NewKEAHandler(ver, dhcpConfPath, agentPath)}
}

func (s *DHCPService) StartDHCPv4(content context.Context, req *pb.StartDHCPv4Req) (*pb.OperResult, error) {
	err := s.handler.StartDHCPv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (s *DHCPService) StopDHCPv4(content context.Context, req *pb.StopDHCPv4Req) (*pb.OperResult, error) {
	err := s.handler.StopDHCPv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (s *DHCPService) StartDHCPv6(content context.Context, req *pb.StartDHCPv6Req) (*pb.OperResult, error) {
	err := s.handler.StartDHCPv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (s *DHCPService) StopDHCPv6(content context.Context, req *pb.StopDHCPv6Req) (*pb.OperResult, error) {
	err := s.handler.StopDHCPv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) Close() {
	service.handler.Close()
}
