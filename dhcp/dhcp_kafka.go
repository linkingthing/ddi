package dhcp

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

const (
	KafkaServer = "localhost:9092"
	DhcpTopic   = "test"
	Dhcpv4Topic = "testv4"
	Dhcpv6Topic = "testv6"
)

func produce(msg kafka.Message) {
	fmt.Printf("into produce\n")
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{KafkaServer},
		Topic:   Dhcpv6Topic,
	})

	w.WriteMessages(context.Background(), msg)
}

func consumer() {

	r := kafka.NewReader(kafka.ReaderConfig{

		Brokers:     []string{KafkaServer},
		Topic:       Dhcpv6Topic,
		StartOffset: 34,
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
