package main

import (
	"github.com/linkingthing/ddi/dns"
	"github.com/linkingthing/ddi/pb"
)

func main() {
	//var handler dns.DNSHandler
	handler := dns.NewBindHandler("/root/bindtest/", "/root/bindtest/")
	//handler = p
	config := ""
	dnsStartReq := pb.DNSStartReq{Config: config}
	handler.StartDNS(dnsStartReq)

	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}

	createACLReq := pb.CreateACLReq{
		Name: "southchina",
		ID:   "ACL001",
		IPs:  ipList}
	handler.CreateACL(createACLReq)
	deleteACLReq := pb.DeleteACLReq{ID: "ACL001"}
	handler.DeleteACL(deleteACLReq)
	handler.CreateACL(createACLReq)

	createViewReq := pb.CreateViewReq{
		ViewName: "DianXinView",
		ViewID:   "viewID001",
		Priority: 1,
		ACLIDs:   []string{"ACL001"}}
	handler.CreateView(createViewReq)
	createViewReq2 := pb.CreateViewReq{
		ViewName: "DianXinView02",
		ViewID:   "viewID002",
		Priority: 1,
		ACLIDs:   []string{}}
	//ACLIDs:   []string{"ACL001"}}
	handler.CreateView(createViewReq2)

	//var deleteIPList = []string{"192.168.199.0/24", "192.168.198.0/24"}
	//var newIPList = []string{"192.168.196.0/24", "192.168.197.0/24"}
	updateViewReq := pb.UpdateViewReq{ViewID: "viewID001", Priority: 2, DeleteACLIDs: []string{"ACL001"}, AddACLIDs: []string{}}
	handler.UpdateView(updateViewReq)
	createZoneReq := pb.CreateZoneReq{ViewID: "viewID001", ZoneID: "zoneID001", ZoneName: "test1031.com", ZoneFileName: "test1031.com.zone"}
	handler.CreateZone(createZoneReq)
	createRRReq := pb.CreateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rrid001", Name: "wwww", TTL: "1000", Type: "A", Value: "192.168.179.116"}
	handler.CreateRR(createRRReq)
	updateRRReq := pb.UpdateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rrid001", Name: "wwww", TTL: "1000", Type: "A", Value: "192.168.179.117"}
	handler.UpdateRR(updateRRReq)
	deleteRRReq := pb.DeleteRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rrid001"}
	handler.DeleteRR(deleteRRReq)
	deleteZoneReq := pb.DeleteZoneReq{ViewID: "viewID001", ZoneID: "zoneID001"}
	handler.DeleteZone(deleteZoneReq)
	deleteViewReq := pb.DeleteViewReq{ViewID: "viewID001"}
	handler.DeleteView(deleteViewReq)

	handler.StopDNS()
}
