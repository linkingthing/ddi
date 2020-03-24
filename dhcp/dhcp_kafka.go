package dhcp

import (
	"context"

	"log"

	"github.com/linkingthing/ddi/utils"
	kg "github.com/segmentio/kafka-go"
)

var (
	KafkaServer = utils.KafkaServerProm
	Dhcpv4Topic = "dhcpv4"
	Dhcpv6Topic = "dhcpv6"
)

func SendDhcpCmd(data []byte, cmd string) error {
	log.Println("into SendDhcpCmd(), cmd: ", cmd)

	//KafkaServer = utils.KafkaServerProm
	log.Println("after kafkaServer: ", KafkaServer)
	var DhcpkafkaWriter *kg.Writer
	DhcpkafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{KafkaServer},
		Topic:   Dhcpv4Topic,
	})

	postData := kg.Message{
		Key:   []byte(cmd),
		Value: data,
	}
	if err := DhcpkafkaWriter.WriteMessages(context.Background(), postData); err != nil {
		return err
	}
	return nil
}

func produce(msg kg.Message) {
	//log.Printf("into produce\n")
	w := kg.NewWriter(kg.WriterConfig{
		Brokers: []string{KafkaServer},
		Topic:   Dhcpv4Topic,
	})

	w.WriteMessages(context.Background(), msg)
}

func consumer() {

	r := kg.NewReader(kg.ReaderConfig{

		Brokers:     []string{KafkaServer},
		Topic:       Dhcpv4Topic,
		StartOffset: 34,
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		log.Printf("message at offset %d: key: %s = value: %s\n", m.Offset, string(m.Key), string(m.Value))

		//todo
	}

	r.Close()
}
