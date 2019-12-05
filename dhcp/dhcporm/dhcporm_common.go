package dhcporm

import (
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp"
)

const Dhcpv4Ver string = "4"
const Dhcpv6Ver string = "6"
const CRDBAddr = "postgresql://maxroach@localhost:26257/postgres?ssl=true&sslmode=require&sslrootcert=/root/download/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/download/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/download/cockroach-v19.2.0/certs/client.maxroach.crt"

type Dhcpv4Conf struct {
	gorm.Model
	//ID       string `gorm:"primary_key"`
	Subnetv4 []Subnetv4

	ValidLifetime string
	//OptionData    []Option
}

// Subnet is our model, which corresponds to the "subnetv4s" database
// table.
type Subnetv4 struct {
	gorm.Model
	Dhcpv4ConfId int32
	Subnet       string `gorm:"column:subnet"`
	//DhcpVer       string `gorm:"column:dhcpver"`
	ValidLifetime string `gorm:"column:valid_life_time"`
	Reservations  []Reservation
	//Pools         []Pool
}

func (Subnetv4) TableName() string {
	return "subnetv4s"
}

type Reservation struct {
	gorm.Model
	Subnetv4Id   int32
	BootFileName string `json:"boot-file-name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid string `json:"duid"`
	//Hostname   string `json:"hostname"`
	//IpAddress  string `json:"ip-address"`
	//NextServer string `json:"next-server"`
	//OptionData     []Option `json:"option-data"`
	//ServerHostname string   `json:"server-hostname"`
}

type Option struct {
	gorm.Model
	AlwaysSend bool   `gorm:"column:always-send"`
	Code       uint64 `gorm:"column:code"`
	CsvFormat  bool   `json:"csv-format"`
	Data       string `json:"data"`
	Name       string `json:"name"`
	Space      string `json:"space"`
}
type Pool struct {
	gorm.Model
	OptionData []Option `json:"option-data"`
	Pool       string   `json:"pool"`
}

type Dhcpv6Conf struct {
	gorm.Model
	//ID       string `gorm:"primary_key"`
	Subnetv6 []Subnetv6

	ValidLifetime string
	OptionData    []Option
}

type Subnetv6 struct {
	gorm.Model
	//ID            string `gorm:"primary_key"`
	Subnet        string `gorm:"column:subnet"`
	DhcpVer       string `gorm:"column:dhcpver"`
	ValidLifetime string `gorm:"column:validlifetime"`
	Reservations  []dhcp.Reservations
	Pools         []dhcp.Pool
}
