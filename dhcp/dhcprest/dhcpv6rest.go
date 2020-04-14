package dhcprest

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
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

	v6 := s.ConvertSubnetv6FromOrmToRest(v)
	return v6
}

func (s *Dhcpv6) getSubnetv6ByName(name string) *Subnetv6 {
	log.Println("In dhcprest getSubnetv6ByName, name: ", name)

	v := PGDBConn.GetSubnetv6ByName(s.db, name)
	if v.ID == 0 {
		return nil
	}
	v6 := s.ConvertSubnetv6FromOrmToRest(v)

	return v6
}

func (s *Dhcpv6) GetSubnetv6s() []*Subnetv6 {
	s.lock.Lock()
	defer s.lock.Unlock()

	list := PGDBConn.Subnetv6List()

	var v6 []*Subnetv6
	for _, v := range list {
		var subnet *Subnetv6
		subnet = s.ConvertSubnetv6FromOrmToRest(&v)
		v6 = append(v6, subnet)
	}
	return v6
}
func (s *Dhcpv6) getSubnetv6BySubnet(subnet string) *Subnetv6 {
	log.Println("In dhcprest getSubnetv4BySubnet, subnet: ", subnet)

	v := PGDBConn.getSubnetv6BySubnet(subnet)
	if v.ID == 0 {
		return nil
	}
	v4 := s.ConvertSubnetv6FromOrmToRest(v)

	return v4
}
func (s *Dhcpv6) CreateSubnetv6(subnetv6 *Subnetv6) error {
	log.Println("into CreateSubnetv6, subnetv6: ", subnetv6)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv6BySubnet(subnetv6.Subnet); c != nil {
		errStr := "subnet " + subnetv6.Subnet + " already exist"
		return fmt.Errorf(errStr)
	}

	log.Println("in dhcp/dhcprest CreateSubnetv4, subnetv4: ", subnetv6)
	s6, err := PGDBConn.CreateSubnetv6(subnetv6.Name, subnetv6.Subnet, subnetv6.ValidLifetime)
	if err != nil {
		return err
	}
	if s6.Subnet == "" {
		return fmt.Errorf("添加子网失败")
	}

	// set newly inserted id
	subnetv6.ID = strconv.Itoa(int(s6.ID))
	subnetv6.SubnetId = strconv.Itoa(int(s6.ID))
	subnetv6.SetCreationTimestamp(s6.CreatedAt)
	log.Println("newly inserted id: ", s6.ID)

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

func (h *subnetv6Handler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into subnetv6Handler dhcprest.go List")

	return h.subnetv6s.GetSubnetv6s(), nil
}

func (h *subnetv6Handler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {

	return h.subnetv6s.GetSubnetv6(ctx.Resource.GetID()), nil
}
func (r *Poolv6Handler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dhcprest.go subnetv4PoolHandler List")
	pool := ctx.Resource.(*RestPoolv6)
	return r.GetPoolv6s(pool.GetParent().GetID()), nil
}

//func (r *Poolv6Handler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
//	log.Println("into dhcprest.go PoolHandler Get")
//	pool := ctx.Resource.(*RestPool)
//	return r.GetSubnetv4Pool(pool.GetParent().GetID(), pool.GetID()), nil
//}
