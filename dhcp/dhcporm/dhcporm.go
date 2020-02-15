package dhcporm

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Dhcpv4Conf struct {
	gorm.Model
	//ID       string `gorm:"primary_key"`
	OrmSubnetv4 []OrmSubnetv4

	ValidLifetime string
	//OptionData    []Option
}

// Subnet is our model, which corresponds to the "subnetv4s" database
// table.
type OrmSubnetv4 struct {
	gorm.Model
	Dhcpv4ConfId  uint
	Subnet        string         `gorm:"column:subnet"`
	ValidLifetime string         `gorm:"column:valid_life_time"`
	Reservations  []*Reservation `gorm:"foreignkey:Subnetv4ID"`
	//Pools []Pool `gorm:"foreignkey:SubnetRefer"`
	//DhcpVer       string `gorm:"column:dhcpver"`
}

func (OrmSubnetv4) TableName() string {
	return "subnetv4s"
}

type Reservation struct {
	gorm.Model
	//Subnetv4Id   int32  `json:"subnetv4_id"`
	BootFileName string `json:"boot_file_name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid     string `json:"duid"`
	Hostname string `json:"hostname"`
	//IpAddress  string `json:"ip-address"`
	//NextServer string `json:"next-server"`
	//OptionData     []Option `json:"option-data"`
	//ServerHostname string   `json:"server-hostname"`
	Subnetv4ID uint `json:"subnetv4_id" sql:"type:integer REFERENCES subnetv4s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
	//OrmSubnetv4   OrmSubnetv4
	//SubnetRefer uint `json:"subnetv4_refer" sql:"type:bigint REFERENCES subnetv4s(id) ON DELETE CASCADE"`
}

func (Reservation) TableName() string {
	return "reservations"
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
	Subnetv4ID uint     `json:"subnetv4_id" sql:"type:integer REFERENCES subnetv4s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
}
