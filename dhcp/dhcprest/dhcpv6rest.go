package dhcprest

import (
	"fmt"
	goresterr "github.com/ben-han-cn/gorest/error"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	"log"
	"time"
)

var (
	subnetv6Kind = resource.DefaultKindName(Subnetv6{})
)

func NewDhcpv6(db *gorm.DB) *Dhcpv6 {
	return &Dhcpv6{db: db}
}

type subnetv6Handler struct {
	subnetv6s *Dhcpv6
}

func NewSubnetv6Handler(s *Dhcpv6) *subnetv6Handler {
	return &subnetv6Handler{
		subnetv6s: s,
	}
}

func (s *Dhcpv6) GetSubnetv6(id string) *Subnetv6 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getSubnetv6(id)
}

func (s *Dhcpv6) getSubnetv6(id string) *Subnetv6 {
	v := PGDBConn.GetSubnetv6(s.db, id)
	if v.ID == 0 {
		return nil
	}

	v6 := s.convertSubnetv6FromOrmToRest(v)
	return v6
}

func (s *Dhcpv6) getSubnetv6ByName(name string) *Subnetv6 {
	log.Println("In dhcprest getSubnetv6ByName, name: ", name)

	v := PGDBConn.GetSubnetv6ByName(s.db, name)
	if v.ID == 0 {
		return nil
	}
	v6 := s.convertSubnetv6FromOrmToRest(v)

	return v6
}

func (s *Dhcpv6) GetSubnetv6s() []*Subnetv6 {
	s.lock.Lock()
	defer s.lock.Unlock()

	list := PGDBConn.Subnetv6List(s.db)

	var v6 []*Subnetv6
	for _, v := range list {
		var subnet *Subnetv6
		subnet = s.convertSubnetv6FromOrmToRest(&v)
		v6 = append(v6, subnet)
	}
	return v6
}

func (s *Dhcpv6) CreateSubnetv6(subnetv6 *Subnetv6) error {
	fmt.Println("into CreateSubnetv6")

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv6ByName(subnetv6.Subnet); c != nil {
		errStr := "subnet " + subnetv6.Subnet + " already exist"
		return fmt.Errorf(errStr)
	}

	err := PGDBConn.CreateSubnetv6(s.db, subnetv6.Subnet, subnetv6.ValidLifetime)
	if err != nil {
		return err
	}

	return nil
}

func (h *subnetv6Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go v6 Create")

	subnetv6 := ctx.Resource.(*Subnetv6)
	subnetv6.SetID(subnetv6.Subnet)
	subnetv6.SetCreationTimestamp(time.Now())
	if err := h.subnetv6s.CreateSubnetv6(subnetv6); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	} else {
		return subnetv6, nil
	}
}

func (h *subnetv6Handler) List(ctx *resource.Context) interface{} {
	log.Println("into dhcprest.go List")

	return h.subnetv6s.GetSubnetv6s()
}

func (h *subnetv6Handler) Get(ctx *resource.Context) resource.Resource {

	return h.subnetv6s.GetSubnetv6(ctx.Resource.GetID())
}
