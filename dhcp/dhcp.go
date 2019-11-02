package dhcp

import (
	"github.com/sirupsen/logrus"
	"fmt"
)

const (
	host = "10.0.0.15"
	port = "8081"
)

type DHCPConfig struct{
	ver map[string]interface{}
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

	isRunning := isServiceRunning("kea-" + service)
	if(isRunning){
		startCmd := "keactrl stop -s kea-" + service + " >/dev/null 2>&1 &"
		_, err := cmd(startCmd);
		if err != nil {
			logrus.Error("keactrl stop -s kea-" + service + " failed")
			return err
		}
	}
	logrus.Error("not running, no need to stop service: kea-" + service)
	return nil
}

//service: dhcp4, dhcp6, ctrl_agent, ddns
func CreateSubnet(service string, subnetIp string, subnetMask string ) error {

	configJson,err := getConfig(service)
	if(err != nil){
		return err
	}

	fmt.Println("service: " + service + ", configJson: " + configJson)

	return nil
}

