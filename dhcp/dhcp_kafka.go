package dhcp

import (
	"github.com/segmentio/kafka-go"
	"context"
	"fmt"
)

var(
	kafkaServer = "localhost:9092"
	dhcpTopic = "test"
)

func produce(msg kafka.Message){
	fmt.Printf("into produce\n")
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})

	w.WriteMessages(context.Background(), msg)
}


func consumer(){

	r := kafka.NewReader(kafka.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic: dhcpTopic,
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: key: %s = value: %s\n", m.Offset, string(m.Key), string(m.Value))

		//todo
	}

	r.Close()
}
