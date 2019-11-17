package server

import (
	"context"
	"fmt"
	"github.com/linkingthing/ddi/dns"
	"github.com/linkingthing/ddi/pb"
)

const (
	opSuccess = 0
	opFail    = 1
)

type DNSService struct {
	handler *dns.BindHandler
}

func newDNSService(dnsConfPath string, agentPath string) *DNSService {
	return &DNSService{dns.NewBindHandler(dnsConfPath, agentPath)}
}

func (service *DNSService) StartDNS(content context.Context, req *pb.DNSStartReq) (*pb.OperResult, error) {
	err := service.handler.StartDNS(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}

func (service *DNSService) StopDNS(context context.Context, req *pb.DNSStopReq) (*pb.OperResult, error) {
	err := service.handler.StopDNS()
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) CreateACL(context context.Context, req *pb.CreateACLReq) (*pb.OperResult, error) {
	err := service.handler.CreateACL(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) DeleteACL(context context.Context, req *pb.DeleteACLReq) (*pb.OperResult, error) {
	err := service.handler.DeleteACL(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) CreateView(context context.Context, req *pb.CreateViewReq) (*pb.OperResult, error) {
	err := service.handler.CreateView(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) UpdateView(context context.Context, req *pb.UpdateViewReq) (*pb.OperResult, error) {
	err := service.handler.UpdateView(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) DeleteView(context context.Context, req *pb.DeleteViewReq) (*pb.OperResult, error) {
	err := service.handler.DeleteView(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) CreateZone(context context.Context, req *pb.CreateZoneReq) (*pb.OperResult, error) {
	err := service.handler.CreateZone(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) DeleteZone(context context.Context, req *pb.DeleteZoneReq) (*pb.OperResult, error) {
	err := service.handler.DeleteZone(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) CreateRR(context context.Context, req *pb.CreateRRReq) (*pb.OperResult, error) {
	err := service.handler.CreateRR(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) UpdateRR(context context.Context, req *pb.UpdateRRReq) (*pb.OperResult, error) {
	err := service.handler.UpdateRR(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) DeleteRR(context context.Context, req *pb.DeleteRRReq) (*pb.OperResult, error) {
	err := service.handler.DeleteRR(*req)
	if err != nil {
		return &pb.OperResult{RetCode: opFail, RetMsg: fmt.Sprintf("%s", err)}, err
	} else {
		return &pb.OperResult{RetCode: opSuccess}, nil
	}
}
func (service *DNSService) Close() {
	service.handler.Close()
}
