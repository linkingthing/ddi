package dhcp

import (
	"github.com/sirupsen/logrus"
	"fmt"
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
	Result int
	Arguments DHCPConfig
}
type DHCPConfig struct{
	Dhcp4 Dhcp4Config
}
type Dhcp4Config struct{
	Authoritative bool
	BootFileName string `json:boot-file-name`
	//ClientClasses map[string]interface{} `json:client-classes`
	ControlSocket controlSocket `json:control-socket`
	subnet4 []Subnet4Config
}
type controlSocket struct{
	SocketName string `json:socket-name`
	SocketType string `json:socket-type`
}

type Subnet4Config struct{

	Four4o6Interface string `json:4o6-interface`
	Four4o6InterfaceId string `json:4o6-interface-id`
	Four4o6Subnet string `json:4o6-subnet`
	Authoritative string `json:authoritative`
	OptionData []Subnet4OptionData `json:option-data`
	pools []Subnet4Pools
}
type Subnet4OptionData struct{
	AlwaysSend string `json:always-send`
	Code string `json:code`
	CsvFormat bool `json:csv-format`
	Data string
	Name string
	Space string
}
type Subnet4Pools struct{
	pool string
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

func CreateSubnet(service string, subnetIp string, subnetMask string ) error {

	configJson,err := getConfig(service)
	if(err != nil){
		return err
	}
	var configArr interface{}
	err2 := json.Unmarshal([]byte(string(configJson[2:len(configJson)-2])), &configArr)
	if(err2 != nil){
		fmt.Println("configJson unmarshall failed")
	}


	fmt.Println("service: " + service + ", configJson: " + configJson)

	return nil
}

