package dhcpv6

import (
	"testing"
	"time"

	"github.com/linkingthing.com/ddi/dhcp"
	"github.com/linkingthing.com/ddi/pb"

	ut "github.com/ben-han-cn/cement/unittest"
)

var handlerv6 = &KEAv6Handler{
	ConfigPath:   dhcp.DhcpConfigPath,
	MainConfName: dhcp.Dhcp6ConfigFile,
}

func TestStopDHCPv6(t *testing.T) {
	service := pb.StopDHCPv6Req{}
	err := handlerv6.StopDHCPv6(service)
	ut.Assert(t, err == nil, "dhcp6 stop successfully!")

	time.Sleep(2 * time.Second)
}

func TestStartDHCPv6(t *testing.T) {

	configFile := dhcp.DhcpConfigPath + dhcp.Dhcp6ConfigFile
	dhcpv6 := pb.StartDHCPv6Req{Config: configFile}
	err := handlerv6.StartDHCPv6(dhcpv6)
	ut.Assert(t, err == nil, "dhcp6 start successfully!")

	time.Sleep(2 * time.Second)
}

func TestDeleteSubnetv6(t *testing.T) {
	time.Sleep(time.Second)

	req := pb.DeleteSubnetv6Req{Subnet: "192.166.1.0/24"}
	err := handlerv6.DeleteSubnetv6(req)
	ut.Assert(t, err == nil, "delete Subnet 192.166.1.0/24 successfully!")
}

func TestCreateSubnetv6(t *testing.T) {

	time.Sleep(time.Second)
	d1 := &pb.Option{}

	d2 := &pb.Pools{}
	d2.Options = []*pb.Option{d1}
	d2.Pool = "192.166.1.10-192.166.1.40"

	req := pb.CreateSubnetv6Req{Subnet: "192.166.1.0/24",
		Pool: []*pb.Pools{d2},
	}

	err := handlerv6.CreateSubnetv6(req)
	ut.Assert(t, err == nil, "Create Subnet 192.166.1.0/24 successfully!")

}

func TestUpdateSubnetv6(t *testing.T) {
	time.Sleep(time.Second)

	d1 := &pb.Option{}
	d2 := &pb.Pools{}
	d2.Options = []*pb.Option{d1}
	d2.Pool = "192.166.1.10-192.166.1.33"

	req := pb.UpdateSubnetv6Req{Subnet: "192.166.1.0/24", Pool: []*pb.Pools{d2}}
	err := handlerv6.UpdateSubnetv6(req)
	ut.Assert(t, err == nil, "Update Subnet 192.166.1.0/24 successfully!")
}
