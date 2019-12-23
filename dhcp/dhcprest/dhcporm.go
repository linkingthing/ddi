package dhcprest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"log"
	"strconv"
)

const Dhcpv4Ver string = "4"
const Dhcpv6Ver string = "6"
const CRDBAddr = "postgresql://maxroach@localhost:26257/postgres?ssl=true&sslmode=require&sslrootcert=/root/download/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/download/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/download/cockroach-v19.2.0/certs/client.maxroach.crt"

var PGDBConn *PGDB

type PGDB struct {
	DB *gorm.DB
}

func init() {
	PGDBConn = NewPGDB()
}

func NewPGDB() *PGDB {
	p := &PGDB{}
	var err error
	p.DB, err = gorm.Open("postgres", CRDBAddr)
	if err != nil {
		log.Fatal(err)
	}

	p.DB.AutoMigrate(&dhcporm.OrmSubnetv4{})
	p.DB.AutoMigrate(&dhcporm.Reservation{})

	return p
}

func (handler *PGDB) Close() {
	handler.DB.Close()
}

func GetDhcpv4Conf(db *gorm.DB) interface{} {

	return nil
}

func (handler *PGDB) Subnetv4List(db *gorm.DB) []dhcporm.OrmSubnetv4 {
	var subnetv4s []dhcporm.OrmSubnetv4
	query := db.Find(&subnetv4s)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	for k, v := range subnetv4s {
		rsv := []*dhcporm.Reservation{}
		if err := db.Where("subnetv4_id = ?", strconv.Itoa(int(v.ID))).Find(&rsv).Error; err != nil {
			log.Print(err)
		}
		subnetv4s[k].Reservations = rsv
	}
	return subnetv4s
}

func (handler *PGDB) GetSubnetv4ByName(db *gorm.DB, name string) *dhcporm.OrmSubnetv4 {
	log.Println("in GetSubnetv4ByName, name: ", name)

	var subnetv4 dhcporm.OrmSubnetv4
	db.Where(&dhcporm.OrmSubnetv4{Subnet: name}).Find(&subnetv4)

	return &subnetv4
}

func (handler *PGDB) GetSubnetv4(db *gorm.DB, id string) *dhcporm.OrmSubnetv4 {
	dbId := ConvertStringToUint(id)

	subnetv4 := dhcporm.OrmSubnetv4{}
	subnetv4.ID = dbId
	db.Preload("Reservations").First(&subnetv4)

	return &subnetv4
}

func (handler *PGDB) CreateSubnetv4(db *gorm.DB, name string, validLifetime string) error {
	var subnet = dhcporm.OrmSubnetv4{
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
func (handler *PGDB) UpdateSubnetv4(db *gorm.DB, name string, validLifetime string) error {

	log.Println("into dhcporm, UpdateSubnetv4, name: ", name)
	//search subnet, if not exist, return error
	subnet := handler.GetSubnetv4ByName(db, name)
	if subnet == nil {
		return fmt.Errorf(name + " not exists, return")
	}

	db.Model(&subnet).Update("valid_life_time", validLifetime)

	return nil
}
func (handler *PGDB) DeleteSubnetv4(db *gorm.DB, id string) error {
	log.Println("into dhcprest DeleteSubnetv4, id ", id)
	dbId := ConvertStringToUint(id)

	query := db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.OrmSubnetv4{})

	if query.Error != nil {
		return fmt.Errorf("delete subnet error, subnet id: " + id)
	}

	return nil
}

func (handler *PGDB) OrmReservationList(db *gorm.DB, subnetId string) []*dhcporm.Reservation {
	log.Println("in dhcprest, OrmReservationList, subnetId: ", subnetId)
	var reservations []*dhcporm.Reservation
	var rsvs []dhcporm.Reservation

	subnetIdUint := ConvertStringToUint(subnetId)
	if err := db.Where("subnetv4_id = ?", subnetIdUint).Find(&rsvs).Error; err != nil {
		return nil
	}

	for _, rsv := range rsvs {
		rsv2 := rsv
		rsv2.ID = rsv.ID
		rsv2.Subnetv4ID = subnetIdUint

		reservations = append(reservations, &rsv2)
	}

	return reservations
}

func (handler *PGDB) OrmGetReservation(db *gorm.DB, subnetId string, rsv_id string) *dhcporm.Reservation {
	log.Println("into rest OrmGetReservation, subnetId: ", subnetId, "rsv_id: ", rsv_id)
	dbRsvId := ConvertStringToUint(rsv_id)

	rsv := dhcporm.Reservation{}
	if err := db.First(&rsv, int(dbRsvId)).Error; err != nil {
		//fmt.Errorf("get reservation error, subnetId: ", subnetId, " reservation id: ", rsv_id)
		return nil
	}

	return &rsv
}

func (handler *PGDB) OrmCreateReservation(db *gorm.DB, subnetv4_id string, r *RestReservation) (dhcporm.Reservation, error) {
	log.Println("into OrmCreateReservation")
	var rsv = dhcporm.Reservation{
		Duid:         r.Duid,
		BootFileName: r.BootFileName,
		Subnetv4ID:   ConvertStringToUint(subnetv4_id),
		Hostname:     r.Hostname,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := db.Create(&rsv)
	if query.Error != nil {
		return dhcporm.Reservation{}, fmt.Errorf("CreateReservation error, duid: " + r.Duid)
	}

	return rsv, nil
}

func (handler *PGDB) OrmUpdateReservation(db *gorm.DB, subnetv4_id string, r *RestReservation) error {

	log.Println("into dhcporm, OrmUpdateReservation, id: ", r.GetID())

	//search subnet, if not exist, return error
	//subnet := handler.OrmGetReservation(db, subnetv4_id, r.GetID())
	//if subnet == nil {
	//	return fmt.Errorf(name + " not exists, return")
	//}

	ormRsv := dhcporm.Reservation{}
	ormRsv.ID = ConvertStringToUint(r.GetID())
	ormRsv.Hostname = r.Hostname
	ormRsv.Duid = r.Duid
	ormRsv.BootFileName = r.BootFileName

	db.Model(&ormRsv).Updates(ormRsv)

	return nil
}

func (handler *PGDB) OrmDeleteReservation(db *gorm.DB, id string) error {
	log.Println("into dhcprest OrmDeleteReservation, id ", id)
	dbId := ConvertStringToUint(id)

	query := db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.Reservation{})

	if query.Error != nil {
		return fmt.Errorf("delete subnet Reservation error, Reservation id: " + id)
	}

	return nil
}
