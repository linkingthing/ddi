package ipam

import (
	"log"

	"github.com/zdnscloud/gorest/resource"
)

type DividedAddress struct {
	resource.ResourceBase `json:",inline"`
	IP                    string `json:"ip"`
	MacAddress            string `json:"macaddress"`
	MacVender             string `json:"macvender"`
	AddressType           string `json:"AddressType"`
	OperSystem            string `json:"opersystem"`
	NetBIOSName           string `json:"netbiosname"`
	HostName              string `json:"hostname"`
	InterfaceID           string `json:"interfaceid"`
	ScanTime              int64  `json:"scantime"`
	LastAliveTime         int64  `json:"lastalivetime"`
	FingerPrint           string `json:"fingerprint"`
	LeaseStartTime        int64  `json:"leasestarttime"`
	LeaseEndTime          int64  `json:"leaseendtime"`
	DeviceTypeFlag        bool   `json:"devicetypeflag"`
	DeviceType            string `json:"devicetype"`
	BusinessFlag          bool   `json:"businessflag"`
	Business              string `json:"business"`
	ChargePersonFlag      bool   `json:"chargepersonflag"`
	ChargePerson          string `json:"chargeperson"`
	TelFlag               bool   `json:"telflag"`
	Tel                   string `json:"tel"`
	DepartmentFlag        bool   `json:"departmentflag"`
	Department            string `json:"department"`
	PositionFlag          bool   `json:"positionflag"`
	Position              string `json:"position"`
}

/*
type ScanAddress struct {
	resource.ResourceBase `json:",inline"`
	DetectMethod          string                   `json:"detectmethod"`
	IP                    string                   `json:"ip"`
	MacAddress            string                   `json:"macaddress"`
	MacVender             string                   `json:"macvender"`
	AddressType           string                   `json:"AddressType"`
	OperSystem            string                   `json:"opersystem"`
	NetBIOSName           string                   `json:"netbiosname"`
	HostName              string                   `json:"hostname"`
	InterfaceID           string                   `json:"interfaceid"`
	ScanInterfaceID       string                   `json:"scaninterfaceid"`
	ScanTime              int64                    `json:"scantime"`
	LastAliveTime         int64                    `json:"lastalivetime"`
	FingerPrint           string                   `json:"fingerprint"`
	LeaseStartTime        int64                    `json:"leasestarttime"`
	LeaseEndTime          int64                    `json:"leaseendtime"`
}*/

type IPAttrAppend struct {
	resource.ResourceBase `json:",inline"`
	DeviceTypeFlag        bool
	BusinessFlag          bool
	ChargePersonFlag      bool
	TelFlag               bool
	DepartmentFlag        bool
	PositionFlag          bool
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
	BeginSubnet    string    `json:"beginsubnet"`
	EndSubnet      string    `json:"endsubnet"`
	BeginNodeCode  byte      `json:"beginnodecode"`
	EndNodeCode    byte      `json:"endnodecode"`
	SubtreeBitNum  byte      `json:"subtreebitnum"`
	Depth          int       `json:"depth"`
	SubtreeUseDFor string    `json:"usedfor"`
	Nodes          []Subtree `json:"nodes"`
}

type SplitSubnet struct {
	ID         string `json:"todeleteid"`
	NameFirst  string `json:"namefirst"`
	NameSecond string `json:"namesecond"`
	Bitwith    byte   `json:"bitwith"`
}

type SplitSubnetResult struct {
	First  Subtree `json:"firstnode"`
	Second Subtree `json:"secondnode"`
}

type ChangeData struct {
	Subnetv4Id string `json:"subnetv4Id,omitempty"`
	CurrType   string `json:"currType,omitempty"`
	IpAddress  string `json:"ipAddress,omitempty"`
	HwAddress  string `json:"hwAddress,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	CircuitId  string `json:"circuitId,omitempty"`
	ClientId   string `json:"clientId,omitempty"`
	Duid       string `json:"duid,omitempty"`
	MacAddress string `json:"macaddress,omitempty"`
}

type DividedAddressData struct {
	resource.ResourceBase `json:",inline"`
	Oper                  string     `json:"oper" rest:"required=true,minLen=1,maxLen=20"`
	Data                  ChangeData `json:"data"`
}

func (r DividedAddress) GetActions() []resource.Action {
	log.Println("into DividedAddress GetActions")
	var actions []resource.Action

	action := resource.Action{
		Name:   "change",
		Input:  &DividedAddressData{},
		Output: &DividedAddressData{},
	}
	actions = append(actions, action)

	//log.Println("in cluster GetActions, actions: ", actions)
	return actions
}
func (r DividedAddress) CreateAction(name string) *resource.Action {

	log.Println("into DividedAddress, create action")
	switch name {
	case "change":
		return &resource.Action{
			Name:  "change",
			Input: &DividedAddressData{},
		}
	default:
		return nil
	}
}
