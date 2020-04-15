package dhcporm

import "github.com/jinzhu/gorm"

type Dhcpv6Conf struct {
	gorm.Model
	Subnetv6      []OrmSubnetv6
	ValidLifetime string
	//OptionData    []Option
}

// Subnet is our model, which corresponds to the "subnetv4s" database
// table.
type OrmSubnetv6 struct {
	gorm.Model
	Dhcpv6ConfId   uint
	Name           string              `gorm:"column:name"`
	Subnet         string              `gorm:"column:subnet"`
	ValidLifetime  string              `gorm:"column:valid_life_time"`
	Reservationv6s []*OrmReservationv6 `gorm:"foreignkey:Subnetv6ID"`
	Pools          []*Poolv6           `gorm:"foreignkey:Subnetv6ID"`
	Gateway        string              `gorm:"gateway"`
	//ManualAddresses []ManualAddress    `gorm:"foreignkey:Subnetv6ID"`
	//DhcpVer       string `gorm:"column:dhcpver"`
}

func (OrmSubnetv6) TableName() string {
	return "subnetv6s"
}

type OrmReservationv6 struct {
	gorm.Model
	BootFileName   string `json:"boot_file_name"`
	Duid           string `gorm:"duid"`
	ReservType     string `gorm:"reserv_type"`
	ReservValue    string `gorm:"reserv_value"`
	IpAddress      string `gorm:"ip_address"`
	Hostname       string
	ClientId       string
	CircuitId      string
	NextServer     string
	ServerHostname string
	HwAddress      string   `json:"hw_address"`
	OptionData     []Option `gorm:"foreignkey:ReservationID"`
	Subnetv6ID     uint     `json:"subnetv6_id" sql:"type:integer REFERENCES subnetv6s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
}

type Poolv6 struct {
	gorm.Model
	OptionData       []Option `json:"option-data"`
	BeginAddress     string   `json:"begin-address"`
	EndAddress       string   `json:"end-address"`
	Subnetv6ID       uint     `sql:"type:integer REFERENCES subnetv6s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
	Pool             string   `json:"pool"`
	MaxValidLifetime int      `json:"max-valid-lifetime"`
	ValidLifetime    int      `json:"valid-lifetime"`
}
