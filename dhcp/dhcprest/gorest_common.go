package dhcprest

import (
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"strconv"
	"sync"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing",
		Version: "dhcp/v1",
	}
	subnetv4Kind = resource.DefaultKindName(Subnetv4{})
)

//type Dhcpv4Serv struct {
//	resource.ResourceBase `json:",inline"`
//	ConfigJson            string `json:"configJson" rest:"required=true,minLen=1,maxLen=1000000"`
//}

type Option struct {
	AlwaysSend bool   `gorm:"column:always-send"`
	Code       uint64 `gorm:"column:code"`
	CsvFormat  bool   `json:"csv-format"`
	Data       string `json:"data"`
	Name       string `json:"name"`
	Space      string `json:"space"`
}

type Reservations struct {
	BootFileName string `json:"boot-file-name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid           string   `json:"duid"`
	Hostname       string   `json:"hostname"`
	IpAddress      string   `json:"ip-address"`
	NextServer     string   `json:"next-server"`
	OptionData     []Option `json:"option-data"`
	ServerHostname string   `json:"server-hostname"`
}

type Pool struct {
	OptionData []Option `json:"option-data"`
	Pool       string   `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
}
type Subnetv4 struct {
	resource.ResourceBase `json:"embedded,inline"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	ValidLifetime         string `json:"validLifeTime"`
	Reservations          []Reservations
	Pools                 []Pool
}

type Dhcpv4 struct {
	db        *gorm.DB
	subnetv4s []*Subnetv4
	lock      sync.Mutex
}

//tools func
func (s *Dhcpv4) convertReservationsFromOrmToRest(rs []dhcporm.Reservation) []Reservations {

	var restRs []Reservations
	for _, v := range rs {
		restR := Reservations{
			Duid:         v.Duid,
			BootFileName: v.BootFileName,
		}
		restRs = append(restRs, restR)
	}

	return restRs
}

func (s *Dhcpv4) convertSubnetv4FromOrmToRest(v *dhcporm.Subnetv4) *Subnetv4 {

	v4 := &Subnetv4{}
	v4.SetID(strconv.Itoa(int(v.ID)))
	v4.Subnet = v.Subnet
	v4.ValidLifetime = v.ValidLifetime
	v4.Reservations = s.convertReservationsFromOrmToRest(v.Reservations)
	return v4
}
