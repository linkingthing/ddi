package dhcporm

import (
	"fmt"

	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const dhcpv4Ver string = "4"
const dhcpv6Ver string = "6"

// Subnet is our model, which corresponds to the "subnets" database
// table.
type Subnet struct {
	ID            string `gorm:"primary_key"`
	Subnet        string `gorm:"column:subnet"`
	ValidLifetime string `gorm:"column:validlifetime"`
	DhcpServIp    string `gorm:"column:dhcpservip"`
	DhcpVer       string `gorm:"column:dhcpver"`
}

func Subnetv4List(db *gorm.DB, dhcpVer string, dhcpServIp string) []Subnet {

	var subnets []Subnet
	query := db.Where("dhcpver = ? AND dhcpservip = ?", dhcpVer, dhcpServIp).Find(&subnets)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	return subnets
}

func CreateSubnet(db *gorm.DB, name string, validLifetime string, dhcpServIp string, dhcpVer string) error {
	var subnet = Subnet{
		Subnet:        name,
		ValidLifetime: validLifetime,
		DhcpServIp:    dhcpServIp,
		DhcpVer:       dhcpVer,
	}

	query := db.Create(&subnet)
	if query.Error != nil {
		return fmt.Errorf("create subnet error, subnet name: " + name)
	}

	return nil
}
