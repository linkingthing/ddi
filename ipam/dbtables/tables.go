package gorest

import (
	"github.com/jinzhu/gorm"
)

type DividedAddress struct {
	gorm.Model
	IP               string
	AddressType      string
	HostName         string
	MacAddress       string
	MacVender        string
	OperSystem       string
	NetBIOSName      string
	InterfaceID      string
	FingerPrint      string
	LeaseStartTime   int64
	LeaseEndTime     int64
	DeviceTypeFlag   bool
	DeviceType       string
	BusinessFlag     bool
	Business         string
	ChargePersonFlag bool
	ChargePerson     string
	TelFlag          bool
	Tel              string
	DepartmentFlag   bool
	Department       string
	PositionFlag     bool
	Position         string
}

/*type IPAttrAppend struct {
	IP           string `gorm:"primary_key"`
	DeviceType   bool
	Business     bool
	ChargePerson bool
	Tel          bool
	Department   bool
	Position     bool
}*/

type AliveAddress struct {
	IPAddress     string `gorm:"primary_key"`
	LastAliveTime int64
	ScanTime      int64
	Subnetv4ID    uint `sql:"type:integer REFERENCES subnetv4s(id)"`
}
