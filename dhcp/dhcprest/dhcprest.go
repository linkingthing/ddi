package main

import (
	"sync"

	"fmt"
	"time"

	"log"

	"github.com/ben-han-cn/gorest"
	"github.com/ben-han-cn/gorest/adaptor"
	goresterr "github.com/ben-han-cn/gorest/error"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/ben-han-cn/gorest/resource/schema"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
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

func newDhcpv4(db *gorm.DB) *Dhcpv4 {
	return &Dhcpv4{db: db}
}

func (s *Dhcpv4) AddSubnetv4(subnetv4 *Subnetv4) error {
	fmt.Println("into AddSubnetv4")
	fmt.Print(subnetv4)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv4(subnetv4.Subnet); c != nil {
		return fmt.Errorf("subnet %s already exist", subnetv4.Subnet)
	}

	err := dhcporm.CreateSubnetv4(s.db, subnetv4.Subnet, subnetv4.ValidLifetime)
	if err != nil {
		return err
	}

	return nil
}

func (s *Dhcpv4) GetSubnetv4(subnet string) *Subnetv4 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getSubnetv4(subnet)
}
func (s *Dhcpv4) getSubnetv4(subnet string) *Subnetv4 {
	v := dhcporm.GetSubnetv4(s.db, subnet)
	if len(v) == 0 {
		return nil
	}

	v4 := &Subnetv4{}
	v4.Subnet = v[0].Subnet
	v4.ValidLifetime = v[0].ValidLifetime

	return v4
}
func (s *Dhcpv4) GetSubnetv4s() []*Subnetv4 {
	s.lock.Lock()
	defer s.lock.Unlock()

	list := dhcporm.Subnetv4List(s.db)

	var v4 []*Subnetv4
	for _, v := range list {

		var subnet Subnetv4
		subnet.Subnet = v.Subnet
		subnet.ValidLifetime = v.ValidLifetime
		subnet.ID = string(v.ID)

		v4 = append(v4, &subnet)
	}
	return v4
}

type subnetv4Handler struct {
	subnetv4s *Dhcpv4
}

func newSubnetv4Handler(s *Dhcpv4) *subnetv4Handler {
	return &subnetv4Handler{
		subnetv4s: s,
	}
}
func (h *subnetv4Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {

	subnetv4 := ctx.Resource.(*Subnetv4)
	subnetv4.SetID(subnetv4.Subnet)
	subnetv4.SetCreationTimestamp(time.Now())
	if err := h.subnetv4s.AddSubnetv4(subnetv4); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	} else {
		return subnetv4, nil
	}
}

func (h *subnetv4Handler) List(ctx *resource.Context) interface{} {

	return h.subnetv4s.GetSubnetv4s()
}

func (h *subnetv4Handler) Get(ctx *resource.Context) resource.Resource {

	return h.subnetv4s.GetSubnetv4(ctx.Resource.GetID())
}

func main() {
	db, err := gorm.Open("postgres", dhcporm.CRDBAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	schemas := schema.NewSchemaManager()
	dhcpv4 := newDhcpv4(db)

	schemas.Import(&version, Subnetv4{}, newSubnetv4Handler(dhcpv4))

	//dhcpv6 := newDhcpv6()
	//schemas.Import(&version, Subnetv6{}, newSubnetv4Handler(dhcpv6))

	router := gin.Default()
	adaptor.RegisterHandler(router, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
	router.Run("0.0.0.0:1234")
}
