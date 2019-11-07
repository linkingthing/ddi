package dhcp

import "github.com/linkingthing.com/ddi/pb"

type DHCPHandler interface {
	StartDHCP(req pb.DHCPStartReq) error
	StopDHCP(req pb.DHCPStopReq) error
	CreateSubnet4(req pb.CreateSubnet4Req) error
}