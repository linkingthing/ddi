package dhcp

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"

	"github.com/linkingthing/ddi/pb"
)

type ParseDhcpv6Config struct {
	Result    json.Number
	Arguments DHCPv6Conf
}
type DHCPv6Conf struct {
	Dhcp6 Dhcpv6Config
}
type Dhcpv6Config struct {
	Authoritative bool   `json:"authoritative,omitempty"`
	BootFileName  string `json:"boot-file-name,omitempty"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket  `json:"control-socket,omitempty"`
	OptionData    []Option       `json:"option-data,omitempty"`
	Subnet6       []SubnetConfig `json:"subnet6,omitempty"`

	//T1Percent json.Number `json:"t1-percent"`
	//T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime,omitempty"`
}

type KEAv6Handler struct {
	db           *gorm.DB
	mu           sync.Mutex
	ver          string
	ConfigPath   string
	MainConfName string
}

//func NewKEAv6Handler(ver string, ConfPath string, addr string) *KEAv6Handler {
//
//	instance := &KEAv6Handler{ver: ver, ConfigPath: ConfPath}
//
//	return instance
//}
func NewKEAv6Handler(ver string, ConfPath string, addr string) *KEAv6Handler {
	instance := &KEAv6Handler{ver: ver, ConfigPath: ConfPath}
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

func (handler *KEAv6Handler) GetDhcpv6Config(service string, conf *ParseDhcpv6Config) error {

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
func (handler *KEAv6Handler) getv6Config(conf *ParseDhcpv6Config) error {
	if len(KeaDhcpv4Conf) == 0 {
		log.Print("KeaDhcpv6Conf is nil")
		err := handler.GetDhcpv6Config(KEADHCPv4Service, conf)
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

func (handler *KEAv6Handler) setDhcpv6Config(service string, conf *DHCPv6Conf) error {

	log.Print("dhcp/dhcpv6.go, into setDhcpv6Config()")
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

func (handler *KEAv6Handler) CreateSubnetv6Pool(req pb.CreateSubnetv6PoolReq) error {

	log.Println("into dhcp.go, CreateSubnetv4Pool, req: ", req)
	var conf ParseDhcpv6Config
	err := handler.getv6Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}
	//log.Println("begin conf\n")
	//log.Println(conf)
	//log.Println("end conf\n")

	//找到subnet， todo 存取数据库前端和后端的subnet对应关系

	for k, v := range conf.Arguments.Dhcp6.Subnet6 {
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

				conf.Arguments.Dhcp6.Subnet6[k].ValidLifetime = json.Number(req.ValidLifetime)
			}
			if len(req.MaxValidLifetime) > 0 {

				if err != nil {
					log.Println("CreateSubnetv4Pool, validLifetime error, ", err)
					return err
				}

				conf.Arguments.Dhcp6.Subnet6[k].MaxValidLifetime = json.Number(req.MaxValidLifetime)
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

				conf.Arguments.Dhcp6.Subnet6[k].Pools = append(conf.Arguments.Dhcp6.Subnet6[k].Pools, p)
			}
			//log.Println("begin subnet\n")
			//log.Println(conf.Arguments.Dhcp4)
			//log.Println("end subnet\n")

			err = handler.setDhcpv6Config(KEADHCPv4Service, &conf.Arguments)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("subnet do not exists, error")
}

func (handler *KEAv6Handler) CreateSubnetv6Reservation2(req pb.CreateSubnetv6ReservationReq) error {
	log.Println("into dhcp.go, CreateSubnetv4Reservation, req: ", req)
	var conf ParseDhcpv6Config
	err := handler.getv6Config(&conf)
	if err != nil {
		log.Println(err)
		return err
	}

	//找到subnet， todo 存取数据库前端和后端的subnet对应关系

	for k, v := range conf.Arguments.Dhcp6.Subnet6 {
		//log.Print("in for loop, v.Id: ", v.Id, ", req.Id: ", req.Id)
		//log.Print("v.subnet: ", v.Subnet)
		//log.Print("req.Subnet: ", req.Subnet)
		if v.Subnet == req.Subnet {
			//log.Println("req.IpAddr: ", req.IpAddr)
			//log.Println("req.Duid: ", req.Duid)
			var rsv Reservation
			//rsv.HwAddress = req.HwAddress
			rsv.Duid = req.Duid
			rsv.Hostname = req.Hostname
			rsv.IpAddress = req.IpAddr
			//rsv.CircuitId = req.CircuitId
			//rsv.ClientId = req.ClientId
			//rsv.NextServer = req.NextServer

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
			conf.Arguments.Dhcp6.Subnet6[k].Reservations = append(conf.Arguments.Dhcp6.Subnet6[k].Reservations, rsv)
			//log.Println("new Reservations 0 hwadderss: ", conf.Arguments.Dhcp4.Subnet4[k].Reservations[0].HwAddress)
		}
	}

	log.Println("CreateSubnetv4Reservation begin subnet\n")
	log.Println(conf.Arguments.Dhcp6.Subnet6)
	log.Println("CreateSubnetv4Reservation end subnet\n")
	err = handler.setDhcpv6Config(KEADHCPv4Service, &conf.Arguments)
	if err != nil {
		return err
	}
	return nil
	//return fmt.Errorf("subnet do not exists, error")
	//return nil
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
func (handler *KEAv6Handler) CreateSubnetv6Reservation(req pb.CreateSubnetv6ReservationReq) error {

	return nil
}

func (handler *KEAv6Handler) UpdateSubnetv6Pool(req pb.UpdateSubnetv6PoolReq) error {
	return nil
}

func (handler *KEAv6Handler) DeleteSubnetv6Pool(req pb.DeleteSubnetv6PoolReq) error {

	return nil
}
func (handler *KEAv6Handler) UpdateSubnetv6Reservation(req pb.UpdateSubnetv6ReservationReq) error {
	return nil
}

func (handler *KEAv6Handler) DeleteSubnetv6Reservation(req pb.DeleteSubnetv6ReservationReq) error {

	return nil
}
func (handler *KEAv6Handler) Close() {

}
