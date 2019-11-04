package dhcp

import "testing"

func TestStartDHCP(t *testing.T) {
	//if StartDHCP("dhcp4") != nil {
	//	t.Error(`StartDHCP("dhcp4") errored`)
	//}
}

func TestCreateSubnet(t *testing.T) {

	if CreateSubnet("dhcp4", "192.166.1.0/24", "192.166.1.10-192.168.1.20") != nil {
		t.Error(`create subnet4 failed`)
	}
}

