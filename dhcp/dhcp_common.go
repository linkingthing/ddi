package dhcp

//in this package, func param values
// service: dhcp4, dhcp6, ctrl_agent, ddns

import (
	"fmt"
	"os/exec"
	"encoding/json"
	"github.com/ben-han-cn/cement/shell"
	"github.com/sirupsen/logrus"
)

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}


func isServiceRunning(service string) bool {
	command := "ps -eaf | grep " + service + " | grep -v grep"
	ret, err := cmd(command)
	if (err != nil) {
		return false
	}

	fmt.Println("ps -eaf ret: " + ret)
	return true
}


func getConfig(service string) (string,error) {

	postData := make(map[string]interface{})
	postData["command"] = "config-get"
	postData["service"] = []string{service}
	postStr, err := json.Marshal(postData)

	getCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://"+host+":"+port
	configJson, err := shell.Shell(getCmd)
	if(err != nil) {

		logrus.Error("config get error, service: "+service)
		return "",err
	}

	fmt.Println("configGet: " + configJson)
	return configJson, nil
}
