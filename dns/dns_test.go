package dns

import (
	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/linkingthing/ddi/pb"
	"testing"
)

func TestStartDNS(t *testing.T) {
	var handler DNSHandler
	p := &BindHandler{
		ConfigPath:   "/root/bindtest",
		MainConfName: "named.conf",
	}
	handler = p
	var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"
	dnsStartReq := pb.DNSStartReq{Config: config}
	err := handler.StartDNS(dnsStartReq)
	ut.Assert(t, err == nil, "start successfully!")
}

func TestStopDNS(t *testing.T) {
	var handler DNSHandler
	p := &BindHandler{
		ConfigPath:   "/root/bindtest",
		MainConfName: "named.conf",
	}
	handler = p
	err := handler.StopDNS()
	ut.Assert(t, err == nil, "stop successfully!")
}

func TestCreateACL(t *testing.T) {
	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}

	var handler DNSHandler
	var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"

	p := &BindHandler{ConfContent: config, ConfigPath: "/root/bindtest", MainConfName: "named.conf", ViewList: make(map[int]View), FreeACLList: make(map[string]ACL)}
	handler = p
	createACLReq := pb.CreateACLReq{
		ACLName: "SouthChina",
		ACLID:   "ACL001",
		IpList:  ipList}
	err := handler.CreateACL(createACLReq)
	ut.Assert(t, err == nil, "Create ACL successfully!")
}
