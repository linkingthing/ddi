package dhcp

import (
	"testing"
	"time"

	"log"

	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/linkingthing/ddi/pb"
)

var handlerv4 = &KEAv4Handler{
	ver:          "dhcp4",
	ConfigPath:   DhcpConfigPath,
	MainConfName: Dhcp4ConfigFile,
}
var handlerv6 = &KEAv6Handler{
	ver:          "dhcp6",
	ConfigPath:   DhcpConfigPath,
	MainConfName: Dhcp6ConfigFile,
}

//func TestKafka(t *testing.T) {
//
//	conf := &ParseConfig{}
//	err := getConfig("dhcp4", conf)
//	fmt.Print(conf)
//
//	configFile := DhcpConfigPath + Dhcp4ConfigFile
//	info := &pb.DHCPStartReq{Service: "dhcp4", ConfigFile: configFile}
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

//func TestStopDHCPv4(t *testing.T) {
//	service := pb.StopDHCPv4Req{}
//	err := handlerv4.StopDHCPv4(service)
//	ut.Assert(t, err == nil, "dhcp4 stop successfully!")
//
//	time.Sleep(2 * time.Second)
//}

//func TestStartDHCPv4(t *testing.T) {
//
//	configFile := DhcpConfigPath + Dhcp4ConfigFile
//	dhcpv4 := pb.StartDHCPv4Req{Config: configFile}
//	err := handlerv4.StartDHCPv4(dhcpv4)
//	ut.Assert(t, err == nil, "dhcp4 start successfully!")
//
//	time.Sleep(2 * time.Second)
//}

func TestDeleteSubnetv4(t *testing.T) {
	time.Sleep(time.Second)
	log.Print("begin to delete subnet v4")

	req := pb.DeleteSubnetv4Req{Subnet: "192.166.1.0/24"}
	err := handlerv4.DeleteSubnetv4(req)
	ut.Assert(t, err == nil, "delete Subnet 192.166.1.0/24 successfully!")
}

func TestCreateSubnetv4(t *testing.T) {
	log.Print("begin to create subnet v4")

	time.Sleep(time.Second)
	d1 := &pb.Option{}

	d2 := &pb.Pools{}
	d2.Options = []*pb.Option{d1}
	d2.Pool = "192.166.1.10-192.166.1.40"

	req := pb.CreateSubnetv4Req{Subnet: "192.166.1.0/24",
		Pool: []*pb.Pools{d2},
	}

	err := handlerv4.CreateSubnetv4(req)
	ut.Assert(t, err == nil, "Create Subnet 192.166.1.0/24 successfully!")

}

func TestUpdateSubnetv4(t *testing.T) {
	time.Sleep(time.Second)
	log.Print("begin to update subnet v4")

	d1 := &pb.Option{}
	d2 := &pb.Pools{}
	d2.Options = []*pb.Option{d1}
	d2.Pool = "192.166.1.10-192.166.1.33"

	req := pb.UpdateSubnetv4Req{Subnet: "192.166.1.0/24", Pool: []*pb.Pools{d2}}
	err := handlerv4.UpdateSubnetv4(req)
	ut.Assert(t, err == nil, "Update Subnet 192.166.1.0/24 successfully!")
}

func TestCreateSubnetv4Pool(t *testing.T) {
	log.Print("begin to create subnet v4 pool")

	//d1 := &pb.Option{
	//	AlwaysSend: true,
	//	Code:       1,
	//	CsvFormat:  false,
	//	Data:       "192.166.1.52",
	//	Name:       "name22",
	//	Space:      "dhcp4",
	//}

	p := pb.Pools{
		Options: []*pb.Option{},
		Pool:    "192.166.1.40-192.166.1.70",
	}

	req := pb.CreateSubnetv4PoolReq{Subnet: "192.166.1.0/24",
		Pool: []*pb.Pools{&p},
	}

	err := handlerv4.CreateSubnetv4Pool(req)
	ut.Assert(t, err == nil, "Create Subnet 192.166.1.0/24, pool: 192.166.1.40-70 successfully!")
}
