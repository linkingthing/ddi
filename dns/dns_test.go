package dns

import (
	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/linkingthing/ddi/pb"
	"testing"
)

var (
	handler *BindHandler
)

func init() {
	p := NewBindHandler("/root/bindtest/", "/root/bindtest/")
	handler = p
}

func TestStartDNS(t *testing.T) {
	config := ""
	dnsStartReq := pb.DNSStartReq{Config: config}
	err := handler.StartDNS(dnsStartReq)
	ut.Assert(t, err == nil, "start successfully:%v", err)
}

func TestCreateACL(t *testing.T) {
	var ipList = []string{"10.0.0.0/24", "192.168.198.0/24"}
	createACLReq := pb.CreateACLReq{
		Name: "southchina",
		ID:   "ACL001",
		IPs:  ipList}
	err := handler.CreateACL(createACLReq)
	ut.Assert(t, err == nil, "Create ACL successfully!:%v", err)
}

func TestUpdateACL(t *testing.T) {
	var ipList = []string{"10.0.0.0/24", "192.168.191.0/24"}
	updateACLReq := pb.UpdateACLReq{
		Name:   "southchina",
		ID:     "ACL001",
		NewIPs: ipList}
	err := handler.UpdateACL(updateACLReq)
	ut.Assert(t, err == nil, "Create ACL successfully!:%v", err)
}

func TestDeleteACL(t *testing.T) {
	deleteACLReq := pb.DeleteACLReq{ID: "ACL001"}
	err := handler.DeleteACL(deleteACLReq)
	ut.Assert(t, err == nil, "Delete ACL successfully!:%v", err)
}

func TestCreateView(t *testing.T) {
	TestCreateACL(t)
	createViewReq := pb.CreateViewReq{
		ViewName: "DianXinView",
		ViewID:   "viewID001",
		Priority: 1,
		ACLIDs:   []string{"ACL001"}}
	err := handler.CreateView(createViewReq)
	ut.Assert(t, err == nil, "Create View Success!:%v", err)
}

func TestUpdateView(t *testing.T) {
	var ipList = []string{"192.168.199.0/24", "192.168.198.0/24"}
	createACLReq := pb.CreateACLReq{
		Name: "southchina_2",
		ID:   "ACL002",
		IPs:  ipList}
	err := handler.CreateACL(createACLReq)
	ut.Assert(t, err == nil, "Create ACL successfully!:%v", err)
	updateViewReq := pb.UpdateViewReq{
		ViewID:       "viewID001",
		Priority:     1,
		DeleteACLIDs: []string{"ACL001"},
		AddACLIDs:    []string{"ACL002"}}
	err = handler.UpdateView(updateViewReq)
	ut.Assert(t, err == nil, "Create View Success!:%v", err)
}

func TestCreateZone(t *testing.T) {
	createZoneReq := pb.CreateZoneReq{ViewID: "viewID001", ZoneID: "zoneID001", ZoneName: "test1031.com", ZoneFileName: "test1031.com.zone"}
	err := handler.CreateZone(createZoneReq)
	ut.Assert(t, err == nil, "Create Zone Success!:%v", err)
}

func TestCreateRR(t *testing.T) {
	createRRReq := pb.CreateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rr002", Name: "mail.test1031.com", TTL: "1000", Type: "A", Value: "10.2.21.1"}
	err := handler.CreateRR(createRRReq)
	ut.Assert(t, err == nil, "Create RR Success!:%v", err)
	createRRReq = pb.CreateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rr003", Name: "mail.test1031.com", TTL: "1000", Type: "A", Value: "10.2.21.2"}
	err = handler.CreateRR(createRRReq)
	ut.Assert(t, err == nil, "Create RR Success!:%v", err)
}

func TestUpdateRR(t *testing.T) {
	updateRRReq := pb.UpdateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rr002", Name: "mail.test1031.com", TTL: "1000", Type: "A", Value: "10.2.21.3"}
	err := handler.UpdateRR(updateRRReq)
	ut.Assert(t, err == nil, "Update RR Success!:%v", err)

	updateRRReq = pb.UpdateRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rr003", Name: "mail.test1031.com", TTL: "1000", Type: "A", Value: "10.2.21.4"}
	err = handler.UpdateRR(updateRRReq)
	ut.Assert(t, err == nil, "Update RR Success!:%v", err)
}

func TestDeleteRR(t *testing.T) {
	delRRReq := pb.DeleteRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rr002"}
	err := handler.DeleteRR(delRRReq)
	ut.Assert(t, err == nil, "Delete RR Success!:%v", err)
	delRRReq = pb.DeleteRRReq{ViewID: "viewID001", ZoneID: "zoneID001", RRID: "rr003"}
	err = handler.DeleteRR(delRRReq)
	ut.Assert(t, err == nil, "Delete RR Success!:%v", err)
}

func TestDeleteZone(t *testing.T) {
	delZoneReq := pb.DeleteZoneReq{ViewID: "viewID001", ZoneID: "zoneID001"}
	err := handler.DeleteZone(delZoneReq)
	ut.Assert(t, err == nil, "Create Delete Zone Success!:%v", err)
}

func TestDeleteView(t *testing.T) {
	delViewReq := pb.DeleteViewReq{ViewID: "viewID001"}
	err := handler.DeleteView(delViewReq)
	ut.Assert(t, err == nil, "Delete View Success!:%v", err)
	deleteACLReq := pb.DeleteACLReq{ID: "ACL001"}
	err = handler.DeleteACL(deleteACLReq)
	ut.Assert(t, err == nil, "Delete ACL successfully!:%v", err)
	deleteACLReq.ID = "ACL002"
	err = handler.DeleteACL(deleteACLReq)
	ut.Assert(t, err == nil, "Delete ACL successfully!:%v", err)
}

func TestStopDNS(t *testing.T) {
	err := handler.StopDNS()
	ut.Assert(t, err == nil, "stop successfully!:%v", err)
}
