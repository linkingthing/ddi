package gorest

import (
	"github.com/jinzhu/gorm"
)

type DBACL struct {
	gorm.Model
	Name string
	IPs  []DBIP `gorm:"foreignkey:ACLID"`
}

type DBIP struct {
	IP    string
	ACLID uint `sql:"type:integer REFERENCES dbacls(id) on update cascade on delete cascade"`
}

type DBView struct {
	gorm.Model
	Name     string
	Priority int
	ACLs     []DBACL  `gorm:"many2many:view_acls;"`
	Zones    []DBZone `gorm:"foreignkey:ViewID"`
}

type DBZone struct {
	gorm.Model
	Name     string
	ZoneFile string
	ViewID   uint   `sql:"type:integer REFERENCES db_views(id) on update cascade on delete cascade"`
	RRs      []DBRR `gorm:"foreignkey:ZoneID"`
}

type DBRR struct {
	gorm.Model
	Data   string
	ZoneID uint `sql:"type:integer REFERENCES db_zones(id) on update cascade on delete cascade"`
}
