package dhcp

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linkingthing/ddi/pb"
	"github.com/sirupsen/logrus"
)

const (
	DhcpHost        = "10.0.0.55"
	DhcpPort        = "8081"
	DhcpConfigPath  = "/usr/local/etc/kea/"
	Dhcp4ConfigFile = "kea-dhcp4.conf"
	Dhcp6ConfigFile = "kea-dhcp6.conf"

	KEADHCPv4Service = "dhcp4"
	KEADHCPv6Service = "dhcp6"

	KeaPidPath      = "/usr/local/var/run/kea/"
	KeaDhcp4PidFile = "kea-dhcp4.kea-dhcp4.pid"
	KeaDhcp6PidFile = "kea-dhcp6.kea-dhcp6.pid"

	Dhcpv4AgentAddr = "10.0.0.55:8888"
	Dhcpv6AgentAddr = "10.0.0.55:8889"

	IntfStartDHCPv4 = 1 + iota
	IntfStopDHCPv4
	IntfCreateSubnetv4
	IntfUpdateSubnetv4
	IntfDeleteSubnetv4
)

var KeaDhcpv4Conf []byte // global var, stores config content of dhcpv4 in json format
var KeaDhcpv6Conf []byte // same like dhcpv4 above

//func init() {
//
//	KeaDhcpv4Conf = NewParseDhcpv4Config()
//}

//
//func NewParseDhcpv4Config() *ParseDhcpv4Config {
//
//	dhcpv4Config := new(Dhcpv4Config)
//	dhcpv4Config.ValidLifetime = json.Number(0)
//	dhcpv4Config.Authoritative = false
//	dhcpv4Config.ControlSocket = ControlSocket{}
//	dhcpv4Config.OptionData = []Option{}
//
//	dhcpv4 := new(DHCPv4Conf)
//	dhcpv4.Dhcp4 = *dhcpv4Config
//
//	p := new(ParseDhcpv4Config)
//	p.Result = json.Number(0)
//	p.Arguments = *dhcpv4
//
//	return p
//}

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
	Arguments DHCPv6Conf
}
type DHCPv6Conf struct {
	Dhcp6 Dhcpv6Config
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
	MatchClientId   bool          `json:"match-client-id"`
	NextServer      string        `json:"next-server"`
	OptionData      []Option      `json:"option-data"`
	Pools           []Pool        `json:"pools"`
	RebindTimer     json.Number   `json:"rebind-timer"`
	Relay           SubnetRelay   `json:"relay"`
	RenewTimer      json.Number   `json:"renew-timer"`
	ReservationMode string        `json:"reservation-mode"`
	Reservations    []Reservation `json:"reservations"`
	Subnet          string        `json:"subnet"`

	//T1Percent float64 `json:"t1-percent"`
	//T2Percent float64 `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type SubnetRelay struct {
	IpAddresses []string `json:"ip-addresses"`
}

type Option struct {
	AlwaysSend bool   `json:"always-send"`
	Code       uint64 `json:"code"`
	CsvFormat  bool   `json:"csv-format"`
	Data       string `json:"data"`
	Name       string `json:"name"`
	Space      string `json:"space"`
}
type Pool struct {
	OptionData []Option `json:"option-data"`
	Pool       string   `json:"pool"`
}
type Reservation struct {
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
	mu           sync.Mutex
	ver          string
	ConfigPath   string
	MainConfName string
	//ConfContent  string
	//ViewList     []View
	//FreeACLList  map[string]ACL
}
type KEAv6Handler struct {
	mu           sync.Mutex
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

func (handler *KEAv4Handler) getDhcpv4Config(service string, conf *ParseDhcpv4Config) error {

	handler.mu.Lock()
	defer handler.mu.Unlock()

	postData := map[string]interface{}{
		"command": "config-get",
		"service": []string{service},
	}
	postStr, _ := json.Marshal(postData)

	getCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://" + DhcpHost + ":" + DhcpPort + " 2>/dev/null"

	configJson, err := cmd(getCmd)

	if err != nil {
		return err
	}

	KeaDhcpv4Conf = []byte(string(configJson[2 : len(configJson)-2]))

	err = json.Unmarshal(KeaDhcpv4Conf, conf)
	if err != nil {
		return err
	}

	return nil
}

type curlRet struct {
	result json.Number
	text   string
}

func (handler *KEAv4Handler) setDhcpv4Config(service string, conf *DHCPv4Conf) error {

	log.Print("into  set dhcp config")
	//fmt.Printf("conf: %+v\n", conf)

	handler.mu.Lock()
	defer handler.mu.Unlock()
	postData := map[string]interface{}{
		"command":   "config-set",
		"service":   []string{service},
		"arguments": &conf,
	}
	postStr, _ := json.Marshal(postData)

	curlCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://" + DhcpHost + ":" + DhcpPort + " 2>/dev/null"
	_, err := cmd(curlCmd)

	//log.Print(curlCmd)
	//log.Print("print r")
	//log.Print(r)

	if err != nil {
		return err
	}

	// todo 正则匹配successful.

	//param1 := "-X" + "POST"
	//param2 := "-H" + "\"Content-Type: application/json\""
	//param3 := "-d" + "' " + string(postStr) + "'"
	//param4 := "http://" + DhcpHost + ":" + DhcpPort
	////param5 := "2>/dev/null"
	//if ret, err := shell.Shell("curl", param1, param2, param3, param4); err != nil {
	//	log.Print("shell err")
	//	log.Print(err)
	//	return err
	//} else {
	//	log.Print("shell ok")
	//	log.Print(ret)
	//
	//	var r curlRet
	//	if err := json.Unmarshal([]byte(ret), &r); err != nil {
	//		log.Print("err != nil")
	//		log.Print(err)
	//	} else {
	//		log.Print("err == nil")
	//		log.Print(r)
	//	}
	//
	//}

	KeaDhcpv4Conf = postStr
	return nil
}

func (handler *KEAv4Handler) StartDHCPv4(req pb.StartDHCPv4Req) error {
	startCmd := "nohup keactrl start -s " + KEADHCPv4Service + " >/dev/null 2>&1 &"

	//log.Print("in startdhcp4, cmd: " + startCmd)
	_, err := cmd(startCmd)
	if err != nil {
		logrus.Error("keactrl start -s kea-" + req.Config + " failed")
		return err
	}

	time.Sleep(time.Second)
	return nil
}

func (handler *KEAv4Handler) StopDHCPv4(req pb.StopDHCPv4Req) error {

	KeaDhcpv4Conf = []byte{}

	stopCmd := "keactrl stop -s " + KEADHCPv4Service
	log.Print("in stopdhcp4, cmd: " + stopCmd)
	_, err := cmd(stopCmd)

	if err != nil {
		return err
	}

	return nil
}

func (handler *KEAv4Handler) getv4Config(conf *ParseDhcpv4Config) error {
	if len(KeaDhcpv4Conf) == 0 {
		log.Print("KeaDhcpv4Conf is nil")
		err := handler.getDhcpv4Config(KEADHCPv4Service, conf)
		if err != nil {
			log.Print(err)
			return err
		}
	} else {
		log.Print("KeaDhcpv4Conf is not nil")
		err := json.Unmarshal(KeaDhcpv4Conf, conf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (handler *KEAv4Handler) CreateSubnetv4(req pb.CreateSubnetv4Req) error {
	var conf ParseDhcpv4Config
	if err := handler.getv4Config(&conf); err != nil {
		return err
	}

	for _, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			return fmt.Errorf(req.Subnet + " exists, return")
		}
	}

	newSubnet4 := SubnetConfig{
		ReservationMode: "all",
		Reservations:    []Reservation{},
		OptionData:      []Option{},
		Subnet:          req.Subnet,
		Relay: SubnetRelay{
			IpAddresses: []string{},
		},
	}
	newSubnet4.Pools = []Pool{}

	conf.Arguments.Dhcp4.Subnet4 = append(conf.Arguments.Dhcp4.Subnet4, newSubnet4)
	setErr := handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
	if setErr != nil {
		return setErr
	}
	return nil
}

func (handler *KEAv4Handler) UpdateSubnetv4(req pb.UpdateSubnetv4Req) error {
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
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
			err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
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
	err := handler.getv4Config(&conf)
	if err != nil {
		return err
	}

	tmp := conf.Arguments.Dhcp4.Subnet4
	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			conf.Arguments.Dhcp4.Subnet4 = append(tmp[:k], tmp[k+1:]...)
			err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

func (handler *KEAv4Handler) CreateSubnetv4Pool(req pb.CreateSubnetv4PoolReq) error {

	log.Print("into dhcp.go, CreateSubnetv4Pool")
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Print(err)
		return err
	}
	log.Print("begin conf\n")
	log.Print(conf)
	log.Print("end conf\n")

	//找到subnet， todo 存取数据库前端和后端的subnet对应关系

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		log.Print("in for loop")
		log.Print(v.Subnet)
		log.Print(req.Subnet)
		if v.Subnet == req.Subnet {
			for _, pool := range req.Pool {

				var ops = []Option{}

				if len(pool.Options) > 0 {
					for _, op := range pool.Options {

						var o Option
						o.AlwaysSend = op.AlwaysSend
						o.Code = op.Code
						o.CsvFormat = op.CsvFormat
						o.Data = op.Data
						o.Name = op.Name
						o.Space = op.Space

						ops = append(ops, o)
					}
				}

				var p Pool
				p.Pool = pool.Pool
				//p.OptionData = ops
				p.OptionData = []Option{}
				conf.Arguments.Dhcp4.Subnet4[k].Pools = append(conf.Arguments.Dhcp4.Subnet4[k].Pools, p)
			}
			//log.Print("begin subnet\n")
			//log.Print(conf.Arguments.Dhcp4)
			//log.Print("end subnet\n")

			err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("subnet do not exists, error")
}
func (handler *KEAv4Handler) UpdateSubnetv4Pool(req pb.UpdateSubnetv4PoolReq) error {
	log.Print("into dhcp.go, UpdateSubnetv4Pool")
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Print(err)
		return err
	}

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
		log.Printf("stopCmd ret: %s\n", ret)
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
