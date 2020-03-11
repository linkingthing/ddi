package utils

import (
	"github.com/linkingthing/ddi/utils/config"
	"log"
	"os/exec"
)

// get ip configs from yaml config file, and set global variables
func SetHostIPs() {
	var conf *config.VanguardConf
	conf = config.GetConfig()

	log.Println("in agent.go, cur utils.promServer ip: ", PromServer)
	PromServer = conf.Server.Prometheus.IP
	log.Println("in agent.go, utils.promServer ip: ", PromServer)
}

func Cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}
