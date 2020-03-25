package ipam

import "github.com/ben-han-cn/gorest/resource"

type DividedAddress struct {
	resource.ResourceBase `json:",inline"`
	Reserved              []string `json:"reserved"`
	Dynamic               []string `json:"dynamic"`
	Stable                []string `json:"stable"`
	Manual                []string `json:"manual"`
	Lease                 []string `json:"lease"`
}

type ScanAddress struct {
	resource.ResourceBase `json:",inline"`
	Collision             []string `json:"collision"`
	Dead                  []string `json:"dead"`
}
