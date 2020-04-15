package dhcp

import (
	"context"

	"log"

	"io/ioutil"
	"strconv"

	"github.com/linkingthing/ddi/utils"
	kg "github.com/segmentio/kafka-go"
)

var (
	KafkaServer = utils.KafkaServerProm
	Dhcpv4Topic = "dhcpv4"
	Dhcpv6Topic = "dhcpv6"
)
var KafkaOffsetFileDhcpv4 = "/tmp/kafka-offset-dhcpv4.txt" // store kafka offset num into this file
var KafkaOffsetDhcpv4 int64 = 0

func SendDhcpCmd(data []byte, cmd string) error {
	log.Println("into SendDhcpCmd(), cmd: ", cmd)

	KafkaServer = utils.KafkaServerProm
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

func SendDhcpv6Cmd(data []byte, cmd string) error {
	log.Println("into SendDhcpv6Cmd(), cmd: ", cmd)

	KafkaServer = utils.KafkaServerProm
	log.Println("after kafkaServer: ", KafkaServer)
	var DhcpkafkaWriter *kg.Writer
	DhcpkafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{KafkaServer},
		Topic:   Dhcpv6Topic,
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

		Brokers: []string{KafkaServer},
		Topic:   Dhcpv4Topic,
		//StartOffset: 34,
	})
	var KafkaOffsetDhcpv4 int64
	size, err := ioutil.ReadFile(KafkaOffsetFileDhcpv4)
	if err == nil {
		offset, err2 := strconv.Atoi(string(size))
		if err2 != nil {
			log.Println(err2)
		}
		KafkaOffsetDhcpv4 = int64(offset)
		r.SetOffset(KafkaOffsetDhcpv4)
	}
	log.Println("kafka Offset: ", KafkaOffsetDhcpv4)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		log.Printf("message at offset %d: key: %s = value: %s\n", m.Offset, string(m.Key), string(m.Value))

		//todo

		//store curOffset into KafkaOffsetFile
		curOffset := r.Stats().Offset
		if curOffset > KafkaOffsetDhcpv4 {
			KafkaOffsetDhcpv4 = curOffset
			byteOffset := []byte(strconv.Itoa(int(curOffset)))
			err = ioutil.WriteFile(KafkaOffsetFileDhcpv4, byteOffset, 0644)
			if err != nil {
				log.Println(err)
			}
		}
	}

	r.Close()
}
