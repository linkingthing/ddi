package dhcp

import (
	"testing"
	"time"
)
func TestStopDHCP(t *testing.T) {
	if StopDHCP("dhcp4") != nil {
		t.Error(`StopDHCP("dhcp4") errored`)
	}

}

func TestStartDHCP(t *testing.T) {
	if StartDHCP("dhcp4") != nil {
		t.Error(`StartDHCP("dhcp4") errored`)
	}

}

func TestSubnet(t *testing.T) {

	time.Sleep(2 * time.Second)
	if err := deleteSubnet4("dhcp4", "192.166.1.0/24"); err != nil {
		t.Error(err)
	}


	if err := CreateSubnet4("dhcp4", "192.166.1.0/24", "192.166.1.10-192.166.1.20"); err != nil {
		t.Error(err)
	}


	if err := updateSubnet4("dhcp4", "192.166.1.0/24", "192.166.1.0-192.166.1.55"); err != nil {
		t.Error(err)
	}


}