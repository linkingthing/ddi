package dhcprest

import (
	"log"
	"strconv"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"github.com/zdnscloud/gorest/resource"
)

type Dhcpv6 struct {
	db        *gorm.DB
	subnetv6s []*RestSubnetv6
	lock      sync.Mutex
}

type Subnetv6s struct {
	Subnetv6s []*RestSubnetv6
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
	Subnetv4Id            string         `json:"subnetv4Id"`
	OptionData            []RestOptionv6 `json:"optionData"`
	BeginAddress          string         `json:"beginAddress,omitempty" rest:"required=true,minLen=1,maxLen=12"`
	EndAddress            string         `json:"endAddress,omitempty" rest:"required=true,minLen=1,maxLen=12"`
	MaxValidLifetime      int            `json:"maxValidLifetime,omitempty"`
	ValidLifetime         int            `json:"validLifetime,omitempty"`
	Total                 uint32         `json:"total"`
	Usage                 float32        `json:"usage"`
	AddressType           string         `json:"addressType"`
	PoolName              string         `json:"poolName"`
}
type Poolv6Handler struct {
	subnetv6s *Subnetv6s
	db        *gorm.DB
	lock      sync.Mutex
}

func NewPoolv6Handler(s *Subnetv6s) *Poolv6Handler {
	return &Poolv6Handler{
		subnetv6s: s,
		db:        s.db,
	}
}

type RestSubnetv6 struct {
	resource.ResourceBase `json:"embedded,inline"`
	Name                  string `json:"name,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	SubnetId              string `json:"subnet_id"`
	ValidLifetime         string `json:"validLifetime"`
	Reservations          []*RestReservationv6
	Pools                 []*RestPoolv6
	SubnetTotal           string `json:"total"`
	SubnetUsage           string `json:"usage"`
}

//tools func
func ConvertReservationv6sFromOrmToRest(rs []*dhcporm.OrmReservationv6) []*RestReservationv6 {

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

func (s *Dhcpv6) ConvertSubnetv6FromOrmToRest(v *dhcporm.OrmSubnetv6) *RestSubnetv6 {

	v6 := &RestSubnetv6{}
	v6.SetID(strconv.Itoa(int(v.ID)))
	v6.Subnet = v.Subnet
	v6.ValidLifetime = v.ValidLifetime
	v6.Reservations = ConvertReservationv6sFromOrmToRest(v.Reservationv6s)
	return v6
}

//tools func
func ConvertPoolv6sFromOrmToRest(ps []*dhcporm.Poolv6) []*RestPoolv6 {
	log.Println("into ConvertPoolsFromOrmToRest")

	var restPs []*RestPoolv6
	for _, v := range ps {
		restP := RestPoolv6{
			BeginAddress: v.BeginAddress,
			EndAddress:   v.EndAddress,
		}
		restP.Total = ipv42Long(v.EndAddress) - ipv42Long(v.BeginAddress) + 1
		restP.ID = strconv.Itoa(int(v.ID))

		// todo get usage of a pool, (put it to somewhere)

		restP.Usage = 15.32
		restP.AddressType = "resv"
		restP.CreationTimestamp = resource.ISOTime(v.CreatedAt)
		restP.PoolName = v.BeginAddress + "-" + v.EndAddress

		restPs = append(restPs, &restP)

	}

	return restPs
}
func (r *Poolv6Handler) GetPoolv6s(subnetId string) []*RestPoolv6 {
	list := PGDBConn.OrmPoolv6List(subnetId)
	pool := ConvertPoolv6sFromOrmToRest(list)

	return pool
}
