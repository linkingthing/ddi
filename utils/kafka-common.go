package utils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	KafkaTopicProm = "prom"
	//checkDuration   = 24 * time.Hour
	Hostname = "ip-55"
)

type PromRole struct {
	Hostname string `json:"hostname"`
	PromHost string `json:"promHost"`
	PromPort string `json:"promPort"`
	IP       string `json:"ip"`
	Role     string `json:"role"`     // 3 roles: Controller, Db, Kafka
	State    uint   `json:"state"`    // 1 online 0 offline
	HbTime   int64  `json:"hbTime"`   //timestamp of most recent heartbeat time
	OnTime   int64  `json:"onTime"`   //timestamp of the nearest online time
	ParentIP string `json:"parentIP"` //parent node ip in node management graph
}

var OnlinePromHosts = make(map[string]PromRole)
var OfflinePromHosts = make(map[string]PromRole)
var KafkaOffset int64 = 0
var KafkaOffsetFile = "/tmp/kafka-offset.txt" // store kafka offset num into this file

// produceProm node uses kafka to report it's alive state
func ProduceProm(msg kafka.Message) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{KafkaServerProm},
		Topic:   KafkaTopicProm,
	})

	w.WriteMessages(context.Background(), msg)
}

// consumerProm server get msg from kafka topic regularly, if not accept, turn the machine's state to offline
func ConsumerProm() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{KafkaServerProm},
		Topic:       KafkaTopicProm,
		StartOffset: kafka.LastOffset,
	})
	defer r.Close()

	//read kafkaoffset from KafkaOffsetFile and set it to KafkaOffset
	size, err := ioutil.ReadFile(KafkaOffsetFile)
	if err == nil {
		offset, err2 := strconv.Atoi(string(size))
		if err2 != nil {
			log.Println(err2)
		}
		KafkaOffset = int64(offset)
		r.SetOffset(KafkaOffset)
	}
	log.Println("kafka Offset: ", KafkaOffset)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("ConsumerProm:", err)
			break
		}
		log.Printf(", message at offset %d: key: %s, value: %s\n", m.Offset, string(m.Key),
			string(m.Value))

		if string(m.Key) == "prom" {
			var Role PromRole
			err := json.Unmarshal(m.Value, &Role)
			if err != nil {
				log.Println(err)
				return
			}
			//put Role struct into OnlinePromHosts map
			Role.OnTime = time.Now().Unix()
			OnlinePromHosts[Role.IP+Role.Role] = Role

			//log.Println("+++ OnlinePromHosts")
			//log.Println(OnlinePromHosts)
			//log.Println("--- OnlinePromHosts")
		}

		//store curOffset into KafkaOffsetFile
		curOffset := r.Stats().Offset
		if curOffset > KafkaOffset {
			KafkaOffset = curOffset
			byteOffset := []byte(strconv.Itoa(int(curOffset)))
			err = ioutil.WriteFile(KafkaOffsetFile, byteOffset, 0644)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
