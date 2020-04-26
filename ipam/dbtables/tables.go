package gorest

import (
	"github.com/jinzhu/gorm"
)

type AliveAddress struct {
	IPAddress     string `gorm:"primary_key"`
	LastAliveTime int64
	ScanTime      int64
	Subnetv4ID    uint `sql:"type:integer REFERENCES subnetv4s(id)"`
}

type Ipv6PlanedAddrTree struct {
	gorm.Model
	Depth         int
	Name          string
	ParentID      uint
	BeginSubnet   string
	EndSubnet     string
	BeginNodeCode int
	EndNodeCode   int
	MaxCode       int
	IsLeaf        bool
}

type BitsUseFor struct {
	Parentid uint `gorm:"primary_key"`
	UsedFor  string
}
