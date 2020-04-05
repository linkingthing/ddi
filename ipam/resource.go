package ipam

import "github.com/ben-han-cn/gorest/resource"

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

type IPNodes struct {
	ID       string `json:"id"`
	NodeCode byte   `json:"nodecode"`
	NodeName string `json:"nodename"`
	Subnet   string `json:"subnet"`
}
type AlloPrefix struct {
	ParentID     string    `json:"parentid"`
	ParentIPv6   string    `json:"parentipv6"`
	PrefixLength byte      `json:"parentprefixlength"`
	BitsUsedFor  string    `json:"bitsusedfor"`
	BitNum       byte      `json:"bitnum"`
	Depth        int       `json:"depth"`
	Nodes        []IPNodes `json:"nodes"`
}

type GenerationNodes struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Subnet         string            `json:"subnet"`
	NodeCode       byte              `json:"nodecode"`
	SubtreeBitNum  byte              `json:"subtreebitnum"`
	Depth          int               `json:"depth"`
	SubtreeUseDFor string            `json:"usedfor"`
	Nodes          []GenerationNodes `json:"nodes"`
}

type NodesTree struct {
	Nodes GenerationNodes `json:"nodes"`
}
