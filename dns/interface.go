package dns

import (
	"github.com/linkingthing/ddi/pb"
)

type DNSHandler interface {
	StartDNS(req pb.DNSStartReq) error
	StopDNS() error
	CreateACL(req pb.CreateACLReq) error
	DeleteACL(req pb.DeleteACLReq) error
	CreateView(reqeust pb.CreateViewReq) error
	UpdateView(req pb.UpdateViewReq) error
	DeleteView(req pb.DeleteViewReq) error
	CreateZone(req pb.CreateZoneReq) error
	DeleteZone(req pb.DeleteZoneReq) error
	CreateRR(req pb.CreateRRReq) error
	UpdateRR(req pb.UpdateRRReq) error
	DeleteRR(req pb.DeleteRRReq) error
}
