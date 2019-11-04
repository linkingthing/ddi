package dns

import (
	"github.com/linkingthing/ddi/pb"
)

type DNSHandler interface {
	StartDNS(request pb.DNSStartReq) error
	StopDNS()
	CreateACL(request pb.CreateACLReq)
	DeleteACL(request pb.DeleteACLReq)
	CreateView(reqeust pb.CreateViewReq)
	UpdateView(request pb.UpdateViewReq)
	DeleteView(request pb.DeleteViewReq)
	CreateZone(request pb.CreateZoneReq)
	DeleteZone(request pb.DeleteZoneReq)
	CreateRR(request pb.CreateRRReq)
	UpdateRR(request pb.UpdateRRReq)
	DeleteRR(request pb.DeleteRRReq)
}
