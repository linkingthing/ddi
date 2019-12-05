package dhcporm

import (
	"fmt"

	"log"

	"strconv"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func GetDhcpv4Conf(db *gorm.DB) interface{} {

	return nil
}

func Subnetv4List(db *gorm.DB) []Subnetv4 {

	var subnetv4s []Subnetv4
	//query := db.Find(&subnets)

	//Dhcpv4Conf{gorm.Model{ID:1}}
	query := db.Find(&subnetv4s)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	sid := []string{}
	for _, v := range subnetv4s {

		sid = append(sid, string(v.ID))
		rsv := []Reservation{}
		if err := db.Where("subnetv4_id3 = ?", strconv.Itoa(int(v.ID))).Find(&rsv).Error; err != nil {
			log.Print(err)
		}

		log.Println("+++ rsv, v.id: ", strconv.Itoa(int(v.ID)))
		log.Print(rsv)
		log.Println("--- rsv v.id: ", strconv.Itoa(int(v.ID)))

	}
	log.Println("+++ sid")
	log.Print(sid)
	log.Println("--- sid")
	ret := db.Where("subnetv4_id3 in (?)", sid).Find(&Reservation{})
	if ret.Error != nil {
		log.Print(ret.Error)
	}

	log.Println("+++ ret")
	log.Print(ret)
	log.Println("--- ret")

	return subnetv4s
}

func GetSubnetv4(db *gorm.DB, name string) []Subnetv4 {

	var subnetv4s []Subnetv4
	v := db.Where(" subnet = ? ", name).Find(&subnetv4s)
	log.Println("in getsubnetv4")
	log.Print(v)
	log.Println("in getsubnetv4 over")
	return subnetv4s
}

func CreateSubnetv4(db *gorm.DB, name string, validLifetime string) error {
	var subnet = Subnetv4{
		Dhcpv4ConfId:  1,
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

	query := db.Where("subnet = ? ", name).Delete(Subnetv4{})

	if query.Error != nil {
		return fmt.Errorf("create subnet error, subnet name: " + name)
	}

	return nil
}
