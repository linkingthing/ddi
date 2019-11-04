package dhcp

import (
	"github.com/sirupsen/logrus"
	"encoding/json"
	"fmt"
)

const (
	host = "10.0.0.15"
	port = "8081"
)

type ParseConfig struct{
	Result json.Number
	Arguments DHCPConfig
}
type DHCPConfig struct{
	Dhcp4 Dhcp4Config
}
type Dhcp4Config struct{
	Authoritative bool `json:"authoritative"`
	BootFileName string `json:"boot-file-name"`
	//ClientClasses map[string]interface{} `json:"client-classes"`
	ControlSocket ControlSocket `json:"control-socket"`
	OptionData []Subnet4OptionData `json:"option-data"`
	Subnet4 []Subnet4Config `json:"subnet4"`

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
	CalculateTeeTimes bool `json:"calculate-tee-times"`
	Id json.Number `json:"id"`
	MatchClientId bool `json:"match-client-id"`
	NextServer string `json:"next-server"`
	OptionData []Subnet4OptionData `json:"option-data"`
	Pools []Subnet4Pools `json:"pools"`
	RebindTimer json.Number `json:"rebind-timer"`
	Relay Subnet4Relay `json:"relay"`
	RenewTimer json.Number `json:"renew-timer"`
	ReservationMode string `json:"reservation-mode"`
	Reservations []Subnet4Reservations `json:"reservations"`
	Subnet string `json:"subnet"`

	T1Percent json.Number `json:"t1-percent"`
	T2Percent json.Number `json:"t2-percent"`
	ValidLifetime json.Number `json:"valid-lifetime"`
}
type Subnet4Relay struct{
	IpAddresses []string `json:"ip-addresses"`
}

type Subnet4OptionData struct{
	AlwaysSend bool `json:"always-send"`
	Code json.Number `json:"code"`
	CsvFormat bool `json:"csv-format"`
	Data string `json:"data"`
	Name string `json:"name"`
	Space string `json:"space"`
}
type Subnet4Pools struct{
	OptionData []Subnet4OptionData `json:"option-data"`
	Pool string `json:"pool"`
}
type Subnet4Reservations struct{
	BootFileName string `json:"boot-file-name"`
	ClientClasses []interface{} `json:"client-classes"`
	ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid string `json:"duid"`
	Hostname string `json:"hostname"`
	IpAddress string `json:"ip-address"`
	NextServer string `json:"next-server"`
	OptionData []Subnet4OptionData `json:"option-data"`
	ServerHostname string `json:"server-hostname"`
}








//service: dhcp4, dhcp6, ctrl_agent, ddns
func StartDHCP(service string) error {
	startCmd := "nohup keactrl start -s " + service + " >/dev/null 2>&1 &"
	_, err := cmd(startCmd);
	if err != nil {
		logrus.Error("keactrl start -s kea-" + service + " failed")
		return err
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

	newSubnet4 := Subnet4Config{
		ReservationMode:"all",
		Subnet: subnetName,
		Pools: []Subnet4Pools{
			{
				[]Subnet4OptionData{},
				pools,
			},
		},
	}
	configArr.Arguments.Dhcp4.Subnet4 = append(configArr.Arguments.Dhcp4.Subnet4, newSubnet4)
	newDhcp4Json,newDhcp4Err := json.Marshal(&configArr.Arguments)
	if(newDhcp4Err != nil) {
		return newDhcp4Err
	}

	fmt.Printf("new dhcp4 json: %s\n", newDhcp4Json)

	//fmt.Printf("new configArr: %+v\n", configArr)

	setErr := setConfig(service, newDhcp4Json)
	if setErr != nil {
		return setErr
	}
	return nil
}

func configSet(service string, newDhcp4Json []byte) error {


	return nil
}