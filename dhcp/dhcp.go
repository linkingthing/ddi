package dhcp

const (
	DhcpHost        = "10.0.0.15"
	DhcpPort        = "8081"
	DhcpConfigPath  = "/usr/local/etc/kea/"
	Dhcp4ConfigFile = "kea-dhcp4.conf"
	Dhcp6ConfigFile = "kea-dhcp6.conf"

	KEADHCPv4Service = "dhcp4"
	KEADHCPv6Service = "dhcp6"
)
const (
	IntfStartDHCPv4 = 1 + iota
	IntfStopDHCPv4
	IntfCreateSubnetv4
	IntfUpdateSubnetv4
	IntfDeleteSubnetv4
)
