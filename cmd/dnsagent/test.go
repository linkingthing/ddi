package main

import (
	//"fmt"
	"github.com/linkingthing/ddi/dns"
	"github.com/linkingthing/ddi/pb"
)

func main() {
	var handler dns.DNSHandler
	p := &dns.BindHandler{
		ConfigPath:   "/root/bindtest",
		MainConfName: "named.conf",
		ViewList:     make(map[int]dns.View),
		FreeACLList:  make(map[string]dns.ACL),
	}
	handler = p
	/*var config string = "options {\n\tdirectory \"/root/bindtest/\";\n\tpid-file \"named.pid\";\n\tallow-new-zones yes;\n\tallow-query {any;};\n};\n" +
		"view \"default\" {\n\tmatch-clients {\n\tany;\n\t};\n};\n" +
		"key \"rndc-key\" {\n\talgorithm hmac-sha256;\n\tsecret \"4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=\";\n};\n" +
		"controls {\n\tinet 127.0.0.1 port 953\n\tallow { 127.0.0.1; } keys { \"rndc-key\"; };\n};\n" +
		"include \"/root/bindtest/named.rfc1912.zones\";\n"
	dnsStartReq := pb.DNSStartReq{Config: config}
	handler.StartDNS(dnsStartReq)*/

	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}

	createACLReq := pb.CreateACLReq{
		ACLName: "SouthChina",
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
	createZoneReq := pb.CreateZoneReq{ViewID: "viewID001", ZoneID: "zoneID001", ZoneName: "test1031.com"}
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
