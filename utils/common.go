package utils

import (
	"log"
	"os/exec"

	"github.com/linkingthing/ddi/utils/config"
)

// get ip configs from yaml config file, and set global variables
func SetHostIPs(confPath string) {
	var conf *config.VanguardConf
	conf = config.GetConfig(confPath)

	PromServer = conf.Server.Prometheus.IP
	if conf.Localhost.IP != PromServer {
		PromLocalInstance = conf.Localhost.IP + ":" + PromLocalPort
	}
	KafkaServerProm = conf.Server.Kafka.Host + ":" + conf.Server.Kafka.Port

	DHCPGrpcServer = conf.Server.DHCPGrpc
	Dhcpv4AgentAddr = conf.Server.DHCPGrpc
	/*IsController = conf.Localhost.IsController
	IsDHCP = conf.Localhost.IsDHCP
	IsDNS = conf.Localhost.IsDNS*/
	log.Println("in common.go, confpath: ", confPath)
	log.Println("in common.go, conf.Server.Kafka.Host: ", conf.Server.Kafka.Host)
	log.Println("in common.go, conf.Server.Kafka.Port: ", conf.Server.Kafka.Port)
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
