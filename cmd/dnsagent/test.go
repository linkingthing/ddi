package main

import (
	//"fmt"
	"github.com/linkingthing/ddi/dns"
	"github.com/linkingthing/ddi/pb"
)

func main() {
	var handler dns.DNSHandler
	aclsSouth := make(map[string]dns.ACL, 2)
	aclsSouth["acl001"] = dns.ACL{ID: "acl001", Name: "liantongsouth", IpList: []string{"10.2.1.1", "10.2.1.2"}}
	aclsSouth["acl002"] = dns.ACL{ID: "acl002", Name: "dianxinsouth", IpList: []string{"10.2.1.3", "10.2.1.4"}}
	aclsNorth := make(map[string]dns.ACL, 2)
	aclsNorth["acl003"] = dns.ACL{ID: "acl003", Name: "liantongnorth", IpList: []string{"10.2.1.5", "10.2.1.6"}}
	aclsNorth["acl004"] = dns.ACL{ID: "acl004", Name: "dianxinnorth", IpList: []string{"10.2.1.7", "10.2.1.8"}}
	rrmap := make(map[string]dns.RR, 2)
	rrmap["rr001"] = dns.RR{"rr001", "test"}
	zones := make(map[string]dns.Zone, 2)
	zones["zone001"] = dns.Zone{"zone001", "baidu.com", "baidu.com.zone1", rrmap}
	zones["zone002"] = dns.Zone{"zone002", "qq.com", "baidu.com.zone2", rrmap}
	views := []dns.View{dns.View{"001", "SouthChinaView", aclsSouth, zones}, dns.View{"002", "NorthChinaView", aclsNorth, zones}}
	p := &dns.BindHandler{ConfContent: "", ConfigPath: "/root/bindtest", MainConfName: "named.conf", ViewList: views, FreeACLList: make(map[string]dns.ACL, 2)}
	handler = p
	var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"
	dnsStartReq := pb.DNSStartReq{Config: config}
	handler.StartDNS(dnsStartReq)

	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}

	createACLReq := pb.CreateACLReq{
		ACLName: "southchina",
		ACLID:   "ACL001",
		IpList:  ipList}
	handler.CreateACL(createACLReq)
	deleteACLReq := pb.DeleteACLReq{ACLID: "ACL001"}
	handler.DeleteACL(deleteACLReq)
	handler.CreateACL(createACLReq)

	createViewReq := pb.CreateViewReq{
		ViewName: "DianXinView",
		ViewID:   "viewID001",
		Priority: 1,
		ACLID:    "ACL001"}
	handler.CreateView(createViewReq)

	var deleteIPList = []string{"192.168.199.0/24", "192.168.198.0/24"}
	var newIPList = []string{"192.168.196.0/24", "192.168.197.0/24"}
	updateViewReq := pb.UpdateViewReq{ViewID: "viewID001", Priority: 2, ACLID: "ACL001", DeleteIPList: deleteIPList, NewIPList: newIPList}
	handler.UpdateView(updateViewReq)
	createZoneReq := pb.CreateZoneReq{ViewID: "viewID001", ZoneID: "zoneID001", ZoneName: "test1031.com", ZoneFileName: "test1031.com.zone"}
	handler.CreateZone(createZoneReq)
	createRRReq := pb.CreateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RrID: "rrid001", Rrdata: "wwww    A   192.168.179.116"}
	handler.CreateRR(createRRReq)
	updateRRReq := pb.UpdateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RrID: "rrid001", NewrrData: "wwww    A   192.168.179.117"}
	handler.UpdateRR(updateRRReq)
	deleteRRReq := pb.DeleteRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RrID: "rrid001"}
	handler.DeleteRR(deleteRRReq)
	deleteZoneReq := pb.DeleteZoneReq{ViewID: "viewID001", ZoneID: "zoneID001"}
	handler.DeleteZone(deleteZoneReq)
	deleteViewReq := pb.DeleteViewReq{ViewID: "viewID001"}
	handler.DeleteView(deleteViewReq)

	handler.StopDNS()
}
