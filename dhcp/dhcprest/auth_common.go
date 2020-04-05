package dhcprest

import (
	"github.com/zdnscloud/gorest/resource"
)

type AuthRest struct {
	resource.ResourceBase `json:"embedded,inline"`
	Username              string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	Password              string
	ValidLifetime         string `json:"validLifetime"`
}
