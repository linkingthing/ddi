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
	subnetv6Kind = resource.DefaultKindName(RestSubnetv6{})
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

func (s *Dhcpv6) GetSubnetv6(id string) *RestSubnetv6 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getSubnetv6(id)
}

func (s *Dhcpv6) getSubnetv6(id string) *RestSubnetv6 {
	v := PGDBConn.GetSubnetv6(s.db, id)
	if v.ID == 0 {
		return nil
	}

	v6 := s.ConvertSubnetv6FromOrmToRest(v)
	return v6
}

func (s *Dhcpv6) getSubnetv6ByName(name string) *RestSubnetv6 {
	log.Println("In dhcprest getSubnetv6ByName, name: ", name)

	v := PGDBConn.GetSubnetv6ByName(s.db, name)
	if v.ID == 0 {
		return nil
	}
	v6 := s.ConvertSubnetv6FromOrmToRest(v)

	return v6
}

func (s *Dhcpv6) GetSubnetv6s(search *SubnetSearch) []*RestSubnetv6 {
	s.lock.Lock()
	defer s.lock.Unlock()

	list := PGDBConn.Subnetv6List(search)

	var v6 []*RestSubnetv6
	for _, v := range list {
		var subnet *RestSubnetv6
		subnet = s.ConvertSubnetv6FromOrmToRest(&v)
		v6 = append(v6, subnet)
	}
	return v6
}

func (s *Dhcpv6) CreateSubnetv6(subnetv6 *RestSubnetv6) error {
	log.Println("into CreateSubnetv6, subnetv6: ", subnetv6)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv6BySubnet(subnetv6.Subnet); c != nil {
		errStr := "subnet " + subnetv6.Subnet + " already exist"
		return fmt.Errorf(errStr)
	}
	subnetv6.DhcpEnable = 1
	s6, err := PGDBConn.CreateSubnetv6(subnetv6)
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
	log.Println("in CreateSubnetv6, subnetv6.Name: ", subnetv6.Name)
	log.Println("in CreateSubnetv6, subnetv6.ZoneName: ", subnetv6.ZoneName)
	log.Println("in CreateSubnetv6, subnetv6.dhcpenable: ", subnetv6.DhcpEnable)
	subnetv6.ZoneName = s6.Name

	log.Println("newly inserted id: ", s6.ID)

	return nil
}

func (s *Dhcpv6) UpdateSubnetv6(subnetv6 *RestSubnetv6) error {

	log.Println("in UpdateSubnetv6(), subnetv6 subnet: ", subnetv6.Subnet)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv6ById(subnetv6.ID); c == nil {
		return fmt.Errorf("subnet %s not exist", subnetv6.ID)
	}

	err := PGDBConn.OrmUpdateSubnetv6(subnetv6)
	if err != nil {
		return err
	}
	log.Println("in UpdateSubnetv6, subnetv6.Name: ", subnetv6.Name)
	log.Println("in UpdateSubnetv6, subnetv6.ZoneName: ", subnetv6.ZoneName)
	if len(subnetv6.ZoneName) == 0 {
		subnetv6.ZoneName = subnetv6.Name
	}
	subnetv6.CreationTimestamp = resource.ISOTime(subnetv6.GetCreationTimestamp())
	log.Println("subnetv6.CreationTimestamp ", subnetv6.CreationTimestamp)

	return nil
}

func (s *Dhcpv6) DeleteSubnetv6(s6 *RestSubnetv6) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Println("dhcp/dhcprest DeleteSubnetv4() subnetv4 id: ", s6.ID)
	if c := s.getSubnetv6ById(s6.ID); c == nil {
		return fmt.Errorf("subnet %s not exist", s6.Subnet)
	}

	err := PGDBConn.DeleteSubnetv6(s6.ID)
	if err != nil {
		return err
	}

	return nil
}

func (h *subnetv6Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go v6 Create")

	subnetv6 := ctx.Resource.(*RestSubnetv6)
	subnetv6.SetID(subnetv6.Subnet)
	subnetv6.SetCreationTimestamp(time.Now())
	if err := h.subnetv6s.CreateSubnetv6(subnetv6); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	} else {
		return subnetv6, nil
	}
}
func (h *subnetv6Handler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Update")

	subnetv6 := ctx.Resource.(*RestSubnetv6)
	if err := h.subnetv6s.UpdateSubnetv6(subnetv6); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	if subnetv6.SubnetId == "" {
		subnetv6.SubnetId = subnetv6.ID
	}

	return subnetv6, nil
}

func (h *subnetv6Handler) Delete(ctx *resource.Context) *goresterr.APIError {
	log.Println("into dhcprest.go Delete")
	subnetv6 := ctx.Resource.(*RestSubnetv6)

	if err := h.subnetv6s.DeleteSubnetv6(subnetv6); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil

}

func (h *subnetv6Handler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into subnetv6Handler dhcprest.go List")
	var search *SubnetSearch
	// no search now
	return h.subnetv6s.GetSubnetv6s(search), nil
}

func (h *subnetv6Handler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {

	return h.subnetv6s.GetSubnetv6(ctx.Resource.GetID()), nil
}

func (r *Poolv6Handler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dhcprest.go subnetv6PoolHandler List")
	pool := ctx.Resource.(*RestPoolv6)
	return r.GetPoolv6s(pool.GetParent().GetID()), nil
}

func (s *Dhcpv6) getSubnetv6ById(id string) *RestSubnetv6 {

	v := PGDBConn.GetSubnetv6ById(id)
	if v.ID == 0 {
		return nil
	}

	v4 := s.ConvertSubnetv6FromOrmToRest(v)
	return v4
}
func (s *Dhcpv6) getSubnetv6BySubnet(subnet string) *RestSubnetv6 {
	log.Println("In dhcprest getSubnetv6BySubnet, subnet: ", subnet)

	v := PGDBConn.getOrmSubnetv6BySubnet(subnet)
	if v.ID == 0 {
		return nil
	}
	v6 := s.ConvertSubnetv6FromOrmToRest(&v)

	return v6
}
func (r *Poolv6Handler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Poolv6Handler Get")
	pool := ctx.Resource.(*RestPool)
	return r.GetSubnetv6Pool(pool.GetParent().GetID(), pool.GetID()), nil
}

func (r *Poolv6Handler) UpdateSubnetv6Server(subnetId string, pool *RestPoolv6) error {
	ormSubnetv6 := PGDBConn.GetSubnetv6ById(subnetId)
	//log.Println("into UpdateSubnetv4Server, pool.DnsServer: ", pool.DnsServer)
	//log.Println("into UpdateSubnetv4Server, pool.Gateway: ", pool.Gateway)
	ormSubnetv6.DnsServer = pool.DnsServer
	ormSubnetv6.ValidLifetime = pool.ValidLifetime
	ormSubnetv6.MaxValidLifetime = pool.MaxValidLifetime
	var s Dhcpv6
	restSubnetv6 := s.ConvertSubnetv6FromOrmToRest(ormSubnetv6)
	//restSubnetv4.Gateway = pool.Gateway
	//restSubnetv4.DnsServer = pool.DnsServer
	if err := s.UpdateSubnetv6(restSubnetv6); err != nil {
		log.Println("in UpdatePoolv6, update subnetv6 dnsServer error: ", err)
		return err
	}
	return nil
}
func (r *Poolv6Handler) CreatePoolv6(pool *RestPoolv6) (*RestPoolv6, error) {
	log.Println("into dhcprest/CreatePoolv6, pool: ", pool)

	r.lock.Lock()
	defer r.lock.Unlock()

	//todo check whether it exists
	subnetv6ID := pool.GetParent().GetID()

	log.Println("before CreatePool, subnetv6ID:", subnetv6ID)

	if err := r.UpdateSubnetv6Server(subnetv6ID, pool); err != nil {
		return nil, err
	}

	log.Println("before OrmCreatePoolv6, subnetv6ID, pool.getparent.getid: ", subnetv6ID)
	pool2, err := PGDBConn.OrmCreatePoolv6(subnetv6ID, pool)
	if err != nil {
		log.Println("OrmCreatePoolv6 error")
		log.Println(err)
		return &RestPoolv6{}, err
	}

	pool.SetID(strconv.Itoa(int(pool2.ID)))
	pool.SetCreationTimestamp(pool2.CreatedAt)
	pool.Subnetv6Id = subnetv6ID

	return pool, nil
}

func (r *Poolv6Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go pool Create")

	pool := ctx.Resource.(*RestPoolv6)

	log.Println("in PoolHandler, create(), pool: ", pool)
	log.Println("dhcp/dhcprest. pool.Subnetv4Id: ", pool.Subnetv6Id)
	if _, err := r.CreatePoolv6(pool); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	log.Println("dhcp/dhcprest. pool.id: ", pool.ID)

	return pool, nil
}

func (r *Poolv6Handler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into rest pool Update")

	pool := ctx.Resource.(*RestPoolv6)
	if err := r.UpdatePoolv6(pool); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	return pool, nil
}
func (r *Poolv6Handler) Delete(ctx *resource.Context) *goresterr.APIError {
	pool := ctx.Resource.(*RestPoolv6)

	if err := r.DeletePoolv6(pool); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil

}

func (r *Poolv6Handler) UpdatePoolv6(pool *RestPoolv6) error {
	log.Println("into UpdatePoolv6")
	log.Println(pool)

	r.lock.Lock()
	defer r.lock.Unlock()

	subnetId := pool.GetParent().GetID()
	log.Println("+++subnetId")
	log.Println(subnetId)
	err := PGDBConn.OrmUpdatePoolv6(subnetId, pool)
	if err != nil {
		return err
	}

	return nil
}
func (r *Poolv6Handler) DeletePoolv6(pool *RestPoolv6) error {
	log.Println("into DeletePool, pool.ID: ", pool.ID)
	//log.Println(pool)
	r.lock.Lock()
	defer r.lock.Unlock()

	err := PGDBConn.OrmDeletePoolv6(pool.ID)
	if err != nil {
		return err
	}

	return nil
}
