package dhcp

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"strconv"

	"strings"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/linkingthing/ddi/dhcp/postgres"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	"github.com/sirupsen/logrus"
)

const (
	DhcpHostlocal   = "127.0.0.1"
	DhcpHost        = "10.0.0.31"
	DhcpPort        = "8000"
	DhcpConfigPath  = "/usr/local/etc/kea/"
	Dhcp4ConfigFile = "kea-dhcp4.conf"
	Dhcp6ConfigFile = "kea-dhcp6.conf"

	KEADHCPv4Service = "dhcp4"
	KEADHCPv6Service = "dhcp6"

	KeaPidPath      = "/usr/local/var/run/kea/"
	KeaDhcp4PidFile = "kea-dhcp4.kea-dhcp4.pid"
	KeaDhcp6PidFile = "kea-dhcp6.kea-dhcp6.pid"

	IntfStartDHCPv4 = 1 + iota
	IntfStopDHCPv4
	IntfCreateSubnetv4
	IntfUpdateSubnetv4
	IntfDeleteSubnetv4
	postgresqlAddress = "host=127.0.0.1 port=5432 user=ddi dbname=ddi password=linkingthing.com sslmode=disable"
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
type CmdRet struct {
	Result int    `json:"result"`
	Text   string `json:"text"`
}

type ParseDhcpv4Config struct {
	Result    json.Number
	Arguments DHCPv4Conf
}
type DHCPv4Conf struct {
	Dhcp4 Dhcpv4Config
}
type Dhcpv4Config struct {
	Authoritative bool   `json:"authoritative,omitempty"`
	BootFileName  string `json:"boot-file-name,omitempty"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket  `json:"control-socket,omitempty"`
	OptionData    []Option       `json:"option-data,omitempty"`
	Subnet4       []SubnetConfig `json:"subnet4"`

	//T1Percent json.Number `json:"t1-percent"`
	//T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime,omitempty"`
}

type ControlSocket struct {
	SocketName string `json:"socket-name,omitempty"`
	SocketType string `json:"socket-type,omitempty"`
}

type SubnetConfig struct {
	//Four4o6Interface   string        `json:"4o6-interface"`
	//Four4o6InterfaceId string        `json:"4o6-interface-id"`
	//Four4o6Subnet      string        `json:"4o6-subnet"`
	//Authoritative     bool          `json:"authoritative"`
	//CalculateTeeTimes bool          `json:"calculate-tee-times"`
	Id json.Number `json:"id"`
	//MatchClientId   bool          `json:"match-client-id"`
	//NextServer      string        `json:"next-server"`
	OptionData []Option `json:"option-data,omitempty"`
	Pools      []Pool   `json:"pools,omitempty"`
	//RebindTimer     json.Number   `json:"rebind-timer"`
	//Relay           SubnetRelay   `json:"relay"`
	//RenewTimer      json.Number   `json:"renew-timer"`
	ReservationMode string        `json:"reservation-mode"`
	Reservations    []Reservation `json:"reservations"`
	Subnet          string        `json:"subnet"`

	//T1Percent float64 `json:"t1-percent"`
	//T2Percent float64 `json:"t2-percent"`
	ValidLifetime    json.Number `json:"valid-lifetime,omitempty"`
	MaxValidLifetime json.Number `json:"max-valid-lifetime,omitempty"`
}
type SubnetRelay struct {
	IpAddresses []string `json:"ip-addresses"`
}

type Option struct {
	AlwaysSend bool   `json:"always-send,omitempty"`
	Code       uint64 `json:"code,omitempty"`
	CsvFormat  bool   `json:"csv-format,omitempty"`
	Data       string `json:"data,omitempty"`
	Name       string `json:"name,omitempty"`
	Space      string `json:"space,omitempty"`
}
type Pool struct {
	OptionData []*Option `json:"option-data"`
	Pool       string    `json:"pool"`
}
type Reservation struct {
	BootFileName string `json:"boot-file-name,omitempty"`
	//ClientClasses []interface{} `json:"client-classes"`
	ClientId       string    `json:"client-id,omitempty"` //reservations can be multi-types, need to split  todo
	Duid           string    `json:"duid,omitempty"`
	Hostname       string    `json:"hostname,omitempty"`
	IpAddress      string    `json:"ip-address,omitempty"`
	HwAddress      string    `json:"hw-address,omitempty"`
	CircuitId      string    `json:"circuit-id,omitempty"`
	NextServer     string    `json:"next-server,omitempty"`
	OptionData     []*Option `json:"option-data,omitempty"`
	ServerHostname string    `json:"server-hostname,omitempty"`
}

type KEAv4Handler struct {
	db           *gorm.DB
	mu           sync.Mutex
	ver          string
	ConfigPath   string
	MainConfName string
	//ConfContent  string
	//ViewList     []View
	//FreeACLList  map[string]ACL
}

func NewKEAv4Handler(ver string, ConfPath string, addr string) *KEAv4Handler {
	instance := &KEAv4Handler{ver: ver, ConfigPath: ConfPath}
	var err error

	instance.db, err = gorm.Open("postgres", postgresqlAddress)
	yamlConfig := config.GetConfig("/etc/vanguard/vanguard.conf")
	if yamlConfig.Localhost.IP == "10.0.0.55" {
		log.Println("in NewKEAv4Handler, use db:  utils.DBAddr")
		instance.db, err = gorm.Open("postgres", utils.DBAddr)
		if err != nil {
			log.Fatal(err)
		}
	}

	return instance
}

func (handler *KEAv4Handler) GetDhcpv4Config(service string, conf *ParseDhcpv4Config) error {

	handler.mu.Lock()
	defer handler.mu.Unlock()

	postData := map[string]interface{}{
		"command": "config-get",
		"service": []string{service},
	}
	postStr, _ := json.Marshal(postData)

	getCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://" + DhcpHost + ":" + DhcpPort + " 2>/dev/null"

	//log.Println("in GetDhcpv4config, getCmd: ", getCmd)
	configJson, err := cmd(getCmd)

	if err != nil {
		return err
	}
	//log.Println("config json: ", configJson)
	log.Println("dhcphost: ", DhcpHost)
	log.Println("DhcpPort: ", DhcpPort)

	KeaDhcpv4Conf = []byte(string(configJson[2 : len(configJson)-2]))

	err = json.Unmarshal(KeaDhcpv4Conf, conf)
	if err != nil {
		return err
	}

	return nil
}

func (handler *KEAv4Handler) setDhcpv4Config(service string, conf *DHCPv4Conf) error {

	log.Print("dhcp/dhcp.go, into setDhcpv4Config()")
	//fmt.Printf("conf: %+v\n", conf)

	handler.mu.Lock()
	defer handler.mu.Unlock()
	postData := map[string]interface{}{
		"command":   "config-set",
		"service":   []string{service},
		"arguments": conf,
	}
	postStr, err := json.Marshal(postData)
	if err != nil {
		log.Println("json.Marshal error: ", err)
	}

	//log.Println("postStr: ", postStr)
	curlCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://" + DhcpHost + ":" + DhcpPort + " 2>/dev/null"
	log.Println("curlCmd: ", curlCmd)
	var cmdRet CmdRet
	str, err := cmd(curlCmd)
	if err != nil {
		log.Println("cmd Error, err: ", err)
		return err
	}
	//log.Println("cmd ret: ", str)
	if err := json.Unmarshal([]byte(str[1:len(str)-1]), &cmdRet); err != nil {
		log.Println("cmd unmarshal Error, err: ", err)
		return err
	}
	if cmdRet.Result != 0 {
		log.Println("set dhcpv4 config Error, err: ", cmdRet.Text)
		return fmt.Errorf(cmdRet.Text)
	}
	//log.Print(curlCmd)
	//log.Print("print r")
	//log.Print(r)

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
	//var r curlRet
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
		err := handler.GetDhcpv4Config(KEADHCPv4Service, conf)
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
	log.Println("in getv4Config, conf: ", conf)
	return nil
}

func (handler *KEAv4Handler) CreateSubnetv4(req pb.CreateSubnetv4Req) error {
	log.Println("into dhcp/dhcp.go CreateSubnetv4(), req.subnet: ", req.Subnet)
	var conf ParseDhcpv4Config
	if err := handler.getv4Config(&conf); err != nil {
		return err
	}

	//var subnetv4 []SubnetConfig
	var maxId int
	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		//log.Println("conf Subnet4: ", v.Subnet)
		//log.Println("conf Subnet4 id: ", v.Id, ", maxId: ", maxId)
		curId, err := strconv.Atoi(string(v.Id))
		if err != nil {
			return err
		}
		if curId >= maxId {
			maxId = curId + 1
		}
		if v.ReservationMode == "" {
			log.Println("reserationMode == nil, subnet: ", v.Subnet)
			conf.Arguments.Dhcp4.Subnet4[k].ReservationMode = "all"
		}
		if v.Subnet == req.Subnet {
			return fmt.Errorf(req.Subnet + " exists, return")
		}
		//subnetv4 = append(subnetv4, v)
	}

	newSubnet4 := SubnetConfig{
		ReservationMode: "all",
		Reservations:    []Reservation{},
		OptionData:      []Option{},
		Subnet:          req.Subnet,
		ValidLifetime:   json.Number(req.ValidLifetime),
		Id:              json.Number(strconv.Itoa(maxId)),
		//Relay: SubnetRelay{
		//	IpAddresses: []string{},
		//},
	}
	newSubnet4.Pools = []Pool{}
	//subnetv4 = append(subnetv4, newSubnet4)
	//log.Println("---subnetv4: ", subnetv4)

	log.Println("req.gateway: ", req.Gateway)
	if len(req.Gateway) > 0 {
		option := Option{
			Name: "routers",
			Data: req.Gateway,
		}
		options := []Option{}
		options = append(options, option)
		newSubnet4.OptionData = options
		log.Println("new subnetv4 optionData: ", newSubnet4.OptionData)
	}

	conf.Arguments.Dhcp4.Subnet4 = append(conf.Arguments.Dhcp4.Subnet4, newSubnet4)
	//log.Println("---2 subnetv4: ", conf.Arguments.Dhcp4.Subnet4)
	setErr := handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
	if setErr != nil {
		return setErr
	}
	return nil
}

func (handler *KEAv4Handler) UpdateSubnetv4(req pb.UpdateSubnetv4Req) error {
	log.Println("into dhcp/UpdateSubnetv4, req.subnet: ", req.Subnet)
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Println("in dhcp/UpdateSubnetv4(), getv4config error: ", err)
		return err
	}

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			log.Println("v.Subnet: ", v.Subnet)
			conf.Arguments.Dhcp4.Subnet4[k].ValidLifetime = json.Number(req.ValidLifetime)
			if len(req.Pool) > 0 {
				log.Println("req.pool: ", req.Pool)
				conf.Arguments.Dhcp4.Subnet4[k].Pools = []Pool{
					{ //p.OptionData = ops
						[]*Option{},
						req.Pool[0].Pool,
					},
				}
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
	//log.Println("into dhcp/DeleteSubnetv4, req.id: ", req.Id)
	log.Println("into dhcp/DeleteSubnetv4, req.Subnet: ", req.Subnet)
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		return err
	}

	//todo,loop and found subnet id
	//tmp := conf.Arguments.Dhcp4.Subnet4
	tmp := []SubnetConfig{}
	flag := false
	for _, v := range conf.Arguments.Dhcp4.Subnet4 {
		//log.Println("dhcp/DeleteSubnetv4, k: ", k, ", v: ", v)
		if v.Subnet != req.Subnet {
			tmp = append(tmp, v)
		} else {
			flag = true
		}
	}

	if flag {
		conf.Arguments.Dhcp4.Subnet4 = tmp
		err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
		if err != nil {
			return err
		}
	}
	return nil
}

func (handler *KEAv4Handler) CreateSubnetv4Pool(req pb.CreateSubnetv4PoolReq) error {

	log.Println("into dhcp.go, CreateSubnetv4Pool, req: ", req)
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}
	//log.Println("begin conf\n")
	//log.Println(conf)
	//log.Println("end conf\n")

	//找到subnet， todo 存取数据库前端和后端的subnet对应关系

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		//log.Print("in for loop, v.Id: ", v.Id, ", req.Id: ", req.Id)
		//log.Print("v.subnet: ", v.Subnet)
		//log.Print("req.Subnet: ", req.Subnet)
		if v.Subnet == req.Subnet {
			log.Println("req.validlifetime: ", req.ValidLifetime)
			log.Println("req.MaxValidLifetime: ", req.MaxValidLifetime)
			if len(req.ValidLifetime) > 0 {

				if err != nil {
					log.Println("CreateSubnetv4Pool, validLifetime error, ", err)
					return err
				}

				conf.Arguments.Dhcp4.Subnet4[k].ValidLifetime = json.Number(req.ValidLifetime)
			}
			if len(req.MaxValidLifetime) > 0 {

				if err != nil {
					log.Println("CreateSubnetv4Pool, validLifetime error, ", err)
					return err
				}

				conf.Arguments.Dhcp4.Subnet4[k].MaxValidLifetime = json.Number(req.MaxValidLifetime)
			}
			for _, pool := range req.Pool {

				var p Pool
				p.Pool = pool.Pool
				p.OptionData = []*Option{}

				//p.OptionData = ops
				var ops []*Option
				if ops, err = ConvertOptionsFromPb(req.Options); err != nil {
					log.Println("ConvertOptionsFromPb error: ", err)
					return err
				}
				p.OptionData = ops

				conf.Arguments.Dhcp4.Subnet4[k].Pools = append(conf.Arguments.Dhcp4.Subnet4[k].Pools, p)
			}
			//log.Println("begin subnet\n")
			//log.Println(conf.Arguments.Dhcp4)
			//log.Println("end subnet\n")

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
	log.Println("into dhcp.go, UpdateSubnetv4Pool")
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}
	//log.Println("begin conf\n")
	//log.Println(conf.Arguments.Dhcp4.Subnet4)
	//log.Println("end conf\n")

	changeFlag := false

	//find current pool, and replace it with new pool
	for k, v := range conf.Arguments.Dhcp4.Subnet4 {

		//log.Print("v.subnet: ", v.Subnet)
		//log.Print("req.Subnet: ", req.Subnet)
		if v.Subnet == req.Subnet {
			log.Println("req.validlifetime: ", req.ValidLifetime)
			log.Println("req.MaxValidLifetime: ", req.MaxValidLifetime)
			if len(req.ValidLifetime) > 0 {
				if err != nil {
					log.Println("UpdateSubnetv4Pool, validLifetime error, ", err)
					return err
				}
				conf.Arguments.Dhcp4.Subnet4[k].ValidLifetime = json.Number(req.ValidLifetime)
			}
			if len(req.MaxValidLifetime) > 0 {
				if err != nil {
					log.Println("UpdateSubnetv4Pool, validLifetime error, ", err)
					return err
				}
				conf.Arguments.Dhcp4.Subnet4[k].MaxValidLifetime = json.Number(req.MaxValidLifetime)
			}

			conf.Arguments.Dhcp4.Subnet4[k].Pools = []Pool{}

			for _, p := range v.Pools {
				log.Println("in range pools, pool name: ", p.Pool, ", req.oldPool: ", req.Oldpool)
				if p.Pool == req.Oldpool {
					changeFlag = true
					log.Println("p.pool == req.pool")
					p.Pool = req.Pool

					var ops []*Option
					if ops, err = ConvertOptionsFromPb(req.Options); err != nil {
						log.Println("ConvertOptionsFromPb error: ", err)
						return err
					}
					p.OptionData = ops
				}
				conf.Arguments.Dhcp4.Subnet4[k].Pools = append(conf.Arguments.Dhcp4.Subnet4[k].Pools, p)
			}

			//log.Println("begin subnet pools")
			//log.Println(conf.Arguments.Dhcp4.Subnet4[k].Pools)
			//log.Println("end subne poolst")

			if changeFlag {
				err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
				if err != nil {
					return err
				}
			}

			return nil
		}
	}

	return nil
}
func (handler *KEAv4Handler) DeleteSubnetv4Pool(req pb.DeleteSubnetv4PoolReq) error {
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}
	changeFlag := false

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			tmp := []Pool{}

			for _, p := range conf.Arguments.Dhcp4.Subnet4[k].Pools {
				if p.Pool != req.Pool {
					tmp = append(tmp, p)
				}
				if p.Pool == req.Pool {
					changeFlag = true
				}
			}
			conf.Arguments.Dhcp4.Subnet4[k].Pools = tmp
			if changeFlag {
				err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	return nil
}
func (handler *KEAv4Handler) CreateSubnetv4Reservation(req pb.CreateSubnetv4ReservationReq) error {
	log.Println("into dhcp.go, CreateSubnetv4Reservation, req: ", req)
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}

	//找到subnet， todo 存取数据库前端和后端的subnet对应关系

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		//log.Print("in for loop, v.Id: ", v.Id, ", req.Id: ", req.Id)
		//log.Print("v.subnet: ", v.Subnet)
		//log.Print("req.Subnet: ", req.Subnet)
		if v.Subnet == req.Subnet {
			//log.Println("req.IpAddr: ", req.IpAddr)
			//log.Println("req.Duid: ", req.Duid)
			var rsv Reservation
			rsv.HwAddress = req.HwAddress
			rsv.Duid = req.Duid
			rsv.Hostname = req.Hostname
			rsv.IpAddress = req.IpAddr
			rsv.CircuitId = req.CircuitId
			rsv.ClientId = req.ClientId
			rsv.NextServer = req.NextServer

			//rsv.OptionData = req.Options
			var ops = []*Option{}
			for _, op := range req.Options {
				var o *Option
				o.AlwaysSend = op.AlwaysSend
				o.Code = op.Code
				o.CsvFormat = op.CsvFormat
				o.Data = op.Data
				o.Name = op.Name
				o.Space = op.Space

				ops = append(ops, o)
			}
			rsv.OptionData = ops

			log.Println("new rsv: ", rsv)
			conf.Arguments.Dhcp4.Subnet4[k].Reservations = append(conf.Arguments.Dhcp4.Subnet4[k].Reservations, rsv)
			//log.Println("new Reservations 0 hwadderss: ", conf.Arguments.Dhcp4.Subnet4[k].Reservations[0].HwAddress)
		}
	}

	log.Println("CreateSubnetv4Reservation begin subnet\n")
	log.Println(conf.Arguments.Dhcp4.Subnet4)
	log.Println("CreateSubnetv4Reservation end subnet\n")
	err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
	if err != nil {
		return err
	}
	return nil
	//return fmt.Errorf("subnet do not exists, error")
	//return nil
}

func (handler *KEAv4Handler) UpdateSubnetv4Reservation(req pb.UpdateSubnetv4ReservationReq) error {
	log.Println("into dhcp.go, UpdateSubnetv4Reservation, req: ", req)
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}

	//找到subnet， todo 存取数据库前端和后端的subnet对应关系

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		//log.Print("in for loop, v.Id: ", v.Id, ", req.Id: ", req.Id)
		//log.Print("v.subnet: ", v.Subnet)
		//log.Print("req.Subnet: ", req.Subnet)
		if v.Subnet == req.Subnet {
			log.Println("in if, req.IpAddr:[", req.IpAddr, "]")
			log.Println("in if, req.OldRsvIP:[", req.OldRsvIP, "]")
			conf.Arguments.Dhcp4.Subnet4[k].Reservations = []Reservation{}

			for _, confRsv := range v.Reservations {
				oldIP := strings.TrimSpace(confRsv.IpAddress)
				reqIP := strings.TrimSpace(req.OldRsvIP)

				if oldIP != reqIP {
					log.Println("in for,confRsv.IpAddr!=[", oldIP, "],req.OldRsvIP:[", reqIP, "]")

					//delete cur IPAddress
					conf.Arguments.Dhcp4.Subnet4[k].Reservations = append(conf.Arguments.Dhcp4.Subnet4[k].Reservations,
						confRsv)
				} else {
					log.Println("in for, confRsv.IpAddr == ", confRsv.IpAddress, ", req.OldRsvIP: ", req.OldRsvIP)
					newRsv := Reservation{
						IpAddress: req.IpAddr,
						Duid:      req.Duid,
						Hostname:  req.Hostname,
					}
					if len(req.Duid) > 0 {
						log.Println("set duid")
						newRsv.Duid = req.Duid
					}
					if len(req.Hostname) > 0 {
						log.Println("set duid")
						newRsv.Hostname = req.Hostname
					}
					//if len(req.) > 0 {
					//	log.Println("set duid")
					//	newRsv.Hostname = req.Hostname
					//}
					if len(req.NextServer) > 0 {
						log.Println("req.NextServer: ", req.NextServer)
						newRsv.NextServer = req.NextServer
					}

					var ops []*Option
					if ops, err = ConvertOptionsFromPb(req.Options); err != nil {
						log.Println("ConvertOptionsFromPb error: ", err)
						return err
					}
					newRsv.OptionData = ops

					conf.Arguments.Dhcp4.Subnet4[k].Reservations = append(conf.Arguments.Dhcp4.Subnet4[k].Reservations,
						newRsv)
				}
			}
			log.Println("tobe configed rsvs:", conf.Arguments.Dhcp4.Subnet4[k].Reservations)
		}
	}
	log.Println("CreateSubnetv4Reservation begin subnet\n")
	log.Println(conf.Arguments.Dhcp4.Subnet4)
	log.Println("CreateSubnetv4Reservation end subnet\n")

	err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
	if err != nil {
		return err
	}

	return nil
}
func (handler *KEAv4Handler) DeleteSubnetv4Reservation(req pb.DeleteSubnetv4ReservationReq) error {
	var conf ParseDhcpv4Config
	err := handler.getv4Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}
	changeFlag := false

	for k, v := range conf.Arguments.Dhcp4.Subnet4 {
		if v.Subnet == req.Subnet {
			tmp := []Reservation{}

			for _, p := range conf.Arguments.Dhcp4.Subnet4[k].Reservations {
				if p.IpAddress != req.IpAddr {
					tmp = append(tmp, p)
				} else {
					changeFlag = true
				}
			}
			conf.Arguments.Dhcp4.Subnet4[k].Reservations = tmp
			if changeFlag {
				err = handler.setDhcpv4Config(KEADHCPv4Service, &conf.Arguments)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	return nil
}
func (handler *KEAv4Handler) GetLeases(req pb.GetLeasesReq) (*pb.GetLeasesResp, error) {
	var ls []postgres.Lease4
	if err := handler.db.Where("subnet_id = ?", req.Subnetid).Find(&ls).Error; err != nil {
		return nil, err
	}
	var resp pb.GetLeasesResp
	for _, l := range ls {
		var address string
		address = strconv.Itoa(l.Address&0xff000000>>24&0xff) + "." + strconv.Itoa(l.Address&0x00ff0000>>16&0xff) + "." + strconv.Itoa(l.Address&0x0000ff00>>8&0xff) + "." + strconv.Itoa(l.Address&0x000000ff&0xff)
		tmp := pb.Lease{}
		tmp.IpAddress = address
		tmp.HwAddress = l.Hwaddr
		tmp.ValidLifetime = l.ValidLifetime
		tmp.Expire = l.Expire.Unix()
		resp.Leases = append(resp.Leases, &tmp)
	}
	return &resp, nil
}
func (handler *KEAv4Handler) Close() {
	handler.db.Close()

}
