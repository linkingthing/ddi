package dhcprest

import (
	"testing"

	"github.com/ben-han-cn/cement/unittest"
	"log"
)

var (
	dhcpv4 = &Dhcpv4{}
)

func init() {
	PGDBConn = NewPGDB()
	defer PGDBConn.Close()
	dhcpv4 = NewDhcpv4(NewPGDB().DB)
}

func TestSubnetv4List(t *testing.T) {
	log.Print("---begin to list subnet v4")

	v4s := dhcpv4.GetSubnetv4s()
	len := len(v4s)
	log.Println("len: ", len)

	unittest.Assert(t, len > 0, "length of subnetv4 list is 0")
}
func TestSubnetv4Create(t *testing.T) {
	log.Print("---begin to create subnet v4")

	var subnetv4 Subnetv4
	subnetv4.Subnet = "test01"
	subnetv4.ValidLifetime = "3000"

	err := dhcpv4.AddSubnetv4(&subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 create error")
}

func TestSubnetv4Update(t *testing.T) {
	log.Print("---begin to update subnet v4")

	s := dhcpv4.getSubnetv4ByName("test01")
	log.Print(s)

	var subnetv4 Subnetv4
	subnetv4.Subnet = "test01"
	subnetv4.ValidLifetime = "3001"

	err := dhcpv4.UpdateSubnetv4(&subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 update error")
}

func TestSubnetv4Delete(t *testing.T) {
	log.Print("---begin to delete subnet v4")

	var subnetv4 Subnetv4
	subnetv4.Subnet = "test01"
	//subnetv4.ValidLifetime = "3001"
	subnetv4.ID = dhcpv4.getSubnetv4ByName(subnetv4.Subnet).ID

	err := dhcpv4.DeleteSubnetv4(&subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 delete error")
}
