package dhcp

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/linkingthing.com/ddi/pb"
)

const (
	DhcpHost = "10.0.0.15"
	DhcpPort = "8081"
	DhcpConfigPath = "/usr/local/etc/kea/"
	Dhcp4ConfigFile = "kea-dhcp4.conf"
	Dhcp6ConfigFile = "kea-dhcp6.conf"
)

type KEAHandler struct {
	ConfigPath   string
	MainConfName string
	//ConfContent  string
	//ViewList     []View
	//FreeACLList  map[string]ACL
}

type ParseConfig struct {
	Result    json.Number
	Arguments DHCPConfig
}
type DHCPConfig struct {
	Dhcp4 Dhcp4Config
}
type Dhcp4Config struct {
	Authoritative bool   `json:"authoritative"`
	BootFileName  string `json:"boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket       `json:"control-socket"`
	OptionData    []Subnet4OptionData `json:"option-data"`
	Subnet4       []Subnet4Config     `json:"subnet4"`

	//T1Percent json.Number `json:"t1-percent"`
	//T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type ControlSocket struct {
	SocketName string `json:"socket-name"`
	SocketType string `json:"socket-type"`
}

type Subnet4Config struct {
	Four4o6Interface   string `json:"4o6-interface"`
	Four4o6InterfaceId string `json:"4o6-interface-id"`
	Four4o6Subnet      string `json:"4o6-subnet"`
	Authoritative      bool   `json:"authoritative"`
	CalculateTeeTimes  bool   `json:"calculate-tee-times"`
	//Id json.Number `json:"id"`
	MatchClientId   bool                  `json:"match-client-id"`
	NextServer      string                `json:"next-server"`
	OptionData      []Subnet4OptionData   `json:"option-data"`
	Pools           []Subnet4Pools        `json:"pools"`
	RebindTimer     json.Number           `json:"rebind-timer"`
	Relay           Subnet4Relay          `json:"relay"`
	RenewTimer      json.Number           `json:"renew-timer"`
	ReservationMode string                `json:"reservation-mode"`
	Reservations    []Subnet4Reservations `json:"reservations"`
	Subnet          string                `json:"subnet"`

	//T1Percent float64 `json:"t1-percent"`
	//T2Percent float64 `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type Subnet4Relay struct {
	IpAddresses []string `json:"ip-addresses"`
}

type Subnet4OptionData struct {
	AlwaysSend bool        `json:"always-send"`
	Code       json.Number `json:"code"`
	CsvFormat  bool        `json:"csv-format"`
	Data       string      `json:"data"`
	Name       string      `json:"name"`
	Space      string      `json:"space"`
}
type Subnet4Pools struct {
	OptionData []Subnet4OptionData `json:"option-data"`
	Pool       string              `json:"pool"`
}
type Subnet4Reservations struct {
	BootFileName string `json:"boot-file-name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid           string              `json:"duid"`
	Hostname       string              `json:"hostname"`
	IpAddress      string              `json:"ip-address"`
	NextServer     string              `json:"next-server"`
	OptionData     []Subnet4OptionData `json:"option-data"`
	ServerHostname string              `json:"server-hostname"`
}

func (t *KEAHandler) StartDHCP(req pb.DHCPStartReq) error {
	startCmd := "nohup keactrl start -s " + req.Service + " >/dev/null 2>&1 &"

	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + req.ConfigFile + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (t *KEAHandler) StopDHCP(req pb.DHCPStopReq) error {

	stopCmd := "keactrl stop -s " + req.Service

	ret, err := cmd(stopCmd)

	if err != nil {
		fmt.Printf("stopCmd ret: %s\n", ret)
		return err
	}

	return nil
}

func (t *KEAHandler) CreateSubnet4(req pb.CreateSubnet4Req) error {
	//CreateSubnet4(service string, subnetName string, pools string)
	var conf ParseConfig
	err := getConfig(req.Service, &conf)
	if err != nil {
		return err
	}

	for _, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.SubnetName {
			return fmt.Errorf("subnet %s exists, create failed", req.SubnetName)
		}
	}

	newSubnet4 := Subnet4Config{
		ReservationMode: "all",
		Reservations:    []Subnet4Reservations{},
		OptionData:      []Subnet4OptionData{},
		Subnet:          req.SubnetName,
		Relay: Subnet4Relay{
			IpAddresses: []string{},
		},
		Pools: []Subnet4Pools{
			{
				[]Subnet4OptionData{},
				req.Pools,
			},
		},
	}

	conf.Arguments.Dhcp4.Subnet4 = append(conf.Arguments.Dhcp4.Subnet4, newSubnet4)
	setErr := setConfig(req.SubnetName, &conf.Arguments)
	if setErr != nil {
		return setErr
	}
	return nil
}

func (t *KEAHandler) updateSubnet4(req pb.UpdateSubnet4Req) error {
	var conf ParseConfig
	err := getConfig(req.Service, &conf)
	if err != nil {
		return err
	}

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.SubnetName {
			conf.Arguments.Dhcp4.Subnet4[k].Pools = []Subnet4Pools{
				{
					[]Subnet4OptionData{},
					req.Pools,
				},
			}
			err = setConfig(req.Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("subnet %s not exist, update error", req.SubnetName)
}

func (t *KEAHandler) deleteSubnet4(req pb.DeleteSubnet4Req) error {
	var conf ParseConfig
	err := getConfig(req.Service, &conf)
	if err != nil {
		return err
	}

	tmp := conf.Arguments.Dhcp4.Subnet4
	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.SubnetName {
			conf.Arguments.Dhcp4.Subnet4 = append(tmp[:k], tmp[k+1:]...)
			err = setConfig(req.Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}