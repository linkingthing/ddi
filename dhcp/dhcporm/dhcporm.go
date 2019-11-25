package dhcporm

import (
	"fmt"

	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func Subnetv4List(db *gorm.DB, dhcpVer string) []Subnet {

	var subnets []Subnet
	query := db.Where("dhcpver = ?", dhcpVer).Find(&subnets)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	return subnets
}

func GetSubnetv4(db *gorm.DB, name string) []Subnet {

	var subnetv4s []Subnet
	db.Where("dhcpver = ? and subnet = ? ", dhcpv4Ver, name).Find(&subnetv4s)

	return subnetv4s
}

func CreateSubnetv4(db *gorm.DB, name string, validLifetime string) error {
	var subnet = Subnet{
		Subnet:        name,
		ValidLifetime: validLifetime,
		DhcpVer:       dhcpv4Ver,
	}

	query := db.Create(&subnet)

	if query.Error != nil {
		return fmt.Errorf("create subnet error, subnet name: " + name)
	}

	return nil
}
func UpdateSubnetv4(db *gorm.DB, name string, validLifetime string) error {

	//search subnet, if not exist, return error
	subnetv4s := GetSubnetv4(db, name)
	if len(subnetv4s) == 0 {
		return fmt.Errorf(name + " not exists, return")
	}
	subnet := subnetv4s[0]

	db.Model(&subnet).Update("validlifetime", validLifetime)

	return nil
}
func DeleteSubnetv4(db *gorm.DB, name string) error {

	query := db.Where("subnet = ? ", name).Delete(Subnet{})

	if query.Error != nil {
		return fmt.Errorf("create subnet error, subnet name: " + name)
	}

	return nil
}
