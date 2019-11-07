package dhcp

import (
	"testing"
	"github.com/linkingthing.com/ddi/pb"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"fmt"
	"time"
	ut "github.com/ben-han-cn/cement/unittest"
)

var handler = &KEAHandler{
	ConfigPath: 	DhcpConfigPath,
	MainConfName: 	Dhcp4ConfigFile,
}

func TestKafka(t *testing.T) {
	configFile 	:= DhcpConfigPath + Dhcp4ConfigFile
	msg 		:= &pb.DHCPStartReq{Service:"dhcp4", ConfigFile:configFile}
	pData, err 	:= json.Marshal(msg)

	postData := kafka.Message{
		Key: []byte("DHCPStart"),
		Value: pData,
	}
	if err != nil {
		panic(err)
	}

	produce(postData)

	fmt.Printf("kafka send data")
	time.Sleep(time.Second)
	consumer()
}


func TestStopDHCP(t *testing.T) {
	service := pb.DHCPStopReq{Service:"dhcp4"}
	err := handler.StopDHCP(service)
	ut.Assert(t, err == nil, "dhcp4 stop successfully!")

	service = pb.DHCPStopReq{Service:"dhcp6"}
	err = handler.StopDHCP(service)
	ut.Assert(t, err == nil, "dhcp6 stop successfully!")

	time.Sleep(2 * time.Second)
}

func TestStartDHCP(t *testing.T) {

	configFile := DhcpConfigPath + Dhcp4ConfigFile
	service := pb.DHCPStartReq{Service:"dhcp4", ConfigFile:configFile}
	err := handler.StartDHCP(service)
	ut.Assert(t, err == nil, "dhcp4 start successfully!")

	configFile = DhcpConfigPath + Dhcp6ConfigFile
	service = pb.DHCPStartReq{Service:"dhcp6", ConfigFile:configFile}
	err = handler.StartDHCP(service)
	ut.Assert(t, err == nil, "dhcp6 start successfully!")

	time.Sleep(2 * time.Second)
}

//func TestSubnet(t *testing.T) {
//
//	time.Sleep(2 * time.Second)
//	if err := deleteSubnet4("dhcp4", "192.166.1.0/24"); err != nil {
//		t.Error(err)
//	}
//
//
//	if err := CreateSubnet4("dhcp4", "192.166.1.0/24", "192.166.1.10-192.166.1.20"); err != nil {
//		t.Error(err)
//	}
//
//
//	if err := updateSubnet4("dhcp4", "192.166.1.0/24", "192.166.1.0-192.166.1.55"); err != nil {
//		t.Error(err)
//	}
//
//
//}