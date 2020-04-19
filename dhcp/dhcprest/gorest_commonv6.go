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
	Subnetv6Id            string         `json:"subnetv6Id"`
	OptionData            []RestOptionv6 `json:"optionData"`
	BeginAddress          string         `json:"beginAddress,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	EndAddress            string         `json:"endAddress,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	MaxValidLifetime      string         `json:"maxValidLifetime,omitempty"`
	ValidLifetime         string         `json:"validLifetime,omitempty"`
	Total                 uint32         `json:"total"`
	Usage                 float32        `json:"usage"`
	AddressType           string         `json:"addressType"`
	PoolName              string         `json:"poolName"`
	DnsServer             string         `json:"dnsServer"`
}

type Poolv6Handler struct {
	subnetv6s *Subnetv6s
	db        *gorm.DB
	lock      sync.Mutex
}

func (n RestPoolv6) GetParents() []resource.ResourceKind {
	log.Println("dhcprest, into RestPoolv6 GetParents")
	return []resource.ResourceKind{RestSubnetv6{}}
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
	MaxValidLifetime      string `json:"maxValidLifetime"`
	Reservations          []*RestReservationv6
	Pools                 []*RestPoolv6
	SubnetTotal           string `json:"total"`
	SubnetUsage           string `json:"usage"`
	//Gateway               string `json:"gateway"`
	DnsServer string `json:"dnsServer"`
	//added for new zone handler
	DhcpEnable int    `json:"dhcpEnable"`
	DnsEnable  int    `json:"dnsEnable"`
	ZoneName   string `json:"zoneName"`
	ViewId     string `json:"viewId"`
	Notes      string `json:"notes"`
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
	v6.SubnetId = strconv.Itoa(int(v.ID))
	v6.Subnet = v.Subnet
	v6.ValidLifetime = v.ValidLifetime
	v6.MaxValidLifetime = v.MaxValidLifetime
	v6.Reservations = ConvertReservationv6sFromOrmToRest(v.Reservationv6s)

	v6.DnsServer = v.DnsServer
	v6.DhcpEnable = v.DhcpEnable
	v6.DnsEnable = v.DnsEnable
	v6.ViewId = v.ViewId
	v6.Notes = v.Notes

	v6.CreationTimestamp = resource.ISOTime(v.CreatedAt)

	if len(v6.ZoneName) == 0 {
		v6.ZoneName = v.Name
	}
	return v6
}

//tools func
func (r *Poolv6Handler) ConvertPoolv6sFromOrmToRest(ps []*dhcporm.Poolv6) []*RestPoolv6 {
	log.Println("into ConvertPoolv6sFromOrmToRest")

	var restPs []*RestPoolv6
	for _, v := range ps {

		restP := r.convertSubnetv6PoolFromOrmToRest(v)

		restPs = append(restPs, restP)

	}

	return restPs
}
func (r *Poolv6Handler) GetPoolv6s(subnetId string) []*RestPoolv6 {
	list := PGDBConn.OrmPoolv6List(subnetId)
	pool := r.ConvertPoolv6sFromOrmToRest(list)

	return pool
}
func (r *Poolv6Handler) convertSubnetv6PoolFromOrmToRest(v *dhcporm.Poolv6) *RestPoolv6 {
	pool := &RestPoolv6{}

	if v == nil {
		return pool
	}

	pool.SetID(strconv.Itoa(int(v.ID)))
	pool.BeginAddress = v.BeginAddress
	pool.EndAddress = v.EndAddress
	pool.Total = ipv42Long(v.EndAddress) - ipv42Long(v.BeginAddress) + 1
	pool.ID = strconv.Itoa(int(v.ID))

	// todo get usage of a pool, (put it to somewhere)
	pool.Usage = 0
	pool.AddressType = "resv"
	pool.CreationTimestamp = resource.ISOTime(v.CreatedAt)
	pool.PoolName = v.BeginAddress + "-" + v.EndAddress

	//get ormSubnetv6 from subnetv6Id
	pgdb := NewPGDB(r.db)
	subnetv6Id := strconv.Itoa(int(v.Subnetv6ID))

	s6 := pgdb.GetSubnetv6ById(subnetv6Id)

	pool.DnsServer = s6.DnsServer
	pool.Subnetv6Id = subnetv6Id
	pool.MaxValidLifetime = strconv.Itoa(v.MaxValidLifetime)
	pool.ValidLifetime = strconv.Itoa(v.ValidLifetime)

	return pool
}
func (r *Poolv6Handler) GetSubnetv6Pool(subnetId string, pool_id string) *RestPoolv6 {
	orm := PGDBConn.OrmGetPoolv6(subnetId, pool_id)
	pool := r.convertSubnetv6PoolFromOrmToRest(orm)

	return pool
}
