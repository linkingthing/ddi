package main

import (
	"context"
	"fmt"
	"strconv"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing.com/ddi/dhcp"
	"github.com/linkingthing.com/ddi/pb"
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
		case dhcp.IntfStartDHCP:
			var data pb.DHCPStartReq
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handler.StartDHCP(pb.DHCPStartReq{Service: data.Service})

		case dhcp.IntfStopDHCP:
			var data pb.DHCPStopReq
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handler.StopDHCP(pb.DHCPStopReq{Service: data.Service})

		case dhcp.IntfCreateSubnet:
			var data pb.CreateSubnetReq
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				fmt.Printf("unmarshal error, m.key: %s\n", m.Key)
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}

			fmt.Printf("begin to call createsubnet, m.value: %s\n", string(m.Value))

			handler.CreateSubnet(pb.CreateSubnetReq{Service: data.Service, SubnetName: data.SubnetName, Pools: data.Pools})

			time.Sleep(10 * time.Second)
		case dhcp.IntfUpdateSubnet:
			var data pb.UpdateSubnetReq
			err := proto.Unmarshal(m.Value, &data)
			if err != nil {
				logrus.Error("unmarshal error, m.Value: " + string(m.Value))
				continue
			}
			handler.UpdateSubnet4(pb.UpdateSubnetReq{Service: data.Service, SubnetName: data.SubnetName, Pools: data.Pools})

		default:
			logrus.Error("kafka message unknown, m.key: " + string(m.Key))
		}

	}

}

func main() {
	consumer()
}