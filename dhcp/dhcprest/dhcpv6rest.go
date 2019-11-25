package main

import (
	"sync"

	"github.com/ben-han-cn/gorest/resource"
)

var (
	subnetv6Kind = resource.DefaultKindName(Subnetv6{})
)

type Subnetv6 struct {
	resource.ResourceBase `json:",inline"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	ValidLifetime         string `json:"validLifeTime"`
}

type Dhcpv6 struct {
	subnetv6s []*Subnetv6
	lock      sync.Mutex
}

func newDhcpv6() *Dhcpv6 {
	return &Dhcpv6{}
}
