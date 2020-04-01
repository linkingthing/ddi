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
