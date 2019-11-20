package main

import "github.com/ben-han-cn/gorest/resource"

var (
	subnetv6Kind = resource.DefaultKindName(Subnetv6{})
)

type Dhcpv6Serv struct {
	resource.ResourceBase `json:",inline"`
	ConfigJson            string `json:"configJson" rest:"required=true,minLen=1,maxLen=1000000"`
}

type Subnetv6 struct {
	resource.ResourceBase `json:",inline"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	ValidLifetime         string `json:"validLifeTime"`
}
