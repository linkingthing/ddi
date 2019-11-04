package dhcp

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
)

const (
	host = "10.0.0.15"
	port = "8081"
)

//"calculate-tee-times": false,
//                "id": 1,
//                "match-client-id": true,
//                "next-server": "0.0.0.0",
//                "option-data": [{
//                    "always-send": false,
//                    "code": 3,
//                    "csv-format": true,
//                    "data": "192.0.2.1",
//                    "name": "routers",
//                    "space": "dhcp4"
//                }],
//                "pools": [{
//                    "option-data": [],
//                    "pool": "192.0.2.1-192.0.2.200"
//                }],

type ParseConfig struct{
	Result json.Number
	Arguments DHCPConfig
}
type DHCPConfig struct{
	Dhcp4 Dhcp4Config
}
type Dhcp4Config struct{
	Authoritative bool
	BootFileName string `"json:boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket `json:"control-socket"`
	OptionData []Subnet4OptionData `json:"option-data"`
	Subnet4 []Subnet4Config

	T1Percent json.Number `json:"t1-percent"`
	T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type ControlSocket struct{
	SocketName string `json:"socket-name"`
	SocketType string `json:"socket-type"`
}

type Subnet4Config struct{

	Four4o6Interface string `json:"4o6-interface"`
	Four4o6InterfaceId string `json:"4o6-interface-id"`
	Four4o6Subnet string `json:"4o6-subnet"`
	Authoritative bool `json:"authoritative"`
	CalculateTeeTimes string `"calculate-tee-times"`
	Id json.Number
	MatchClientId bool `json:"match-client-id"`
	NextServer string `json:"next-server"`
	OptionData []Subnet4OptionData `json:"option-data"`
	Pools []Subnet4Pools
	RebindTimer json.Number `json:"rebind-timer"`
	Relay interface{}
	RenewTimer json.Number `json:"renew-timer"`
	ReservationMode string `json:"reservation-mode"`
	Reservations []Subnet4Reservations
	Subnet string

	T1Percent json.Number `json:"t1-percent"`
	T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}

type Subnet4OptionData struct{
	AlwaysSend bool `json:"always-send"`
	Code json.Number
	CsvFormat bool `json:"csv-format"`
	Data string
	Name string
	Space string
}
type Subnet4Pools struct{
	OptionData []string `json:"option-data"`
	Pool string
}
type Subnet4Reservations struct{
	BootFileName string `json:"boot-file-name"`
	ClientClasses []interface{}
	ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid string
	Hostname string
	IpAddress string `json:"ip-address"`
	NextServer string `json:"next-server"`
	OptionData []Subnet4OptionData
	ServerHostname string `json:"server-hostname"`
}








//service: dhcp4, dhcp6, ctrl_agent, ddns
func StartDHCP(service string) error {
	isRunning := isServiceRunning("kea-" + service)
	if(isRunning == false){
		startCmd := "nohup keactrl start -s " + service + " >/dev/null 2>&1 &"
		_, err := cmd(startCmd);
		if err != nil {
			logrus.Error("keactrl start -s kea-" + service + " failed")
			return err
		}
	}else{
		logrus.Error("keactrl start -s kea-" + service + " failed")
	}
	return nil
}

//service: dhcp4, dhcp6, ctrl_agent, ddns
func StopDHCP(service string) error{

	startCmd := "keactrl stop -s kea-" + service + " >/dev/null 2>&1 &"
	_, err := cmd(startCmd);
	if err != nil {
		logrus.Error("keactrl stop " + service + " failed")
		return err
	}

	return nil
}

func CreateSubnet(service string, subnetName string, pools string ) error {

	configJson,err := getConfig(service)
	if(err != nil){
		return err
	}

	var configArr ParseConfig
	err = json.Unmarshal([]byte(string(configJson[2:len(configJson)-2])), &configArr)
	if(err != nil){
		return err
	}

	//subnet4Config := configArr.Arguments.Dhcp4.Subnet4



	return nil
}