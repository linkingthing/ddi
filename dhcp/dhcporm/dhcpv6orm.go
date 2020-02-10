package dhcporm

import "github.com/jinzhu/gorm"

type Dhcpv6Conf struct {
	gorm.Model
	Subnetv6      []OrmSubnetv6
	ValidLifetime string
	//OptionData    []Option
}

type OrmSubnetv6 struct {
	gorm.Model
	Dhcpv6ConfId  uint
	Subnet        string `gorm:"column:subnet"`
	ValidLifetime string `gorm:"column:validlifetime"`
	Reservations  []*Reservationv6
	Pools         []*Poolv6
}

type Reservationv6 struct {
	gorm.Model
	BootFileName string `json:"boot_file_name"`
	Duid         string `json:"duid"`
	Hostname     string `json:"hostname"`
	Subnetv6ID   uint   `json:"subnetv6_id" sql:"type:integer REFERENCES orm_subnetv6(id) ON UPDATE CASCADE ON DELETE CASCADE"`
	//Subnetv4Id   int32  `json:"subnetv4_id"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	//IpAddress  string `json:"ip-address"`
	//NextServer string `json:"next-server"`
	//OptionData     []Option `json:"option-data"`
	//ServerHostname string   `json:"server-hostname"`

}

type Poolv6 struct {
	gorm.Model
	OptionData  []Option `json:"option-data"`
	Pool        string   `json:"pool"`
	SubnetRefer uint
}
