package main

import (
	"sync"

	"fmt"
	"time"

	"github.com/ben-han-cn/gorest"
	"github.com/ben-han-cn/gorest/adaptor"
	goresterr "github.com/ben-han-cn/gorest/error"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/ben-han-cn/gorest/resource/schema"
	"github.com/gin-gonic/gin"
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

type Subnetv4 struct {
	resource.ResourceBase `json:",inline"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	ValidLifetime         string `json:"validLifeTime"`
}

type Dhcpv4 struct {
	subnetv4s []*Subnetv4
	lock      sync.Mutex
}

func newDhcpv4() *Dhcpv4 {
	return &Dhcpv4{}
}

func (s *Dhcpv4) AddSubnetv4(subnetv4 *Subnetv4) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv4(subnetv4.Subnet); c != nil {
		return fmt.Errorf("cluster %s already exist", subnetv4.Subnet)
	}
	s.subnetv4s = append(s.subnetv4s, subnetv4)
	return nil
}

func (s *Dhcpv4) GetSubnetv4(subnet string) *Subnetv4 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getSubnetv4(subnet)
}
func (s *Dhcpv4) getSubnetv4(subnet string) *Subnetv4 {
	for _, c := range s.subnetv4s {
		if c.Subnet == subnet {
			return c
		}
	}
	return nil
}
func (s *Dhcpv4) GetSubnetv4s() []*Subnetv4 {
	s.lock.Lock()
	defer s.lock.Unlock()

	v4 := make([]*Subnetv4, len(s.subnetv4s))
	copy(v4, s.subnetv4s)
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
func main() {
	schemas := schema.NewSchemaManager()
	dhcpv4 := newDhcpv4()
	schemas.Import(&version, Subnetv4{}, newSubnetv4Handler(dhcpv4))

	router := gin.Default()
	adaptor.RegisterHandler(router, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
	router.Run("0.0.0.0:1234")
}
