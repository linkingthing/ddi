package dhcp

import (
	ut "github.com/ben-han-cn/cement/unittest"
	"testing"
	"time"
	"github.com/linkingthing.com/ddi/pb"
)

var handler DHCPHandler
func init(){

	p := &KEAHandler{
		ConfigPath: "/usr/local/etc/kea/",
		MainConfName: "kea-dhcp4.conf",
	}
	handler = p
}
func TestStopDHCP(t *testing.T) {
	service := pb.DHCPStopReq{Service:"dhcp4"}
	err := handler.StopDHCP(service)
	ut.Assert(t, err == nil, "stop successfully!")
	time.Sleep(2 * time.Second)
}

func TestStartDHCP(t *testing.T) {

	configFile := configPath + configFileDHCP4
	service := pb.DHCPStartReq{Service:"dhcp4", ConfigFile:configFile}
	err := handler.StartDHCP(service)
	ut.Assert(t, err == nil, "start successfully!")
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