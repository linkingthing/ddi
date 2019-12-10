package dhcprest

import (
	"fmt"
	"time"

	goresterr "github.com/ben-han-cn/gorest/error"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	"log"
)

func NewDhcpv4(db *gorm.DB) *Dhcpv4 {
	return &Dhcpv4{db: db}
}

//func NewSubnetv4(db *gorm.DB) *Dhcpv4 {
//	return &Subnetv4State{db: db}
//}

func (s *Dhcpv4) AddSubnetv4(subnetv4 *Subnetv4) error {
	fmt.Println("into AddSubnetv4")
	fmt.Print(subnetv4)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv4ByName(subnetv4.Subnet); c != nil {
		return fmt.Errorf("subnet %s already exist", subnetv4.Subnet)
	}

	err := PGDBConn.CreateSubnetv4(s.db, subnetv4.Subnet, subnetv4.ValidLifetime)
	if err != nil {
		return err
	}

	return nil
}

func (s *Dhcpv4) UpdateSubnetv4(subnetv4 *Subnetv4) error {
	fmt.Println("into UpdateSubnetv4")
	fmt.Print(subnetv4)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv4ByName(subnetv4.Subnet); c == nil {
		return fmt.Errorf("subnet %s not exist", subnetv4.Subnet)
	}

	err := PGDBConn.UpdateSubnetv4(s.db, subnetv4.Subnet, subnetv4.ValidLifetime)
	if err != nil {
		return err
	}

	return nil
}

func (s *Dhcpv4) DeleteSubnetv4(subnetv4 *Subnetv4) error {
	fmt.Println("into dhcprest DeleteSubnetv4")
	fmt.Print(subnetv4)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv4(subnetv4.ID); c == nil {
		return fmt.Errorf("subnet %s not exist", subnetv4.Subnet)
	}

	err := PGDBConn.DeleteSubnetv4(s.db, subnetv4.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Dhcpv4) GetSubnetv4(id string) *Subnetv4 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getSubnetv4(id)
}

func (s *Dhcpv4) getSubnetv4(id string) *Subnetv4 {
	v := PGDBConn.GetSubnetv4(s.db, id)
	if v.ID == 0 {
		return nil
	}

	v4 := s.convertSubnetv4FromOrmToRest(v)
	return v4
}

func (s *Dhcpv4) getSubnetv4ByName(name string) *Subnetv4 {
	log.Println("In dhcprest getSubnetv4ByName, name: ", name)

	v := PGDBConn.GetSubnetv4ByName(s.db, name)
	if v == nil {
		return nil
	}

	v4 := s.convertSubnetv4FromOrmToRest(v)

	return v4
}

func (s *Dhcpv4) GetSubnetv4s() []*Subnetv4 {
	s.lock.Lock()
	defer s.lock.Unlock()

	list := PGDBConn.Subnetv4List(s.db)

	var v4 []*Subnetv4
	for _, v := range list {
		var subnet *Subnetv4
		subnet = s.convertSubnetv4FromOrmToRest(&v)
		v4 = append(v4, subnet)
	}
	return v4
}

type subnetv4Handler struct {
	subnetv4s *Dhcpv4
}

func NewSubnetv4Handler(s *Dhcpv4) *subnetv4Handler {
	return &subnetv4Handler{
		subnetv4s: s,
	}
}

type subnetv4PoolHandler struct {
	subnetv4s *Dhcpv4
}

func (h *subnetv4Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Create")

	subnetv4 := ctx.Resource.(*Subnetv4)
	subnetv4.SetID(subnetv4.Subnet)
	subnetv4.SetCreationTimestamp(time.Now())
	if err := h.subnetv4s.AddSubnetv4(subnetv4); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	} else {
		return subnetv4, nil
	}
}

func (h *subnetv4Handler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Update")

	subnetv4 := ctx.Resource.(*Subnetv4)
	if err := h.subnetv4s.UpdateSubnetv4(subnetv4); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	return subnetv4, nil
}

func (h *subnetv4Handler) Delete(ctx *resource.Context) *goresterr.APIError {
	subnetv4 := ctx.Resource.(*Subnetv4)
	log.Println("into dhcprest.go Delete")
	log.Print(subnetv4)
	log.Println("subnetv4 over")

	if err := h.subnetv4s.DeleteSubnetv4(subnetv4); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil

}

func (h *subnetv4Handler) List(ctx *resource.Context) interface{} {
	log.Println("into dhcprest.go List")

	return h.subnetv4s.GetSubnetv4s()
}

func (h *subnetv4Handler) Get(ctx *resource.Context) resource.Resource {

	log.Println("into dhcprest.go Get")
	return h.subnetv4s.GetSubnetv4(ctx.Resource.GetID())
}

func (r *reservationHandler) List(ctx *resource.Context) interface{} {
	log.Println("into dhcprest.go subnetv4ReservationHandler List")
	rsv := ctx.Resource.(*RestReservation)
	return r.GetSubnetv4Reservations(rsv.GetParent().GetID())
}
