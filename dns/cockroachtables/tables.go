package gorest

import (
	"github.com/jinzhu/gorm"
)

type ACL struct {
	gorm.Model
	Name  string
	Views []View `gorm:"many2many:view_acls;"`
	IPs   []IP   `gorm:"foreignkey:ACLID"`
	//SortLists []SortList `gorm:"many2many:sortlist_acls;"`
}

type IP struct {
	IP    string
	ACLID uint `sql:"type:integer REFERENCES acls(id) on update cascade on delete cascade"`
}

type View struct {
	gorm.Model
	Name         string
	Priority     int
	ACLs         []ACL         `gorm:"many2many:view_acls;"`
	Zones        []Zone        `gorm:"foreignkey:ViewID"`
	Redirections []Redirection `gorm:"foreignkey:ViewID"`
	DNS64s       []DNS64       `gorm:"foreignkey:ViewID"`
}

type Zone struct {
	gorm.Model
	Name        string
	ZoneFile    string
	ViewID      uint `sql:"type:integer REFERENCES views(id) on update cascade on delete cascade"`
	ZoneType    string
	IsForward   int
	ForwardType string
	RRs         []RR        `gorm:"foreignkey:ZoneID"`
	Forwarders  []Forwarder `gorm:"foreignkey:ZoneID"`
}

type RR struct {
	gorm.Model
	Name     string
	DataType string
	TTL      uint
	Value    string
	ZoneID   uint `sql:"type:integer REFERENCES zones(id) on update cascade on delete cascade"`
}

type Forwarder struct {
	gorm.Model
	IP     string
	ZoneID uint `sql:"type:integer REFERENCES zones(id) on update cascade on delete cascade"`
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
	ViewID       uint `sql:"type:integer REFERENCES views(id) on update cascade on delete cascade"`
}

type DNS64 struct {
	gorm.Model
	Prefix    string
	ClientACL uint `sql:"type:integer REFERENCES acls(id)"`
	AAddress  uint `sql:"type:integer REFERENCES acls(id)"`
	ViewID    uint `sql:"type:integer REFERENCES views(id) on update cascade on delete cascade"`
}

type DefaultDNS64 struct {
	gorm.Model
	Prefix    string
	ClientACL uint `sql:"type:integer REFERENCES acls(id)"`
	AAddress  uint `sql:"type:integer REFERENCES acls(id)"`
}

type IPBlackHole struct {
	gorm.Model
	ACLID uint `sql:"type:integer REFERENCES acls(id)"`
}

type RecursiveConcurrent struct {
	gorm.Model
	RecursiveClients uint
	FetchesPerZone   uint
}

type DDIUserPWD struct {
	gorm.Model
	Name string
	PWD  string
}
type SortListElement struct {
	gorm.Model
	ACLID     uint `sql:"type:integer REFERENCES acls(id)"`
	NextACLID string
	PrevACLID string
}
