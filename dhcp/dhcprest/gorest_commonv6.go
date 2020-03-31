package dhcprest

import (
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"strconv"
	"sync"
)

type Dhcpv6 struct {
	db        *gorm.DB
	subnetv6s []*Subnetv6
	lock      sync.Mutex
}

type Subnetv6s struct {
	Subnetv6s []*Subnetv6
	db        *gorm.DB
}

func NewSubnetv6s(db *gorm.DB) *Subnetv6s {
	return &Subnetv6s{db: db}
}

type RestOptionv6 struct {
	resource.ResourceBase `json:"embedded,inline"`
	AlwaysSend            bool   `gorm:"column:always-send"`
	Code                  uint64 `gorm:"column:code"`
	CsvFormat             bool   `json:"csv-format"`
	Data                  string `json:"data"`
	Name                  string `json:"name"`
	Space                 string `json:"space"`
}

type RestReservationv6 struct {
	resource.ResourceBase `json:"embedded,inline"`
	BootFileName          string        `json:"boot-file-name"`
	Duid                  string        `json:"duid"`
	Hostname              string        `json:"hostname"`
	IpAddress             string        `json:"ip-address"`
	NextServer            string        `json:"next-server"`
	OptionData            []*RestOption `json:"option-data"`
	ServerHostname        string        `json:"server-hostname"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
}

type RestPoolv6 struct {
	resource.ResourceBase `json:"embedded,inline"`
	OptionData            []*RestOptionv6 `json:"option-data"`
	Pool                  string          `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
}

type Subnetv6 struct {
	resource.ResourceBase `json:"embedded,inline"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	ValidLifetime         string `json:"validLifetime"`
	Reservations          []*RestReservationv6
	Pools                 []*RestPoolv6
}

//tools func
func ConvertReservationv6sFromOrmToRest(rs []*dhcporm.Reservationv6) []*RestReservationv6 {

	var restRs []*RestReservationv6
	for _, v := range rs {
		restR := RestReservationv6{
			Duid:         v.Duid,
			BootFileName: v.BootFileName,
			Hostname:     v.Hostname,
		}
		restR.ID = strconv.Itoa(int(v.ID))
		restRs = append(restRs, &restR)
	}

	return restRs
}

func (s *Dhcpv6) convertSubnetv6FromOrmToRest(v *dhcporm.OrmSubnetv6) *Subnetv6 {

	v6 := &Subnetv6{}
	v6.SetID(strconv.Itoa(int(v.ID)))
	v6.Subnet = v.Subnet
	v6.ValidLifetime = v.ValidLifetime
	v6.Reservations = ConvertReservationv6sFromOrmToRest(v.Reservations)
	return v6
}
