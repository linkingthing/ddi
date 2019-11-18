package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/pb"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func produce(msg kafka.Message) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{dhcp.KafkaServer},
		Topic:   dhcp.DhcpTopic,
	})

	w.WriteMessages(context.Background(), msg)
}

func consumer() {
	r := kafka.NewReader(kafka.ReaderConfig{

		Brokers: []string{dhcp.KafkaServer},
		Topic:   dhcp.DhcpTopic,
	})
	defer r.Close()

	var handler = &dhcp.KEAHandler{
		ConfigPath:   dhcp.DhcpConfigPath,
		MainConfName: dhcp.Dhcp4ConfigFile,
	}

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}

		fmt.Printf("message at offset %d: key: %s, value: %s\n", m.Offset, string(m.Key), string(m.Value))

		index, _ := strconv.ParseUint(string(m.Key), 0, 64)
		switch uint(index) {
		case dhcp.IntfStartDHCPv4:
			var data pb.StartDHCPv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handler.StartDHCPv4(pb.StartDHCPv4Req{})

		case dhcp.IntfStopDHCPv4:
			var data pb.StopDHCPv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handler.StopDHCPv4(pb.StopDHCPv4Req{})

		case dhcp.IntfCreateSubnetv4:
			var data pb.CreateSubnetv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				fmt.Printf("unmarshal error, m.key: %s\n", m.Key)
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}

			fmt.Printf("begin to call createsubnet, m.value: %s\n", string(m.Value))

			handler.CreateSubnetv4(pb.CreateSubnetv4Req{Subnet: data.Subnet, Pool: data.Pool})

			time.Sleep(10 * time.Second)
		case dhcp.IntfUpdateSubnetv4:
			var data pb.UpdateSubnetv4Req
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handler.UpdateSubnetv4(pb.UpdateSubnetv4Req{Subnet: data.Subnet, Pool: data.Pool})

		default:
			logrus.Error("kafka message unknown, m.key: " + string(m.Key))
		}

	}

}

func main() {
	consumer()
}
