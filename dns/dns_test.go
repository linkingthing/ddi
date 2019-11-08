package dns

import (
	"context"
	"fmt"
	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/pb"
	kg "github.com/segmentio/kafka-go"
	"testing"
)

var (
	handler     DNSHandler
	kafkaServer = "localhost:9092"
	dhcpTopic   = "test"
	kafkaWriter *kg.Writer
	kafkaReader *kg.Reader
)

func init() {
	aclsSouth := make(map[string]ACL, 2)
	aclsSouth["acl001"] = ACL{ID: "acl001", Name: "liantongsouth", IpList: []string{"10.2.1.1", "10.2.1.2"}}
	aclsSouth["acl002"] = ACL{ID: "acl002", Name: "dianxinsouth", IpList: []string{"10.2.1.3", "10.2.1.4"}}
	aclsNorth := make(map[string]ACL, 2)
	aclsNorth["acl003"] = ACL{ID: "acl003", Name: "liantongnorth", IpList: []string{"10.2.1.5", "10.2.1.6"}}
	aclsNorth["acl004"] = ACL{ID: "acl004", Name: "dianxinnorth", IpList: []string{"10.2.1.7", "10.2.1.8"}}
	rrmap := make(map[string]RR, 2)
	rrmap["rr001"] = RR{"rr001", "wwww	A	10.2.1.1"}
	zonesSouth := make(map[string]Zone, 2)
	zonesSouth["zone001"] = Zone{"zone001", "baidu.com", "baidu.com.zone1", rrmap}
	zonesSouth["zone002"] = Zone{"zone002", "qq.com", "qq.com.zone1", rrmap}
	zonesNorth := make(map[string]Zone, 2)
	zonesNorth["zone001"] = Zone{"zone001", "baidu.com", "baidu.com.zone2", rrmap}
	zonesNorth["zone002"] = Zone{"zone002", "qq.com", "qq.com.zone2", rrmap}
	views := []View{View{"001", "SouthChinaView", aclsSouth, zonesSouth}, View{"002", "NorthChinaView", aclsNorth, zonesNorth}}
	p := &BindHandler{ConfContent: "", ConfigPath: "/root/bindtest", MainConfName: "named.conf", ViewList: views, FreeACLList: make(map[string]ACL, 2)}

	handler = p
	kafkaWriter = kg.NewWriter(kg.WriterConfig{
		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})
	kafkaReader = kg.NewReader(kg.ReaderConfig{

		Brokers: []string{kafkaServer},
		Topic:   dhcpTopic,
	})

}

func TestStartDNS(t *testing.T) {
	var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"
	dnsStartReq := pb.DNSStartReq{Config: config}
	data, err := proto.Marshal(&dnsStartReq)
	ut.Assert(t, err == nil, "StarDNS Marshal success!")
	postData := kg.Message{
		Key:   []byte("DNSStart"),
		Value: data,
	}
	err = kafkaWriter.WriteMessages(context.Background(), postData)
	ut.Assert(t, err == nil, "StarDNS WriteMessages success!")

	for {
		m, err := kafkaReader.ReadMessage(context.Background())
		var target pb.DNSStartReq
		s := string(m.Key)
		fmt.Println(s)
		if string(m.Key) == "DNSStart" {
			fmt.Println("is StartDNS!")
			err = proto.Unmarshal(m.Value, &target)
			ut.Assert(t, err == nil, "StarDNS Unmarshal success!")
			err = handler.StartDNS(target)
			ut.Assert(t, err == nil, "start successfully!")
			break
		} else {
			fmt.Println("is not StarDNS! ")
		}
	}

}

func TestStopDNS(t *testing.T) {
	err := handler.StopDNS()
	ut.Assert(t, err == nil, "stop successfully!")
}

func TestCreateACL(t *testing.T) {
	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}

	createACLReq := pb.CreateACLReq{
		ACLName: "southchina",
		ACLID:   "ACL001",
		IpList:  ipList}
	err := handler.CreateACL(createACLReq)
	ut.Assert(t, err == nil, "Create ACL successfully!")
}

func TestDeleteACL(t *testing.T) {
	deleteACLReq := pb.DeleteACLReq{ACLID: "ACL001"}
	err := handler.DeleteACL(deleteACLReq)
	ut.Assert(t, err == nil, "Delete ACL successfully!")
}

func TestCreateView(t *testing.T) {
	TestCreateACL(t)

	createViewReq := pb.CreateViewReq{
		ViewName: "DianXinView",
		ViewID:   "viewID001",
		Priority: 1,
		ACLIDs:   []string{"ACL001"}}
	err := handler.CreateView(createViewReq)
	ut.Assert(t, err == nil, "Create View Success!")

}

func TestDeleteView(t *testing.T) {
	TestCreateView(t)

	delViewReq := pb.DeleteViewReq{ViewID: "viewID001"}
	err := handler.DeleteView(delViewReq)
	ut.Assert(t, err == nil, "Delete View Success!")

}

func TestCreateZone(t *testing.T) {
	TestCreateView(t)

	createZoneReq := pb.CreateZoneReq{ViewID: "viewID001", ZoneID: "zoneID001", ZoneName: "test1031.com", ZoneFileName: "test1031.com.zone"}
	err := handler.CreateZone(createZoneReq)
	ut.Assert(t, err == nil, "Create Zone Success!")
}

func TestDeleteZone(t *testing.T) {
	TestCreateZone(t)
	delZoneReq := pb.DeleteZoneReq{ViewID: "viewID001", ZoneID: "zoneID001"}
	err := handler.DeleteZone(delZoneReq)
	ut.Assert(t, err == nil, "Create Delete Zone Success!")
}
