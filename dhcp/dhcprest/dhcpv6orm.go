package dhcprest

import (
	"log"

	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
)

const Dhcpv6Ver string = "6"

func (handler *PGDB) GetSubnetv6ByName(db *gorm.DB, name string) *dhcporm.OrmSubnetv6 {
	log.Println("in GetSubnetv6ByName, name: ", name)

	var subnetv6 dhcporm.OrmSubnetv6
	db.Where(&dhcporm.OrmSubnetv6{Subnet: name}).Find(&subnetv6)

	return &subnetv6
}
