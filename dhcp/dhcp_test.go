package dhcp

import "testing"

func TestStartDHCP(t *testing.T) {
	//if StartDHCP("dhcp4") != nil {
	//	t.Error(`StartDHCP("dhcp4") errored`)
	//}
}

func TestCreateSubnet(t *testing.T) {

	if CreateSubnet("dhcp4", "192.166.1.0", "255.255.255.0") != nil {
		t.Error(`create subnet4 failed`)
	}
}

