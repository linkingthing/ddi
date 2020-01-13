package utils

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

var (
	KafkaServerProm = HostIP + ":" + KafkaPort
	KafkaPort       = "9092"
	KafkaTopicProm  = "prom"
	//checkDuration   = 24 * time.Hour
	Hostname        = "ip-15"
	HostIP          = "10.0.0.15"
	PromMetricsPort = "9100"
)

const (
	RoleController = iota
	RoleDB
	RoleKafka
)

type PromRole struct {
	Hostname string
	PromHost string
	PromPort string
	IP       string
	Role     uint  // 3 roles: Controller, Db, Kafka
	State    uint  // 1 online 0 offline
	OnTime   int64 //timestamp of the nearest online time
}

var OnlinePromHosts = make(map[string]PromRole)
var OfflinePromHosts = make(map[string]PromRole)
var KafkaOffset int64 = 0

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
	log.Println("kafkaoffset: ", KafkaOffset)
	r.SetOffset(KafkaOffset)

	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}

		//_, file, line, _ := runtime.Caller(1)
		log.Printf(", message at offset %d: key: %s, value: %s\n", m.Offset, string(m.Key),
			string(m.Value))

		KafkaOffset = int64(m.Offset) + 1
		err = r.SetOffset(KafkaOffset)
		if err != nil {
			log.Println("r.setoffset error", err)
		}
		log.Println("r.setoffset ok ", KafkaOffset)

		var Role PromRole

		if string(m.Key) == "prom" {
			err := json.Unmarshal(m.Value, &Role)
			if err != nil {
				log.Println(err)
				return
			}

			//put Role struct into OnlinePromHosts map
			Role.OnTime = time.Now().Unix()
			OnlinePromHosts[Role.IP] = Role

			log.Println("+++ OnlinePromHosts")
			log.Println(OnlinePromHosts)
			log.Println("--- OnlinePromHosts")
		} else {
			log.Println(", message at offset %d: key: %s, value: %s\n", m.Offset, string(m.Key),
				string(m.Value))
		}
	}
}
