package dhcp

import (
	"testing"
	"time"

	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/linkingthing.com/ddi/pb"
)

var handler = &KEAHandler{
	ConfigPath:   DhcpConfigPath,
	MainConfName: Dhcp4ConfigFile,
}

//func TestKafka(t *testing.T) {
//
//	conf := &ParseConfig{}
//	err := getConfig("dhcp4", conf)
//	fmt.Print(conf)
//
//	//configFile := DhcpConfigPath + Dhcp4ConfigFile
//	//info := &pb.DHCPStartReq{Service: "dhcp4", ConfigFile: configFile}
//	d1 := &pb.Subnet4OptionData{}
//	d2 := &pb.Subnet4Pools{}
//	d2.OptionData = []*pb.Subnet4OptionData{d1}
//	d2.Pool = "192.166.1.10-192.166.1.33"
//
//	info := &pb.CreateSubnet4Req{Service: "dhcp4", SubnetName: "192.166.1.0/24", Pools: []*pb.Subnet4Pools{d2}}
//	data, err := proto.Marshal(info)
//	if err != nil {
//		panic(err)
//	}
//
//	msg := kafka.Message{
//		Key:   []byte(strconv.Itoa(IntfCreateSubnet4)),
//		Value: data,
//	}
//	produce(msg)
//
//	fmt.Printf("kafka send data\n")
//	time.Sleep(5 * time.Second)
//	//consumer()
//}

func TestStopDHCP(t *testing.T) {
	service := pb.DHCPStopReq{Service: "dhcp4"}
	err := handler.StopDHCP(service)
	ut.Assert(t, err == nil, "dhcp4 stop successfully!")

	service = pb.DHCPStopReq{Service: "dhcp6"}
	err = handler.StopDHCP(service)
	ut.Assert(t, err == nil, "dhcp6 stop successfully!")

	time.Sleep(2 * time.Second)
}

func TestStartDHCP(t *testing.T) {

	configFile := DhcpConfigPath + Dhcp4ConfigFile
	service := pb.DHCPStartReq{Service: "dhcp4", ConfigFile: configFile}
	err := handler.StartDHCP(service)
	ut.Assert(t, err == nil, "dhcp4 start successfully!")

	configFile = DhcpConfigPath + Dhcp6ConfigFile
	service = pb.DHCPStartReq{Service: "dhcp6", ConfigFile: configFile}
	err = handler.StartDHCP(service)
	ut.Assert(t, err == nil, "dhcp6 start successfully!")

	time.Sleep(2 * time.Second)
}

func TestDeleteSubnet(t *testing.T) {
	time.Sleep(time.Second)

	req := pb.DeleteSubnetReq{Service: "dhcp4", SubnetName: "192.166.1.0/24"}
	err := handler.DeleteSubnet4(req)
	ut.Assert(t, err == nil, "delete Subnet 192.166.1.0/24 successfully!")
}

func TestCreateSubnet(t *testing.T) {

	time.Sleep(time.Second)
	d1 := &pb.SubnetOptionData{}

	d2 := &pb.SubnetPools{}
	d2.OptionData = []*pb.SubnetOptionData{d1}
	d2.Pool = "192.166.1.10-192.166.1.40"

	req := pb.CreateSubnetReq{Service: "dhcp4", SubnetName: "192.166.1.0/24",
		Pools: []*pb.SubnetPools{d2},
	}

	err := handler.CreateSubnet(req)
	ut.Assert(t, err == nil, "Create Subnet 192.166.1.0/24 successfully!")

}

func TestUpdateSubnet(t *testing.T) {
	time.Sleep(time.Second)

	d1 := &pb.SubnetOptionData{}
	d2 := &pb.SubnetPools{}
	d2.OptionData = []*pb.SubnetOptionData{d1}
	d2.Pool = "192.166.1.10-192.166.1.33"

	req := pb.UpdateSubnetReq{Service: "dhcp4", SubnetName: "192.166.1.0/24", Pools: []*pb.SubnetPools{d2}}
	err := handler.UpdateSubnet4(req)
	ut.Assert(t, err == nil, "Update Subnet 192.166.1.0/24 successfully!")
}
