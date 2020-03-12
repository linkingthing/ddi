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

	PromServer = conf.Server.Prometheus.IP
	if conf.Localhost.IP != PromServer {
		PromLocalInstance = conf.Localhost.IP + ":" + PromLocalPort
	}
	KafkaServerProm = conf.Server.Kafka.Host + ":" + conf.Server.Kafka.Port
	log.Println("in common.go, utils.promServer ip: ", PromServer)
	log.Println("in common.go, utils.KafkaServerProm ip: ", KafkaServerProm)
	log.Println("in common.go, utils.PromLocalInstance ip: ", PromLocalInstance)

}

func Cmd(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	out, err := cmd.CombinedOutput()
	result := string(out)
	return result, err
}
