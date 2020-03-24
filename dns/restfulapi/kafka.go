package restfulapi

import (
	"context"

	"github.com/linkingthing/ddi/utils"
	kg "github.com/segmentio/kafka-go"
)

var (
	kafkaServer       = "localhost:9092"
	dnsTopic          = "dns"
	dhcpv4Topic       = "dhpv4"
	dhcpv6Topic       = "dhpv6"
	kafkaWriterDhcpv4 *kg.Writer
	kafkaWriterDhcpv6 *kg.Writer
	kafkaWriter       *kg.Writer
)

func init() {
	kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{utils.KafkaServerProm},
		Topic:   dnsTopic,
	})
	kafkaWriterDhcpv4 = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{utils.KafkaServerProm},
		Topic:   dhcpv4Topic,
	})
	kafkaWriterDhcpv6 = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{utils.KafkaServerProm},
		Topic:   dhcpv6Topic,
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

func SendCmdDhcpv4(data []byte, cmd string) error {
	postData := kg.Message{
		Key:   []byte(cmd),
		Value: data,
	}
	if err := kafkaWriterDhcpv4.WriteMessages(context.Background(), postData); err != nil {
		return err
	}
	return nil
}
func SendCmdDhcpv6(data []byte, cmd string) error {
	postData := kg.Message{
		Key:   []byte(cmd),
		Value: data,
	}
	if err := kafkaWriterDhcpv6.WriteMessages(context.Background(), postData); err != nil {
		return err
	}
	return nil
}
