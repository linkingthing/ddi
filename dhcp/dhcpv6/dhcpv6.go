package dhcpv6

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/linkingthing.com/ddi/dhcp"
	"github.com/linkingthing.com/ddi/pb"
	"github.com/sirupsen/logrus"
)

type Parse6Config struct {
	Result    json.Number
	Arguments DHCP6Conf
}
type DHCP6Conf struct {
	Dhcp6 Dhcp6Config
}

type Dhcp6Config struct {
	Authoritative bool   `json:"authoritative"`
	BootFileName  string `json:"boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket   `json:"control-socket"`
	OptionData    []Option        `json:"option-data"`
	Subnet6       []Subnet6Config `json:"subnet6"`

	//T1Percent json.Number `json:"t1-percent"`
	//T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type ControlSocket struct {
	SocketName string `json:"socket-name"`
	SocketType string `json:"socket-type"`
}

type Subnet6Config struct {
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
	Relay           Subnet6Relay   `json:"relay"`
	RenewTimer      json.Number    `json:"renew-timer"`
	ReservationMode string         `json:"reservation-mode"`
	Reservations    []Reservations `json:"reservations"`
	Subnet          string         `json:"subnet"`

	//T1Percent float64 `json:"t1-percent"`
	//T2Percent float64 `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type Subnet6Relay struct {
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
type KEAv6Handler struct {
	ConfigPath   string
	MainConfName string
	//ConfContent  string
	//ViewList     []View
	//FreeACLList  map[string]ACL
}

func (t *KEAv6Handler) StartDHCPv6(req pb.StartDHCPv6Req) error {
	startCmd := "nohup keactrl start -s " + dhcp.KEADHCPv6Service + " >/dev/null 2>&1 &"

	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + dhcp.KEADHCPv6Service + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (t *KEAv6Handler) StopDHCPv6(req pb.StopDHCPv6Req) error {

	stopCmd := "keactrl stop -s "

	ret, err := cmd(stopCmd)

	if err != nil {
		fmt.Printf("stopCmd ret: %s\n", ret)
		return err
	}

	return nil
}

func (t *KEAv6Handler) CreateSubnetv6(req pb.CreateSubnetv6Req) error {

	var conf Parse6Config
	err := getConfigv6(dhcp.KEADHCPv6Service, &conf)
	if err != nil {

		return err
	}

	for _, v := range conf.Arguments.Dhcp6.Subnet6 {
		if v.Subnet == req.Subnet {
			return fmt.Errorf("subnet %s exists, create failed", req.Subnet)
		}
	}

	newSubnet6 := Subnet6Config{
		ReservationMode: "all",
		Reservations:    []Reservations{},
		OptionData:      []Option{},
		Subnet:          req.Subnet,
		Relay: Subnet6Relay{
			IpAddresses: []string{},
		},
		Pools: []Pool{
			{
				[]Option{},
				req.Pool[0].Pool,
			},
		},
	}

	conf.Arguments.Dhcp6.Subnet6 = append(conf.Arguments.Dhcp6.Subnet6, newSubnet6)

	setErr := setConfigv6(dhcp.KEADHCPv6Service, &conf.Arguments)
	if setErr != nil {

		return setErr
	}
	return nil
}

func (t *KEAv6Handler) UpdateSubnetv6(req pb.UpdateSubnetv6Req) error {
	var conf Parse6Config
	err := getConfigv6(dhcp.KEADHCPv6Service, &conf)
	if err != nil {
		return err
	}

	for k, v := range conf.Arguments.Dhcp6.Subnet6 {
		if v.Subnet == req.Subnet {
			conf.Arguments.Dhcp6.Subnet6[k].Pools = []Pool{
				{
					[]Option{},
					req.Pool[0].Pool,
				},
			}
			err = setConfigv6(dhcp.KEADHCPv6Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("subnet %s not exist, update error", req.Subnet)
}

func (t *KEAv6Handler) DeleteSubnetv6(req pb.DeleteSubnetv6Req) error {
	var conf Parse6Config
	err := getConfigv6(dhcp.KEADHCPv6Service, &conf)
	if err != nil {
		return err
	}

	tmp := conf.Arguments.Dhcp6.Subnet6
	for k, v := range conf.Arguments.Dhcp6.Subnet6 {
		if v.Subnet == req.Subnet {
			conf.Arguments.Dhcp6.Subnet6 = append(tmp[:k], tmp[k+1:]...)
			err = setConfigv6(dhcp.KEADHCPv6Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}
