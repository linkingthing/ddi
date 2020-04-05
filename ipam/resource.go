package ipam

import "github.com/zdnscloud/gorest/resource"

type StatusAddress struct {
	MacAddress      string `json:"macaddress"`
	MacVender       string `json:"macvender"`
	AddressType     string `json:"AddressType"`
	OperSystem      string `json:"opersystem"`
	NetBIOSName     string `json:"netbiosname"`
	HostName        string `json:"hostname"`
	InterfaceID     string `json:"interfaceid"`
	ScanInterfaceID string `json:"scaninterfaceid"`
	ScanTime        int64  `json:"scantime"`
	LastAliveTime   int64  `json:"lastalivetime"`
	FingerPrint     string `json:"fingerprint"`
	LeaseStartTime  int64  `json:"leasestarttime"`
	LeaseEndTime    int64  `json:"leaseendtime"`
}

type DividedAddress struct {
	resource.ResourceBase `json:",inline"`
	Data                  map[string]StatusAddress `json:"data"`
}

type ScanAddress struct {
	resource.ResourceBase `json:",inline"`
	DetectMethod          string                   `json:"detect method"`
	Data                  map[string]StatusAddress `json:"data"`
}

/*type GenerationNode struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Subnet         string            `json:"subnet"`
	NodeCode       byte              `json:"nodecode"`
	SubtreeBitNum  byte              `json:"subtreebitnum"`
	Depth          int               `json:"depth"`
	SubtreeUseDFor string            `json:"usedfor"`
	Nodes          []*GenerationNode `json:"nodes"`
}*/
type Subtree struct {
	//resource.ResourceBase `json:",inline"`
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Subnet         string    `json:"subnet"`
	NodeCode       byte      `json:"nodecode"`
	SubtreeBitNum  byte      `json:"subtreebitnum"`
	Depth          int       `json:"depth"`
	SubtreeUseDFor string    `json:"usedfor"`
	Nodes          []Subtree `json:"nodes"`
}
