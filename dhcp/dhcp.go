package dhcp

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/linkingthing/ddi/pb"
	"github.com/sirupsen/logrus"
)

const (
	DhcpHost        = "10.0.0.15"
	DhcpPort        = "8081"
	DhcpConfigPath  = "/usr/local/etc/kea/"
	Dhcp4ConfigFile = "kea-dhcp4.conf"
	Dhcp6ConfigFile = "kea-dhcp6.conf"

	KEADHCPv4Service = "dhcp4"
	KEADHCPv6Service = "dhcp6"
)
const (
	IntfStartDHCPv4 = 1 + iota
	IntfStopDHCPv4
	IntfCreateSubnetv4
	IntfUpdateSubnetv4
	IntfDeleteSubnetv4
)

type ParseConfig struct {
	Result    json.Number
	Arguments DHCPConf
}

type DHCPConf map[string]DhcpConfig

type DhcpConfig struct {
	Authoritative bool   `json:"authoritative"`
	BootFileName  string `json:"boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket  `json:"control-socket"`
	OptionData    []Option       `json:"option-data"`
	Subnet        []SubnetConfig `json:"subnet6"`

	//T1Percent json.Number `json:"t1-percent"`
	//T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type ControlSocket struct {
	SocketName string `json:"socket-name"`
	SocketType string `json:"socket-type"`
}

type SubnetConfig struct {
	Four4o6Interface   string `json:"4o6-interface"`
	Four4o6InterfaceId string `json:"4o6-interface-id"`
	Four4o6Subnet      string `json:"4o6-subnet"`
	Authoritative      bool   `json:"authoritative"`
	CalculateTeeTimes  bool   `json:"calculate-tee-times"`
	//Id json.Number `json:"id"`
	MatchClientId   bool           `json:"match-client-id"`
	NextServer      string         `json:"next-server"`
	OptionData      []Option       `json:"option-data"`
	Pools           []Pool         `json:"pools"`
	RebindTimer     json.Number    `json:"rebind-timer"`
	Relay           SubnetRelay    `json:"relay"`
	RenewTimer      json.Number    `json:"renew-timer"`
	ReservationMode string         `json:"reservation-mode"`
	Reservations    []Reservations `json:"reservations"`
	Subnet          string         `json:"subnet"`

	//T1Percent float64 `json:"t1-percent"`
	//T2Percent float64 `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type SubnetRelay struct {
	IpAddresses []string `json:"ip-addresses"`
}

type Option struct {
	AlwaysSend bool        `json:"always-send"`
	Code       json.Number `json:"code"`
	CsvFormat  bool        `json:"csv-format"`
	Data       string      `json:"data"`
	Name       string      `json:"name"`
	Space      string      `json:"space"`
}
type Pool struct {
	OptionData []Option `json:"option-data"`
	Pool       string   `json:"pool"`
}
type Reservations struct {
	BootFileName string `json:"boot-file-name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid           string   `json:"duid"`
	Hostname       string   `json:"hostname"`
	IpAddress      string   `json:"ip-address"`
	NextServer     string   `json:"next-server"`
	OptionData     []Option `json:"option-data"`
	ServerHostname string   `json:"server-hostname"`
}

type KEAHandler struct {
	ver          string
	ConfigPath   string
	MainConfName string
	//ConfContent  string
	//ViewList     []View
	//FreeACLList  map[string]ACL
}

func NewKEAHandler(ver string, ConfPath string, agentPath string) *KEAHandler {

	instance := &KEAHandler{ver: ConfPath, ConfigPath: ConfPath}

	return instance
}

func (handler *KEAHandler) StartDHCPv4(req pb.StartDHCPv4Req) error {
	startCmd := "nohup keactrl start -s dhcp4 >/dev/null 2>&1 &"

	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + req.Config + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (handler *KEAHandler) StopDHCPv4(req pb.StopDHCPv4Req) error {

	stopCmd := "keactrl stop -s "

	ret, err := cmd(stopCmd)

	if err != nil {
		fmt.Printf("stopCmd ret: %s\n", ret)
		return err
	}

	return nil
}

func (handler *KEAHandler) CreateSubnetv4(req pb.CreateSubnetv4Req) error {

	var conf ParseConfig
	err := getConfig(KEADHCPv4Service, &conf)
	if err != nil {

		return err
	}

	dhcpConf := conf.Arguments
	dhcpConfig := dhcpConf["Dhcp4"]

	for _, v := range dhcpConfig.Subnet {
		if v.Subnet == req.Subnet {
			return fmt.Errorf("subnet %s exists, create failed", req.Subnet)
		}
	}

	newSubnet4 := SubnetConfig{
		ReservationMode: "all",
		Reservations:    []Reservations{},
		OptionData:      []Option{},
		Subnet:          req.Subnet,
		Relay: SubnetRelay{
			IpAddresses: []string{},
		},
		Pools: []Pool{
			{
				[]Option{},
				req.Pool[0].Pool,
			},
		},
	}

	dhcpConfig.Subnet = append(dhcpConfig.Subnet, newSubnet4)

	setErr := setConfig(KEADHCPv4Service, &conf.Arguments)
	if setErr != nil {

		return setErr
	}
	return nil
}

func (handler *KEAHandler) UpdateSubnetv4(req pb.UpdateSubnetv4Req) error {
	var conf ParseConfig
	err := getConfig(KEADHCPv4Service, &conf)
	if err != nil {
		return err
	}

	for k, v := range conf.Arguments["Dhcp4"].Subnet {
		if v.Subnet == req.Subnet {
			conf.Arguments["Dhcp4"].Subnet[k].Pools = []Pool{
				{
					[]Option{},
					req.Pool[0].Pool,
				},
			}
			err = setConfig(KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("subnet %s not exist, update error", req.Subnet)
}

func (handler *KEAHandler) DeleteSubnetv4(req pb.DeleteSubnetv4Req) error {
	var conf ParseConfig
	err := getConfig(KEADHCPv4Service, &conf)
	if err != nil {
		return err
	}

	tmp := conf.Arguments["Dhcp4"].Subnet
	for k, v := range conf.Arguments["Dhcp4"].Subnet {
		if v.Subnet == req.Subnet {
			conf.Arguments["Dhcp4"].Subnet = append(tmp[:k], tmp[k+1:]...)
			err = setConfig(KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func (handler *KEAHandler) StartDHCPv6(req pb.StartDHCPv6Req) error {
	startCmd := "nohup keactrl start -s " + KEADHCPv6Service + " >/dev/null 2>&1 &"

	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + KEADHCPv6Service + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (handler *KEAHandler) StopDHCPv6(req pb.StopDHCPv6Req) error {

	stopCmd := "keactrl stop -s "

	ret, err := cmd(stopCmd)

	if err != nil {
		fmt.Printf("stopCmd ret: %s\n", ret)
		return err
	}

	return nil
}

func (handler *KEAHandler) CreateSubnetv6(req pb.CreateSubnetv6Req) error {

	var conf ParseConfig
	err := getConfig(KEADHCPv6Service, &conf)
	if err != nil {

		return err
	}

	for _, v := range conf.Arguments["Dhcp6"].Subnet {
		if v.Subnet == req.Subnet {
			return fmt.Errorf("subnet %s exists, create failed", req.Subnet)
		}
	}

	newSubnet6 := SubnetConfig{
		ReservationMode: "all",
		Reservations:    []Reservations{},
		OptionData:      []Option{},
		Subnet:          req.Subnet,
		Relay: SubnetRelay{
			IpAddresses: []string{},
		},
		Pools: []Pool{
			{
				[]Option{},
				req.Pool[0].Pool,
			},
		},
	}

	conf.Arguments["Dhcp6"].Subnet = append(conf.Arguments["Dhcp6"].Subnet, newSubnet6)

	setErr := setConfig(KEADHCPv6Service, &conf.Arguments)
	if setErr != nil {

		return setErr
	}
	return nil
}

func (handler *KEAHandler) UpdateSubnetv6(req pb.UpdateSubnetv6Req) error {
	var conf ParseConfig
	err := getConfig(KEADHCPv6Service, &conf)
	if err != nil {
		return err
	}

	for k, v := range conf.Arguments["Dhcp6"].Subnet {
		if v.Subnet == req.Subnet {
			conf.Arguments["Dhcp6"].Subnet[k].Pools = []Pool{
				{
					[]Option{},
					req.Pool[0].Pool,
				},
			}
			err = setConfig(KEADHCPv6Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("subnet %s not exist, update error", req.Subnet)
}

func (handler *KEAHandler) DeleteSubnetv6(req pb.DeleteSubnetv6Req) error {
	var conf ParseConfig
	err := getConfig(KEADHCPv6Service, &conf)
	if err != nil {
		return err
	}

	tmp := conf.Arguments["Dhcp6"].Subnet
	for k, v := range conf.Arguments["Dhcp6"].Subnet {
		if v.Subnet == req.Subnet {
			conf.Arguments["Dhcp6"].Subnet = append(tmp[:k], tmp[k+1:]...)
			err = setConfig(KEADHCPv6Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}
func (handler *KEAHandler) Close() {

}
