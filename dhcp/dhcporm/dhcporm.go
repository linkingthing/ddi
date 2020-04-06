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
	Dhcpv4ConfId uint
	Name         string `gorm:"column:name"`
	//SubnetId        string          `gorm:"column:subnet_id"`
	Subnet          string           `gorm:"column:subnet"`
	ValidLifetime   string           `gorm:"column:valid_life_time"`
	Reservations    []OrmReservation `gorm:"foreignkey:Subnetv4ID"`
	Pools           []Pool           `gorm:"foreignkey:Subnetv4ID"`
	ManualAddresses []ManualAddress  `gorm:"foreignkey:Subnetv4ID"`
	//DhcpVer       string `gorm:"column:dhcpver"`
}
type OrmSubnetv4Front struct {
	DbS4    OrmSubnetv4
	S4Name  string
	S4Total int
	S4Usage float64
}

func (OrmSubnetv4) TableName() string {
	return "subnetv4s"
}

/*type Reservation struct {
	gorm.Model
	//Subnetv4Id   int32  `json:"subnetv4_id"`
	BootFileName string `json:"boot_file_name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid      string `json:"duid"`
	Hostname  string `json:"hostname"`
	IpAddress string `json:"ip-address"`
	//NextServer string `json:"next-server"`
	//OptionData     []Option `json:"option-data"`
	//ServerHostname string   `json:"server-hostname"`
	Subnetv4ID uint `json:"subnetv4_id" sql:"type:integer REFERENCES subnetv4s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
	//OrmSubnetv4   OrmSubnetv4
	//SubnetRefer uint `json:"subnetv4_refer" sql:"type:bigint REFERENCES subnetv4s(id) ON DELETE CASCADE"`
}*/

type OrmReservation struct {
	gorm.Model
	Duid           string `gorm:"duid"`
	ReservType     string `gorm:"reserv_type"`
	ReservValue    string `gorm:"reserv_value"`
	IpAddress      string `gorm:"ip_address"`
	Hostname       string
	ClientId       string
	CircuitId      string
	NextServer     string
	ServerHostname string
	HwAddress      string `json:"hw_address"`
	BootFileName   string
	OptionData     []Option `gorm:"foreignkey:ReservationID"`
	Subnetv4ID     uint     `json:"subnetv4_id" sql:"type:integer REFERENCES subnetv4s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
}

func (OrmReservation) TableName() string {
	return "reservations"
}

type OrmOptionName struct {
	gorm.Model
	OptionName string `json:optionName`
	OptionId   int    `json:optionId`
	OptionType string `json:optionType`
	OptionVer  string `json:optionVer` // v4 or v6
}
type Option struct {
	gorm.Model
	AlwaysSend    bool   `gorm:"column:always-send"`
	Code          uint64 `gorm:"column:code"`
	CsvFormat     bool   `json:"csv-format"`
	Data          string `json:"data"`
	Name          string `json:"name"`
	Space         string `json:"space"`
	ReservationID uint   `sql:"type:integer REFERENCES reservations(id) on update cascade on delete cascade"`
}

type Pool struct {
	gorm.Model
	OptionData []Option `json:"option-data"`
	//Pool       string   `json:"pool"`
	BeginAddress     string `json:"begin-address"`
	EndAddress       string `json:"end-address"`
	Subnetv4ID       uint   `sql:"type:integer REFERENCES subnetv4s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
	MaxValidLifetime int    `json:"max-valid-lifetime"`
	ValidLifetime    int    `json:"valid-lifetime"`
}

type OrmUser struct {
	Username string
	Password string
}

type ManualAddress struct {
	gorm.Model
	IpAddress  string
	Comment    string
	Subnetv4ID uint `sql:"type:integer REFERENCES subnetv4s(id) ON UPDATE CASCADE ON DELETE CASCADE"`
}

type AliveAddress struct {
	IPAddress     string `gorm:"primary_key"`
	LastAliveTime int64
	ScanTime      int64
	Subnetv4ID    uint `sql:"type:integer REFERENCES subnetv4s(id)"`
}

type Ipv6PlanedAddrTree struct {
	gorm.Model
	Depth    int
	Name     string
	ParentID uint
	Subnet   string
	NodeCode int
	MaxCode  int
	IsLeaf   bool
}

type BitsUseFor struct {
	Parentid uint
	UsedFor  string
}
