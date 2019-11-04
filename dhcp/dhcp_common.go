package dhcp

//in this package, func param values
// service: dhcp4, dhcp6, ctrl_agent, ddns

import (
	"os/exec"
	"encoding/json"
)

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}


func isServiceRunning(service string) bool {

	return true
}



func getConfig(service string) (string,error) {
	postData := map[string]interface{}{
		"command" : "config-get",
		"service" : []string{service},
	}
	postStr, _ := json.Marshal(postData)

	getCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://"+host+":"+port + " 2>/dev/null"
	configJson, err := cmd(getCmd)
	if(err != nil) {
		return "",err
	}
	return configJson, nil
}
