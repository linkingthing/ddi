package dhcp

import (
	"fmt"
	"time"

	"encoding/json"

	"log"

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

	KeaPidPath      = "/usr/local/var/run/kea/"
	KeaDhcp4PidFile = "kea-dhcp4.kea-dhcp4.pid"
	KeaDhcp6PidFile = "kea-dhcp6.kea-dhcp6.pid"

	Dhcpv4AgentAddr = "localhost:8888"
	Dhcpv6AgentAddr = "localhost:8889"

	IntfStartDHCPv4 = 1 + iota
	IntfStopDHCPv4
	IntfCreateSubnetv4
	IntfUpdateSubnetv4
	IntfDeleteSubnetv4
)

type ParseDhcpv4Config struct {
	Result    json.Number
	Arguments DHCPv4Conf
}
type DHCPv4Conf struct {
	Dhcp4 Dhcpv4Config
}
type Dhcpv4Config struct {
	Authoritative bool   `json:"authoritative"`
	BootFileName  string `json:"boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket  `json:"control-socket"`
	OptionData    []Option       `json:"option-data"`
	Subnet4       []SubnetConfig `json:"subnet4"`

	//T1Percent json.Number `json:"t1-percent"`
	//T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}

type ParseDhcpv6Config struct {
	Result    json.Number
	Arguments DHCPv4Conf
}
type DHCPv6Conf struct {
	Dhcp4 Dhcpv6Config
}
type Dhcpv6Config struct {
	Authoritative bool   `json:"authoritative"`
	BootFileName  string `json:"boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket  `json:"control-socket"`
	OptionData    []Option       `json:"option-data"`
	Subnet6       []SubnetConfig `json:"subnet6"`

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

type KEAv4Handler struct {
	ver          string
	ConfigPath   string
	MainConfName string
	//ConfContent  string
	//ViewList     []View
	//FreeACLList  map[string]ACL
}
type KEAv6Handler struct {
	ver          string
	ConfigPath   string
	MainConfName string
}

func NewKEAv4Handler(ver string, ConfPath string, addr string) *KEAv4Handler {

	instance := &KEAv4Handler{ver: ver, ConfigPath: ConfPath}

	return instance
}
func NewKEAv6Handler(ver string, ConfPath string, addr string) *KEAv6Handler {

	instance := &KEAv6Handler{ver: ver, ConfigPath: ConfPath}

	return instance
}

func (handler *KEAv4Handler) StartDHCPv4(req pb.StartDHCPv4Req) error {
	startCmd := "nohup keactrl start -s " + KEADHCPv4Service + " >/dev/null 2>&1 &"

	log.Print("in startdhcp4, cmd: " + startCmd)
	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + req.Config + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (handler *KEAv4Handler) StopDHCPv4(req pb.StopDHCPv4Req) error {

	stopCmd := "keactrl stop -s " + KEADHCPv4Service
	log.Print("in stopdhcp4, cmd: " + stopCmd)
	_, err := cmd(stopCmd)

	if err != nil {

		return err
	}

	return nil
}

func (handler *KEAv4Handler) CreateSubnetv4(req pb.CreateSubnetv4Req) error {
	//log.Print("into dhcp.go, CreateSubnetv4")
	var conf ParseDhcpv4Config
	err := getDhcpv4Config(KEADHCPv4Service, &conf)
	if err != nil {

		log.Print(err)
		return err
	}

	//var p *DhcpConfig
	//p = nil
	dhcpv4Config := conf.Arguments.Dhcp4
	//log.Print("before dhcpConfig\n")
	//log.Print(dhcpConfig)
	//log.Print("after  dhcpConfig\n")

	for _, v := range dhcpv4Config.Subnet4 {
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
		//Pools: []Pool{
		//	{
		//		[]Option{},
		//		req.Pool[0].Pool,
		//	},
		//},
	}

	if req.Pool != nil {

	}

	dhcpv4Config.Subnet4 = append(dhcpv4Config.Subnet4, newSubnet4)

	setErr := setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
	if setErr != nil {

		log.Print(setErr)
		return setErr
	}
	return nil
}

func (handler *KEAv4Handler) UpdateSubnetv4(req pb.UpdateSubnetv4Req) error {
	var conf ParseDhcpv4Config
	err := getDhcpv4Config(KEADHCPv4Service, &conf)
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
			err = setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("subnet %s not exist, update error", req.Subnet)
}

func (handler *KEAv4Handler) DeleteSubnetv4(req pb.DeleteSubnetv4Req) error {
	var conf ParseDhcpv4Config
	err := getDhcpv4Config(KEADHCPv4Service, &conf)
	if err != nil {
		return err
	}

	dhcp := conf.Arguments.Dhcp4
	tmp := conf.Arguments.Dhcp4.Subnet4
	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			dhcp.Subnet4 = append(tmp[:k], tmp[k+1:]...)
			err = setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func (handler *KEAv4Handler) CreateSubnetv4Pool(req pb.CreateSubnetv4PoolReq) error {

	return nil
}
func (handler *KEAv4Handler) UpdateSubnetv4Pool(req pb.UpdateSubnetv4PoolReq) error {

	return nil
}
func (handler *KEAv4Handler) DeleteSubnetv4Pool(req pb.DeleteSubnetv4PoolReq) error {

	return nil
}
func (handler *KEAv4Handler) CreateSubnetv4Reservation(req pb.CreateSubnetv4ReservationReq) error {

	return nil
}
func (handler *KEAv4Handler) UpdateSubnetv4Reservation(req pb.UpdateSubnetv4ReservationReq) error {

	return nil
}
func (handler *KEAv4Handler) DeleteSubnetv4Reservation(req pb.DeleteSubnetv4ReservationReq) error {

	return nil
}
func (handler *KEAv4Handler) Close() {

}

func (handler *KEAv6Handler) StartDHCPv6(req pb.StartDHCPv6Req) error {
	startCmd := "nohup keactrl start -s " + KEADHCPv6Service + " >/dev/null 2>&1 &"

	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + KEADHCPv6Service + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (handler *KEAv6Handler) StopDHCPv6(req pb.StopDHCPv6Req) error {

	stopCmd := "keactrl stop -s " + KEADHCPv6Service

	ret, err := cmd(stopCmd)

	if err != nil {
		fmt.Printf("stopCmd ret: %s\n", ret)
		return err
	}

	return nil
}

func (handler *KEAv6Handler) CreateSubnetv6(req pb.CreateSubnetv6Req) error {

	return nil
}

func (handler *KEAv6Handler) UpdateSubnetv6(req pb.UpdateSubnetv6Req) error {
	return nil
}

func (handler *KEAv6Handler) DeleteSubnetv6(req pb.DeleteSubnetv6Req) error {

	return nil
}
func (handler *KEAv6Handler) Close() {

}
