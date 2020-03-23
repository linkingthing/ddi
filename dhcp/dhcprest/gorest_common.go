package dhcprest

import (
	"fmt"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"log"
	"strconv"
	"sync"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing",
		Version: "dhcp/v1",
	}

	subnetv4Kind    = resource.DefaultKindName(Subnetv4{})
	ReservationKind = resource.DefaultKindName(RestReservation{})
	PoolKind        = resource.DefaultKindName(RestPool{})
	OptionKind      = resource.DefaultKindName(RestOption{})

	db *gorm.DB
)

//type Dhcpv4Serv struct {
//	resource.ResourceBase `json:",inline"`
//	ConfigJson            string `json:"configJson" rest:"required=true,minLen=1,maxLen=1000000"`
//}

type RestOption struct {
	resource.ResourceBase `json:"embedded,inline"`
	AlwaysSend            bool   `gorm:"column:always-send"`
	Code                  uint64 `gorm:"column:code"`
	CsvFormat             bool   `json:"csv-format"`
	Data                  string `json:"data"`
	Name                  string `json:"name"`
	Space                 string `json:"space"`
}

type RestReservation struct {
	resource.ResourceBase `json:"embedded,inline"`
	BootFileName          string `json:"boot-file-name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid           string       `json:"duid"`
	Hostname       string       `json:"hostname"`
	IpAddress      string       `json:"ip-address"`
	NextServer     string       `json:"next-server"`
	OptionData     []RestOption `json:"option-data"`
	ServerHostname string       `json:"server-hostname"`
}

type RestPool struct {
	resource.ResourceBase `json:"embedded,inline"`
	OptionData            []RestOption `json:"option-data"`
	BeginAddress          string       `json:"begin address,omitempty" rest:"required=true,minLen=1,maxLen=12"`
	EndAddress            string       `json:"end address,omitempty" rest:"required=true,minLen=1,maxLen=12"`
}

type Subnetv4 struct {
	resource.ResourceBase `json:"embedded,inline"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	ValidLifetime         string `json:"validLifeTime"`
	Reservations          []*RestReservation
	Pools                 []*RestPool
}

type Subnetv4State struct {
	Subnetv4s []*Subnetv4
}
type Subnetv4Handler struct {
	subnetv4s *Subnetv4State
}

func NewSubnetv4State() *Subnetv4State {
	return &Subnetv4State{}
}

type PoolsState struct {
	Pools []*RestPool
}

func NewPoolsState() *PoolsState {
	return &PoolsState{}
}

type Dhcpv4 struct {
	db        *gorm.DB
	subnetv4s []*Subnetv4
	lock      sync.Mutex
}

type Subnetv4s struct {
	Subnetv4s []*Subnetv4
	db        *gorm.DB
}

func NewSubnetv4s(db *gorm.DB) *Subnetv4s {
	return &Subnetv4s{db: db}
}

type ReservationHandler struct {
	subnetv4s *Subnetv4s
	db        *gorm.DB
	lock      sync.Mutex
}

func NewReservationHandler(s *Subnetv4s) *ReservationHandler {
	return &ReservationHandler{
		subnetv4s: s,
		db:        s.db,
	}
}

type PoolHandler struct {
	subnetv4s *Subnetv4s
	db        *gorm.DB
	lock      sync.Mutex
}

func NewPoolHandler(s *Subnetv4s) *PoolHandler {
	return &PoolHandler{
		subnetv4s: s,
		db:        s.db,
	}
}

//tools func
func ConvertReservationsFromOrmToRest(rs []dhcporm.Reservation) []*RestReservation {

	var restRs []*RestReservation
	for _, v := range rs {
		restR := RestReservation{
			Duid:         v.Duid,
			BootFileName: v.BootFileName,
			Hostname:     v.Hostname,
		}
		restR.ID = strconv.Itoa(int(v.ID))
		restRs = append(restRs, &restR)
	}

	return restRs
}

//tools func
func ConvertPoolsFromOrmToRest(ps []*dhcporm.Pool) []*RestPool {
	log.Println("into ConvertPoolsFromOrmToRest")

	var restPs []*RestPool
	for _, v := range ps {
		restP := RestPool{
			BeginAddress: v.BeginAddress,
			EndAddress:   v.EndAddress,
		}
		restP.ID = strconv.Itoa(int(v.ID))
		restPs = append(restPs, &restP)
	}

	return restPs
}

func (s *Dhcpv4) convertSubnetv4FromOrmToRest(v *dhcporm.OrmSubnetv4) *Subnetv4 {

	v4 := &Subnetv4{}
	v4.SetID(strconv.Itoa(int(v.ID)))
	v4.Subnet = v.Subnet
	v4.ValidLifetime = v.ValidLifetime
	v4.Reservations = ConvertReservationsFromOrmToRest(v.Reservations)
	return v4
}
func (r *ReservationHandler) convertSubnetv4ReservationFromOrmToRest(v *dhcporm.Reservation) *RestReservation {
	rsv := &RestReservation{}

	if v == nil {
		return rsv
	}

	rsv.SetID(strconv.Itoa(int(v.ID)))
	rsv.BootFileName = v.BootFileName
	rsv.Duid = v.Duid

	return rsv
}

func (n RestReservation) GetParents() []resource.ResourceKind {
	log.Println("dhcprest, into GetParents")
	return []resource.ResourceKind{Subnetv4{}}
}

//func (s *Dhcpv4) convertSubnetv4ReservationFromOrmToRest(v *dhcporm.Reservation) *RestReservation {
//
//	rr := &RestReservation{}
//	rr.SetID(strconv.Itoa(int(v.ID)))
//	rr.BootFileName = v.BootFileName
//	rr.Duid = v.Duid
//
//	return rr
//}
func ConvertStringToUint(s string) uint {
	dbId, err := strconv.Atoi(s)
	if err != nil {
		fmt.Errorf("convert string to uint error, s: %s", s)
		return 0
	}

	return uint(dbId)
}

func (r *ReservationHandler) GetReservations(subnetId string) []*RestReservation {
	list := PGDBConn.OrmReservationList(subnetId)
	rsv := ConvertReservationsFromOrmToRest(list)

	return rsv
}
func (r *ReservationHandler) GetSubnetv4Reservation(subnetId string, rsv_id string) *RestReservation {
	orm := PGDBConn.OrmGetReservation(subnetId, rsv_id)
	rsv := r.convertSubnetv4ReservationFromOrmToRest(orm)

	return rsv
}
