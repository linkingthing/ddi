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
	log.Print("begin to list subnet v4")

	v4s := dhcpv4.GetSubnetv4s()
	len := len(v4s)

	unittest.Assert(t, len == 0, "length of subnetv4 list is 0")
}
func TestSubnetv4Create(t *testing.T) {
	log.Print("begin to create subnet v4")

	var subnetv4 *Subnetv4
	subnetv4.Subnet = "test01"
	subnetv4.ValidLifetime = "3000"
	var rsv = RestReservation{
		BootFileName: "bootfile1",
		Hostname:     "hostname1",
	}
	var rsvs []*RestReservation
	rsvs = append(rsvs, &rsv)

	subnetv4.Reservations = rsvs

	err := dhcpv4.AddSubnetv4(subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 create error")
}
