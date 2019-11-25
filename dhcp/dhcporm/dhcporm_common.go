package dhcporm

const dhcpv4Ver string = "4"
const dhcpv6Ver string = "6"

// Subnet is our model, which corresponds to the "subnets" database
// table.
type Subnet struct {
	ID            string `gorm:"primary_key"`
	Subnet        string `gorm:"column:subnet"`
	ValidLifetime string `gorm:"column:validlifetime"`
	DhcpVer       string `gorm:"column:dhcpver"`
}
