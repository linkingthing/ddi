package dns

import (
	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/linkingthing/ddi/pb"
	"testing"
)

var handler DNSHandler

func init() {
	aclsSouth := make(map[string]ACL, 2)
	aclsSouth["acl001"] = ACL{ID: "acl001", Name: "liantongsouth", IpList: []string{"10.2.1.1", "10.2.1.2"}}
	aclsSouth["acl002"] = ACL{ID: "acl002", Name: "dianxinsouth", IpList: []string{"10.2.1.3", "10.2.1.4"}}
	aclsNorth := make(map[string]ACL, 2)
	aclsNorth["acl003"] = ACL{ID: "acl003", Name: "liantongnorth", IpList: []string{"10.2.1.5", "10.2.1.6"}}
	aclsNorth["acl004"] = ACL{ID: "acl004", Name: "dianxinnorth", IpList: []string{"10.2.1.7", "10.2.1.8"}}
	rrmap := make(map[string]RR, 2)
	rrmap["rr001"] = RR{"rr001", "test"}
	zones := make(map[string]Zone, 2)
	zones["zone001"] = Zone{"zone001", "baidu.com", "baidu.com.zone1", rrmap}
	views := []View{View{"001", "SouthChinaView", aclsSouth, zones}, View{"002", "NorthChinaView", aclsNorth, zones}}
	p := &BindHandler{ConfContent: "", ConfigPath: "/root/bindtest", MainConfName: "named.conf", ViewList: views, FreeACLList: make(map[string]ACL, 2)}
	handler = p
}

func TestStartDNS(t *testing.T) {
	/*var handler DNSHandler
	aclsSouth := make(map[string]ACL, 2)
	aclsSouth["acl001"] = ACL{ID: "acl001", Name: "liantongsouth", IpList: []string{"10.2.1.1", "10.2.1.2"}}
	aclsSouth["acl002"] = ACL{ID: "acl002", Name: "dianxinsouth", IpList: []string{"10.2.1.3", "10.2.1.4"}}
	aclsNorth := make(map[string]ACL, 2)
	aclsNorth["acl003"] = ACL{ID: "acl003", Name: "liantongnorth", IpList: []string{"10.2.1.5", "10.2.1.6"}}
	aclsNorth["acl004"] = ACL{ID: "acl004", Name: "dianxinnorth", IpList: []string{"10.2.1.7", "10.2.1.8"}}
	rrmap := make(map[string]RR, 2)
	rrmap["rr001"] = RR{"rr001", "test"}
	zones := make(map[string]Zone, 2)
	zones["zone001"] = Zone{"zone001", "baidu.com", "baidu.com.zone1", rrmap}
	views := []View{View{"001", "SouthChinaView", aclsSouth, zones}, View{"002", "NorthChinaView", aclsNorth, zones}}
	p := &BindHandler{ConfContent: "", ConfigPath: "/root/bindtest", MainConfName: "named.conf", ViewList: views, FreeACLList: make(map[string]ACL, 2)}
	handler = p*/
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
	/*var handler DNSHandler
	p := &BindHandler{
		ConfigPath:   "/root/bindtest",
		MainConfName: "named.conf",
	}
	handler = p*/
	err := handler.StopDNS()
	ut.Assert(t, err == nil, "stop successfully!")
}

func TestCreateACL(t *testing.T) {
	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}

	/*var handler DNSHandler
	var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"

	p := &BindHandler{ConfContent: config, ConfigPath: "/root/bindtest", MainConfName: "named.conf", ViewList: []View{}, FreeACLList: make(map[string]ACL)}
	handler = p*/
	createACLReq := pb.CreateACLReq{
		ACLName: "southchina",
		ACLID:   "ACL001",
		IpList:  ipList}
	err := handler.CreateACL(createACLReq)
	ut.Assert(t, err == nil, "Create ACL successfully!")
}

func TestDeleteACL(t *testing.T) {
	/*var handler DNSHandler
	aclsSouth := make(map[string]ACL, 2)
	aclsSouth["acl001"] = ACL{ID: "acl001", Name: "liantongsouth", IpList: []string{"10.2.1.1", "10.2.1.2"}}
	aclsSouth["acl002"] = ACL{ID: "acl002", Name: "dianxinsouth", IpList: []string{"10.2.1.3", "10.2.1.4"}}
	aclsNorth := make(map[string]ACL, 2)
	aclsNorth["acl003"] = ACL{ID: "acl003", Name: "liantongnorth", IpList: []string{"10.2.1.5", "10.2.1.6"}}
	aclsNorth["acl004"] = ACL{ID: "acl004", Name: "dianxinnorth", IpList: []string{"10.2.1.7", "10.2.1.8"}}
	rrmap := make(map[string]RR, 2)
	rrmap["rr001"] = RR{"rr001", "test"}
	zones := make(map[string]Zone, 2)
	zones["zone001"] = Zone{"zone001", "baidu.com", "baidu.com.zone1", rrmap}
	views := []View{View{"001", "SouthChinaView", aclsSouth, zones}, View{"002", "NorthChinaView", aclsNorth, zones}}
	var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"
	acls := make(map[string]ACL, 2)
	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}
	oneAcl := ACL{"ACL001", "southchina", ipList}
	acls["ACL001"] = oneAcl

	p := &BindHandler{ConfContent: config, ConfigPath: "/root/bindtest", MainConfName: "named.conf", ViewList: views, FreeACLList: acls}
	handler = p*/

	deleteACLReq := pb.DeleteACLReq{ACLID: "ACL001"}
	err := handler.DeleteACL(deleteACLReq)
	ut.Assert(t, err == nil, "Delete ACL successfully!")
}
