package dhcp

import (
	"encoding/json"
	"log"
	"os/exec"
)

func cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}

func getDhcpv4Config(service string, conf *ParseDhcpv4Config) error {

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

	err = json.Unmarshal([]byte(string(configJson[2:len(configJson)-2])), conf)
	if err != nil {
		return err
	}

	return nil
}

func setDhcpv4Config(service string, conf *DHCPv4Conf) error {

	postData := map[string]interface{}{
		"command":   "config-set",
		"service":   []string{service},
		"arguments": &conf,
	}
	postStr, _ := json.Marshal(postData)

	curlCmd := "curl -X POST -H \"Content-Type: application/json\" -d '" +
		string(postStr) + "' http://" + DhcpHost + ":" + DhcpPort + " 2>/dev/null"
	_, err := cmd(curlCmd)

	log.Print(curlCmd)

	if err != nil {
		return err
	}
	return nil
}
