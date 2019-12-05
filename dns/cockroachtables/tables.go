package gorest

import (
	"github.com/jinzhu/gorm"
)

type DBACL struct {
	gorm.Model
	Name   string
	ViewID string
}

type DBIP struct {
	IP    string
	ACLID string
}

type DBView struct {
	gorm.Model
	Name     string
	Priority int
}

type DBZone struct {
	gorm.Model
	Name     string
	ZoneFile string
	ViewID   string
}

type DBRR struct {
	gorm.Model
	Data   string
	ZoneID string
}
