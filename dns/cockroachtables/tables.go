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
	Name         string
	Priority     int
	IsUsed       int
	ACLs         []DBACL       `gorm:"many2many:view_acls;"`
	Zones        []DBZone      `gorm:"foreignkey:ViewID"`
	Redirections []Redirection `gorm:"foreignkey:ViewID"`
}

type DBZone struct {
	gorm.Model
	Name        string
	ZoneFile    string
	ViewID      uint `sql:"type:integer REFERENCES db_views(id) on update cascade on delete cascade"`
	IsUsed      int
	IsForward   int
	ForwardType string
	RRs         []DBRR      `gorm:"foreignkey:ZoneID"`
	Forwarders  []Forwarder `gorm:"foreignkey:ZoneID"`
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

type Forwarder struct {
	gorm.Model
	IP     string
	ZoneID uint `sql:"type:integer REFERENCES db_zones(id) on update cascade on delete cascade"`
}

type DefaultForward struct {
	gorm.Model
	ForwardType string
	Forwarders  []DefaultForwarder `gorm:"foreignkey:ForwardID"`
}

type DefaultForwarder struct {
	gorm.Model
	IP        string
	ForwardID uint `sql:"type:integer REFERENCES default_forwards(id) on update cascade on delete cascade"`
}

type Redirection struct {
	gorm.Model
	Name         string
	TTL          uint
	DataType     string
	RedirectType string
	Value        string
	ViewID       uint `sql:"type:integer REFERENCES db_views(id) on update cascade on delete cascade"`
}
