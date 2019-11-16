package dhcp

import "github.com/linkingthing.com/ddi/pb"

type DHCPHandler interface {
	StartDHCPv4(req pb.StartDHCPv4Req) error
	StopDHCPv4(req pb.StartDHCPv4Req) error
	CreateSubnet4(req pb.CreateSubnetv4Req) error
	UpdateSubnet4(req pb.UpdateSubnetv4Req) error
	DeleteSubnet4(req pb.DeleteSubnetv4Req) error

	StartDHCPv6(req pb.StartDHCPv6Req) error
	StopDHCPv6(req pb.StartDHCPv6Req) error
	CreateSubnet6(req pb.CreateSubnetv6Req) error
	UpdateSubnet6(req pb.UpdateSubnetv6Req) error
	DeleteSubnet6(req pb.DeleteSubnetv6Req) error
}
