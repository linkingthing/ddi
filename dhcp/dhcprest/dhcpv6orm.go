package dhcprest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"log"
	"strconv"
)

const Dhcpv6Ver string = "6"

func (handler *PGDB) Subnetv6List(db *gorm.DB) []dhcporm.OrmSubnetv6 {
	var subnetv6s []dhcporm.OrmSubnetv6
	query := db.Find(&subnetv6s)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	for k, v := range subnetv6s {
		rsv := []*dhcporm.Reservationv6{}
		if err := db.Where("subnetv6_id = ?", strconv.Itoa(int(v.ID))).Find(&rsv).Error; err != nil {
			log.Print(err)
		}
		subnetv6s[k].Reservations = rsv
	}
	return subnetv6s
}

func (handler *PGDB) GetSubnetv6ByName(db *gorm.DB, name string) *dhcporm.OrmSubnetv6 {
	log.Println("in GetSubnetv6ByName, name: ", name)

	var subnetv6 dhcporm.OrmSubnetv6
	db.Where(&dhcporm.OrmSubnetv6{Subnet: name}).Find(&subnetv6)

	return &subnetv6
}

func (handler *PGDB) GetSubnetv6(db *gorm.DB, id string) *dhcporm.OrmSubnetv6 {
	dbId := ConvertStringToUint(id)

	subnetv6 := dhcporm.OrmSubnetv6{}
	subnetv6.ID = dbId
	db.Preload("Reservations").First(&subnetv6)

	return &subnetv6
}

func (handler *PGDB) CreateSubnetv6(db *gorm.DB, name string, validLifetime string) error {
	var subnet = dhcporm.OrmSubnetv6{
		Dhcpv6ConfId:  1,
		Subnet:        name,
		ValidLifetime: validLifetime,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := db.Create(&subnet)

	if query.Error != nil {
		return fmt.Errorf("create subnet error, subnet name: " + name)
	}

	return nil
}
