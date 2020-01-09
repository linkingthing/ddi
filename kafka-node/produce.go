package main

import (
	"encoding/json"
	"fmt"
	"github.com/linkingthing/ddi/utils"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func main() {

	var PromInfo = utils.PromRole{
		Hostname: utils.Hostname,
		PromHost: utils.KafkaServerProm,
		PromPort: utils.PromMetricsPort,
		IP:       utils.HostIP,
		Role:     utils.RoleController,
		State:    1,
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
	utils.ProduceProm(msg)

	var PromInfo2 = utils.PromRole{
		Hostname: utils.Hostname,
		PromHost: utils.KafkaServerProm,
		PromPort: utils.PromMetricsPort,
		IP:       "10.0.0.23",
		Role:     utils.RoleController,
		State:    1,
		OnTime:   time.Now().Unix(),
	}
	PromJson2, err := json.Marshal(PromInfo2)
	if err != nil {
		fmt.Println(err)
		return
	}
	key2 := "prom"
	value2 := PromJson2
	msg2 := kafka.Message{
		Topic: utils.KafkaTopicProm,
		Key:   []byte(key2),
		Value: []byte(value2),
	}
	utils.ProduceProm(msg2)

	time.Sleep(10 * time.Second)
	log.Println("--- time sleep ok")
	//}()
}
