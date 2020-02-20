// test for node registeration on boot time
package main

import (
	"encoding/json"
	"fmt"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func main() {
	//v1
	var conf *config.VanguardConf
	conf = config.GetConfig()
	log.Println("in agent.go, localhost ip: ", conf.Localhost.IP)

	//send kafka msg to topic prom
	var PromInfo = utils.PromRole{
		Hostname: utils.Hostname,
		PromHost: utils.KafkaServerProm,
		PromPort: utils.PromMetricsPort,
		IP:       utils.HostIP,
		Role:     utils.RoleController,
		State:    1, // 1 online 0 offline
		OnTime:   time.Now().Unix(),
	}
	PromJson, err := json.Marshal(PromInfo)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("+++ PromJson")
	fmt.Println(PromJson)
	fmt.Println("--- PromJson")
	//go func() {
	key := "prom"
	value := PromJson
	msg := kafka.Message{
		Topic: utils.KafkaTopicProm,
		Key:   []byte(key),
		Value: []byte(value),
	}
	log.Println("kafka.Message: ", msg)
	utils.ProduceProm(msg)
	log.Println("produceProm msg ok")
}
