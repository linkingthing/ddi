package restfulapi

import (
	"context"
	kg "github.com/segmentio/kafka-go"
)

var (
	kafkaServer = "localhost:9092"
	dnsTopic    = "dns"
	kafkaWriter *kg.Writer
)

func init() {
	kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dnsTopic,
	})
}

func SendCmd(data []byte, cmd string) error {
	postData := kg.Message{
		Key:   []byte(cmd),
		Value: data,
	}
	if err := kafkaWriter.WriteMessages(context.Background(), postData); err != nil {
		return err
	}
	return nil
}
