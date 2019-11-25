package dhcporm

const Dhcpv4Ver string = "4"
const Dhcpv6Ver string = "6"
const CRDBAddr = "postgresql://maxroach@localhost:26257/postgres?ssl=true&sslmode=require&sslrootcert=/root/download/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/download/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/download/cockroach-v19.2.0/certs/client.maxroach.crt"

// Subnet is our model, which corresponds to the "subnets" database
// table.
type Subnet struct {
	ID            string `gorm:"primary_key"`
	Subnet        string `gorm:"column:subnet"`
	ValidLifetime string `gorm:"column:validlifetime"`
	DhcpVer       string `gorm:"column:dhcpver"`
}
