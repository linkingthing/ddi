package dhcpv6

import (
	"encoding/json"
	"os/exec"

	"github.com/linkingthing.com/ddi/dhcp"
)

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

func getConfigv6(service string, conf *Parse6Config) error {

	postData := map[string]interface{}{
		"command": "config-get",
		"service": []string{service},
	}
	postStr, _ := json.Marshal(postData)

	getCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://" + dhcp.DhcpHost + ":" + dhcp.DhcpPort + " 2>/dev/null"

	configJson, err := cmd(getCmd)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(string(configJson[2:len(configJson)-2])), conf)
	if err != nil {
		return err
	}

	return nil
}

func setConfigv6(service string, conf *DHCP6Conf) error {

	postData := map[string]interface{}{
		"command":   "config-set",
		"service":   []string{service},
		"arguments": &conf,
	}
	postStr, _ := json.Marshal(postData)

	curlCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://" + dhcp.DhcpHost + ":" + dhcp.DhcpPort + " 2>/dev/null"
	_, err := cmd(curlCmd)

	if err != nil {
		return err
	}
	return nil
}
