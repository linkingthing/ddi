package dhcprest

import (
	"testing"

	"github.com/ben-han-cn/cement/unittest"
	"log"
)

var (
	dhcpv4        = &Dhcpv4{}
	subnetv4s     = &Subnetv4s{}
	rsvController = &ReservationHandler{}
)

func init() {
	PGDBConn = NewPGDB()
	defer PGDBConn.Close()
	dhcpv4 = NewDhcpv4(NewPGDB().DB)

	subnetv4s = NewSubnetv4s(NewPGDB().DB)

	rsvController = NewReservationHandler(subnetv4s)
}

func TestSubnetv4List(t *testing.T) {
	log.Print("---begin to list subnet v4")

	v4s := dhcpv4.GetSubnetv4s()
	len := len(v4s)

	unittest.Assert(t, len > 0, "length of subnetv4 list is 0")
}
func TestCreateSubnetv4(t *testing.T) {
	log.Print("---begin to create subnet v4")

	var subnetv4 Subnetv4
	subnetv4.Subnet = "test01"
	subnetv4.ValidLifetime = "3000"

	err := dhcpv4.CreateSubnetv4(&subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 create error")
}

func TestUpdateSubnetv4(t *testing.T) {
	log.Print("---begin to update subnet v4")

	var subnetv4 Subnetv4
	subnetv4.Subnet = "test01"
	subnetv4.ValidLifetime = "3001"

	err := dhcpv4.UpdateSubnetv4(&subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 update error")
}

func TestCreateReservation(t *testing.T) {
	log.Print("---begin to create reservation")

	subnetv4 := dhcpv4.getSubnetv4ByName("test01")

	var rsv RestReservation
	rsv.ID = subnetv4.GetID()
	rsv.Hostname = "hostname"
	rsv.BootFileName = "boot file 01"
	rsv.IpAddress = "1.1.1.2"

	_, err := PGDBConn.OrmCreateReservation(dhcpv4.db, subnetv4.ID, &rsv)

	unittest.Assert(t, err == nil, "Reservation  create error")
}

func TestUpdateReservation(t *testing.T) {
	log.Print("---begin to update reservation")

	subnetv4 := dhcpv4.getSubnetv4ByName("test01")

	var rsv RestReservation
	rsv.ID = subnetv4.GetID()
	rsv.Hostname = "hostname"
	rsv.BootFileName = "boot file 02"
	rsv.IpAddress = "1.1.1.3"

	err := PGDBConn.OrmUpdateReservation(dhcpv4.db, subnetv4.ID, &rsv)

	unittest.Assert(t, err == nil, "Reservation  create error")
}

func TestDeleteSubnetv4(t *testing.T) {
	log.Print("---begin to delete subnet v4")

	var subnetv4 Subnetv4
	subnetv4.Subnet = "test01"
	//subnetv4.ValidLifetime = "3001"
	subnetv4.ID = dhcpv4.getSubnetv4ByName(subnetv4.Subnet).ID

	err := dhcpv4.DeleteSubnetv4(&subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 delete error")
}
