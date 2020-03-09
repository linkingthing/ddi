package dhcprest

import (
	"github.com/ben-han-cn/gorest/resource"
)

type AuthRest struct {
	resource.ResourceBase `json:"embedded,inline"`
	Username              string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	Password              string
	ValidLifetime         string `json:"validLifeTime"`
}
