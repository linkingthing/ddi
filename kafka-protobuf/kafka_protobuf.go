package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func produce(msg kafka.Message) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{utils.KafkaServer},
		Topic:   utils.Dhcpv4Topic,
	})

	w.WriteMessages(context.Background(), msg)
}

func consumer() {
	r := kafka.NewReader(kafka.ReaderConfig{

		Brokers: []string{utils.KafkaServer},
		Topic:   utils.Dhcpv4Topic,
	})
	defer r.Close()

	var handlerv4 = &dhcp.KEAv4Handler{
		ConfigPath:   dhcp.DhcpConfigPath,
		MainConfName: dhcp.Dhcp4ConfigFile,
	}

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}

		log.Printf("kafka_protobuf.go, message at offset %d: key: %s, value: %s\n", m.Offset, string(m.Key), string(m.Value))

		index, _ := strconv.ParseUint(string(m.Key), 0, 64)
		switch uint(index) {
		case dhcp.IntfStartDHCPv4:
			var data pb.StartDHCPv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handlerv4.StartDHCPv4(pb.StartDHCPv4Req{})

		case dhcp.IntfStopDHCPv4:
			var data pb.StopDHCPv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handlerv4.StopDHCPv4(pb.StopDHCPv4Req{})

		case dhcp.IntfCreateSubnetv4:
			var data pb.CreateSubnetv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				log.Printf("unmarshal error, m.key: %s\n", m.Key)
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}

			log.Printf("begin to call createsubnet, m.value: %s\n", string(m.Value))

			handlerv4.CreateSubnetv4(pb.CreateSubnetv4Req{Subnet: data.Subnet, Pool: data.Pool})

			time.Sleep(10 * time.Second)
		case dhcp.IntfUpdateSubnetv4:
			var data pb.UpdateSubnetv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handlerv4.UpdateSubnetv4(pb.UpdateSubnetv4Req{Subnet: data.Subnet, Pool: data.Pool})

		default:
			logrus.Error("kafka message unknown, m.key: " + string(m.Key))
		}

	}

}

func main() {
	consumer()
}
