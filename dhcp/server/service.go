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

func (service *DHCPService) StartDHCPv4(content context.Context, req *pb.StartDHCPv4Req) (*pb.OperResult, error) {
	err := service.handler.StartDHCPv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) StopDHCPv4(content context.Context, req *pb.StopDHCPv4Req) (*pb.OperResult, error) {
	err := service.handler.StopDHCPv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) CreateSubnetv4(context context.Context, req *pb.CreateSubnetv4Req) (*pb.OperResult, error) {
	err := service.handler.CreateSubnetv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPService) UpdateSubnetv4(context context.Context, req *pb.UpdateSubnetv4Req) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPService) DeleteSubnetv4(context context.Context, req *pb.DeleteSubnetv4Req) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) CreateSubnetv4Pool(context context.Context, req *pb.CreateSubnetv4PoolReq) (*pb.OperResult, error) {
	err := service.handler.CreateSubnetv4Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPService) UpdateSubnetv4Pool(context context.Context, req *pb.UpdateSubnetv4PoolReq) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv4Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPService) DeleteSubnetv4Pool(context context.Context, req *pb.DeleteSubnetv4PoolReq) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv4Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) CreateSubnetv4Reservation(context context.Context, req *pb.CreateSubnetv4ReservationReq) (*pb.OperResult, error) {
	err := service.handler.CreateSubnetv4Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPService) UpdateSubnetv4Reservation(context context.Context, req *pb.UpdateSubnetv4ReservationReq) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv4Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPService) DeleteSubnetv4Reservation(context context.Context, req *pb.DeleteSubnetv4ReservationReq) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv4Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) StartDHCPv6(content context.Context, req *pb.StartDHCPv6Req) (*pb.OperResult, error) {
	err := service.handler.StartDHCPv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) StopDHCPv6(content context.Context, req *pb.StopDHCPv6Req) (*pb.OperResult, error) {
	err := service.handler.StopDHCPv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPService) Close() {
	service.handler.Close()
}
