package server

import (
	"context"
	"fmt"

	"log"

	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/pb"
)

const (
	opSuccess = 0
	opFail    = 1
)

type DHCPv4Service struct {
	handler *dhcp.KEAv4Handler
}

func NewDHCPv4Service(ver string, addr string, dhcpConfPath string) *DHCPv4Service {

	log.Println("into NewDHCPv4Service")
	return &DHCPv4Service{dhcp.NewKEAv4Handler(ver, dhcpConfPath, addr)}
}

func (service *DHCPv4Service) StartDHCPv4(content context.Context, req *pb.StartDHCPv4Req) (*pb.OperResult, error) {
	log.Print("into service, startdhcpv4")
	err := service.handler.StartDHCPv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv4Service) StopDHCPv4(content context.Context, req *pb.StopDHCPv4Req) (*pb.OperResult, error) {
	log.Print("into service, stopdhcpv4")

	err := service.handler.StopDHCPv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv4Service) CreateSubnetv4(context context.Context, req *pb.CreateSubnetv4Req) (*pb.OperResult, error) {
	log.Println("into dhcp/server/service.go CreateSubnetv4()")
	err := service.handler.CreateSubnetv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv4Service) UpdateSubnetv4(context context.Context, req *pb.UpdateSubnetv4Req) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv4Service) DeleteSubnetv4(context context.Context, req *pb.DeleteSubnetv4Req) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv4(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv4Service) CreateSubnetv4Pool(context context.Context, req *pb.CreateSubnetv4PoolReq) (*pb.OperResult, error) {
	err := service.handler.CreateSubnetv4Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv4Service) UpdateSubnetv4Pool(context context.Context, req *pb.UpdateSubnetv4PoolReq) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv4Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv4Service) DeleteSubnetv4Pool(context context.Context, req *pb.DeleteSubnetv4PoolReq) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv4Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv4Service) CreateSubnetv4Reservation(context context.Context, req *pb.CreateSubnetv4ReservationReq) (*pb.OperResult, error) {
	err := service.handler.CreateSubnetv4Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv4Service) UpdateSubnetv4Reservation(context context.Context, req *pb.UpdateSubnetv4ReservationReq) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv4Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv4Service) DeleteSubnetv4Reservation(context context.Context, req *pb.DeleteSubnetv4ReservationReq) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv4Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv4Service) GetLeases(context context.Context, req *pb.GetLeasesReq) (*pb.GetLeasesResp, error) {
	resp, err := service.handler.GetLeases(*req)
	if err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

func (service *DHCPv4Service) Close() {
	service.handler.Close()
}

//dhcpv6 starts
type DHCPv6Service struct {
	handler *dhcp.KEAv6Handler
}

func newDHCPv6Service(ver string, dhcpConfPath string, addr string) *DHCPv6Service {

	return &DHCPv6Service{dhcp.NewKEAv6Handler(ver, dhcpConfPath, addr)}
}
func (service *DHCPv6Service) StartDHCPv6(content context.Context, req *pb.StartDHCPv6Req) (*pb.OperResult, error) {
	err := service.handler.StartDHCPv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv6Service) StopDHCPv6(content context.Context, req *pb.StopDHCPv6Req) (*pb.OperResult, error) {
	err := service.handler.StopDHCPv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv6Service) CreateSubnetv6(context context.Context, req *pb.CreateSubnetv6Req) (*pb.OperResult, error) {
	log.Println("into dhcp/server/service.go CreateSubnetv6()")
	err := service.handler.CreateSubnetv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv6Service) UpdateSubnetv6(context context.Context, req *pb.UpdateSubnetv6Req) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv6Service) DeleteSubnetv6(context context.Context, req *pb.DeleteSubnetv6Req) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv6(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv6Service) CreateSubnetv6Pool(context context.Context, req *pb.CreateSubnetv6PoolReq) (*pb.OperResult, error) {
	err := service.handler.CreateSubnetv6Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv6Service) UpdateSubnetv6Pool(context context.Context, req *pb.UpdateSubnetv6PoolReq) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv6Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv6Service) DeleteSubnetv6Pool(context context.Context, req *pb.DeleteSubnetv6PoolReq) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv6Pool(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv6Service) CreateSubnetv6Reservation(context context.Context, req *pb.CreateSubnetv6ReservationReq) (*pb.OperResult, error) {
	err := service.handler.CreateSubnetv6Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv6Service) UpdateSubnetv6Reservation(context context.Context, req *pb.UpdateSubnetv6ReservationReq) (*pb.OperResult, error) {
	err := service.handler.UpdateSubnetv6Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DHCPv6Service) DeleteSubnetv6Reservation(context context.Context, req *pb.DeleteSubnetv6ReservationReq) (*pb.OperResult, error) {
	err := service.handler.DeleteSubnetv6Reservation(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DHCPv6Service) Close() {
	service.handler.Close()
}
