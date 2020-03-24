package dhcprest

import (
	"testing"

	"log"

	"github.com/ben-han-cn/cement/unittest"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/utils"
)

var (
	dhcpv4        = &Dhcpv4{}
	subnetv4s     = &Subnetv4s{}
	rsvController = &ReservationHandler{}
)

func init() {
	//var db *gorm.DB
	db, err := gorm.Open("postgres", utils.DBAddr)
	if err != nil {
		panic(err)
	}

	PGDBConn = NewPGDB(db)
	defer PGDBConn.Close()
	dhcpv4 = NewDhcpv4(db)

	subnetv4s = NewSubnetv4s(db)

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

	subnetv4 := dhcpv4.getSubnetv4BySubnet("test01")

	var rsv RestReservation
	rsv.ID = subnetv4.GetID()
	rsv.Hostname = "hostname"
	rsv.BootFileName = "boot file 01"
	rsv.IpAddress = "1.1.1.2"

	_, err := PGDBConn.OrmCreateReservation(dhcpv4.db, subnetv4.ID, &rsv)

	unittest.Assert(t, err == nil, "Reservation create error")
}

func TestUpdateReservation(t *testing.T) {
	log.Print("---begin to update reservation")

	subnetv4 := dhcpv4.getSubnetv4BySubnet("test01")

	var err error
	rsvs := rsvController.GetReservations(subnetv4.ID)
	for _, v := range rsvs {

		if v.Hostname == "hostname" {
			var rsv RestReservation
			rsv.ID = v.ID
			rsv.Hostname = v.Hostname
			rsv.BootFileName = "boot file 05"
			rsv.IpAddress = "1.1.1.5"
			err = PGDBConn.OrmUpdateReservation(dhcpv4.db, subnetv4.ID, &rsv)
			log.Println("rsv.ID: ", rsv.ID, ", subnetId: ", subnetv4.ID)
		}
	}

	unittest.Assert(t, err == nil, "Reservation update error")
}

func TestDeleteReservation(t *testing.T) {
	log.Print("---begin to delete reservation")

	subnetv4 := dhcpv4.getSubnetv4BySubnet("test01")

	var err error
	rsvs := rsvController.GetReservations(subnetv4.ID)
	for _, v := range rsvs {
		if v.Hostname == "hostname" {
			err = PGDBConn.OrmDeleteReservation(dhcpv4.db, v.ID)
			log.Println("v.ID: ", v.ID, ", subnetId: ", subnetv4.ID)
		}
	}

	unittest.Assert(t, err == nil, "Reservation delete error")
}

func TestDeleteSubnetv4(t *testing.T) {
	log.Print("---begin to delete Subnetv4")
	var subnetv4 Subnetv4
	subnetv4.Subnet = "test01"
	subnetv4.ID = dhcpv4.getSubnetv4BySubnet(subnetv4.Subnet).ID
	err := dhcpv4.DeleteSubnetv4(&subnetv4)

	unittest.Assert(t, err == nil, "subnetv4 delete error")
}
