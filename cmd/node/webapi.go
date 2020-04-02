// test for node registeration on boot time
package node

import (
	"encoding/json"
	"fmt"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func RegisterNode(confPath string, role string) {
	//v1

	//utils.SetHostIPs(confPath) //set global vars from yaml conf

	var conf *config.VanguardConf
	conf = config.GetConfig(confPath)
	log.Println("in cmd/node, localhost ip: ", conf.Localhost.IP)

	hostname := conf.Localhost.Hostname
	hostIP := conf.Localhost.IP
	promHostIP := conf.Server.Prometheus.IP
	promHostPort := conf.Server.Prometheus.Port
	//send kafka msg to topic prom
	var PromInfo = utils.PromRole{
		Hostname: hostname,
		PromHost: promHostIP,
		PromPort: promHostPort,
		IP:       hostIP,
		Role:     role,
		State:    1, // 1 online 0 offline
		OnTime:   time.Now().Unix(),
		ParentIP: conf.Localhost.ParentIP,
	}
	PromJson, err := json.Marshal(PromInfo)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println("in cmd/node +++ PromJson")
	//fmt.Println(PromJson)
	//fmt.Println("--- PromJson")
	//go func() {
	key := "prom"
	value := PromJson
	msg := kafka.Message{
		Topic: utils.KafkaTopicProm,
		Key:   []byte(key),
		Value: []byte(value),
	}
	//log.Println("kafka.Message: ", msg)
	utils.ProduceProm(msg)
	//log.Println("produceProm msg ok")
}
