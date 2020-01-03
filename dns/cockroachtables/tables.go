package gorest

import (
	"github.com/jinzhu/gorm"
)

type DBACL struct {
	gorm.Model
	Name   string
	IsUsed int
	IPs    []DBIP `gorm:"foreignkey:ACLID"`
}

type DBIP struct {
	IP    string
	ACLID uint `sql:"type:integer REFERENCES dbacls(id) on update cascade on delete cascade"`
}

type DBView struct {
	gorm.Model
	Name     string
	Priority int
	IsUsed   int
	ACLs     []DBACL  `gorm:"many2many:view_acls;"`
	Zones    []DBZone `gorm:"foreignkey:ViewID"`
}

type DBZone struct {
	gorm.Model
	Name     string
	ZoneFile string
	ViewID   uint `sql:"type:integer REFERENCES db_views(id) on update cascade on delete cascade"`
	IsUsed   int
	RRs      []DBRR `gorm:"foreignkey:ZoneID"`
}

type DBRR struct {
	gorm.Model
	Name     string
	DataType string
	TTL      uint
	Value    string
	IsUsed   int
	ZoneID   uint `sql:"type:integer REFERENCES db_zones(id) on update cascade on delete cascade"`
}
