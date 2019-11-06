package dhcp

import (
	"github.com/segmentio/kafka-go"
	"context"
	"fmt"
)

var(
	kafkaServer = "localhost:9092"
	dhcpTopic = "dhcp"
)

func produce(){
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})

	w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("Key-A"),
			Value: []byte("Hello World!"),
		},
	)
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
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	r.Close()
}
