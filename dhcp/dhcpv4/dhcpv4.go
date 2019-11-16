package dhcpv4

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/linkingthing.com/ddi/dhcp"
	"github.com/linkingthing.com/ddi/pb"
	"github.com/sirupsen/logrus"
)

type Parse4Config struct {
	Result    json.Number
	Arguments DHCP4Conf
}
type DHCP4Conf struct {
	Dhcp4 Dhcp4Config
}

type Dhcp4Config struct {
	Authoritative bool   `json:"authoritative"`
	BootFileName  string `json:"boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket   `json:"control-socket"`
	OptionData    []Option        `json:"option-data"`
	Subnet4       []Subnet4Config `json:"subnet4"`

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
	MatchClientId   bool           `json:"match-client-id"`
	NextServer      string         `json:"next-server"`
	OptionData      []Option       `json:"option-data"`
	Pools           []Pool         `json:"pools"`
	RebindTimer     json.Number    `json:"rebind-timer"`
	Relay           Subnet4Relay   `json:"relay"`
	RenewTimer      json.Number    `json:"renew-timer"`
	ReservationMode string         `json:"reservation-mode"`
	Reservations    []Reservations `json:"reservations"`
	Subnet          string         `json:"subnet"`

	//T1Percent float64 `json:"t1-percent"`
	//T2Percent float64 `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type Subnet4Relay struct {
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

type KEAv4Handler struct {
	ConfigPath   string
	MainConfName string
	//ConfContent  string
	//ViewList     []View
	//FreeACLList  map[string]ACL
}

func (t *KEAv4Handler) StartDHCPv4(req pb.StartDHCPv4Req) error {
	startCmd := "nohup keactrl start -s dhcp4 >/dev/null 2>&1 &"

	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + req.Config + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (t *KEAv4Handler) StopDHCPv4(req pb.StopDHCPv4Req) error {

	stopCmd := "keactrl stop -s "

	ret, err := cmd(stopCmd)

	if err != nil {
		fmt.Printf("stopCmd ret: %s\n", ret)
		return err
	}

	return nil
}

func (t *KEAv4Handler) CreateSubnetv4(req pb.CreateSubnetv4Req) error {

	var conf Parse4Config
	err := getConfig(dhcp.KEADHCPv4Service, &conf)
	if err != nil {

		return err
	}

	for _, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			return fmt.Errorf("subnet %s exists, create failed", req.Subnet)
		}
	}

	newSubnet4 := Subnet4Config{
		ReservationMode: "all",
		Reservations:    []Reservations{},
		OptionData:      []Option{},
		Subnet:          req.Subnet,
		Relay: Subnet4Relay{
			IpAddresses: []string{},
		},
		Pools: []Pool{
			{
				[]Option{},
				req.Pool[0].Pool,
			},
		},
	}

	conf.Arguments.Dhcp4.Subnet4 = append(conf.Arguments.Dhcp4.Subnet4, newSubnet4)

	setErr := setConfig(dhcp.KEADHCPv4Service, &conf.Arguments)
	if setErr != nil {

		return setErr
	}
	return nil
}

func (t *KEAv4Handler) UpdateSubnetv4(req pb.UpdateSubnetv4Req) error {
	var conf Parse4Config
	err := getConfig(dhcp.KEADHCPv4Service, &conf)
	if err != nil {
		return err
	}

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			conf.Arguments.Dhcp4.Subnet4[k].Pools = []Pool{
				{
					[]Option{},
					req.Pool[0].Pool,
				},
			}
			err = setConfig(dhcp.KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("subnet %s not exist, update error", req.Subnet)
}

func (t *KEAv4Handler) DeleteSubnetv4(req pb.DeleteSubnetv4Req) error {
	var conf Parse4Config
	err := getConfig(dhcp.KEADHCPv4Service, &conf)
	if err != nil {
		return err
	}

	tmp := conf.Arguments.Dhcp4.Subnet4
	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			conf.Arguments.Dhcp4.Subnet4 = append(tmp[:k], tmp[k+1:]...)
			err = setConfig(dhcp.KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}
