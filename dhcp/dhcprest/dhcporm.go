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

	p.DB.AutoMigrate(&dhcporm.Subnetv4{})
	p.DB.AutoMigrate(&dhcporm.Reservation{})
	//p.DB.AutoMigrate(&dhcporm.User{})
	//p.DB.AutoMigrate(&dhcporm.CreditCard{})
	//p.DB.AutoMigrate(&dhcporm.Userd{})
	//p.DB.AutoMigrate(&dhcporm.CreditCardd{})
	//.AddForeignKey("owner", "users(id)", "CASCADE", "CASCADE")
	//one.db.AutoMigrate(&tb.DBACL{})

	return p
}

func (handler *PGDB) Close() {
	handler.DB.Close()
}

func GetDhcpv4Conf(db *gorm.DB) interface{} {

	return nil
}

func (handler *PGDB) Subnetv4List(db *gorm.DB) []dhcporm.Subnetv4 {
	var subnetv4s []dhcporm.Subnetv4
	query := db.Find(&subnetv4s)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	for _, v := range subnetv4s {
		rsv := []dhcporm.Reservation{}
		if err := db.Where("subnetv4_id3 = ?", strconv.Itoa(int(v.ID))).Find(&rsv).Error; err != nil {
			log.Print(err)
		}
	}
	return subnetv4s
}

func (handler *PGDB) GetSubnetv4ByName(db *gorm.DB, name string) *dhcporm.Subnetv4 {
	log.Println("in GetSubnetv4ByName, name: ", name)

	var subnetv4 *dhcporm.Subnetv4
	//v := db.Raw("SELECT * FROM subnetv4s WHERE id = ?", id).Scan(&subnetv4s)
	//v := db.First(&subnetv4, dbId)
	db.Where(&dhcporm.Subnetv4{Subnet: name}).Find(&subnetv4)

	log.Println("in GetSubnetv4ByName, subnetv4")
	log.Print(subnetv4)
	log.Println("in GetSubnetv4ByName over")

	return subnetv4
}

func (handler *PGDB) GetSubnetv4(db *gorm.DB, id string) *dhcporm.Subnetv4 {
	dbId, err := strconv.Atoi(id)
	if err != nil {
		return nil
	}

	subnetv4 := dhcporm.Subnetv4{}
	reservations := []dhcporm.Reservation{}
	subnetv4.ID = uint(dbId)
	db.Model(&subnetv4).Related(&reservations)
	subnetv4.Reservations = reservations

	return &subnetv4
}

func (handler *PGDB) CreateSubnetv4(db *gorm.DB, name string, validLifetime string) error {
	var subnet = dhcporm.Subnetv4{
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
	dbId, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("string to int convert error, id: ", id)
	}

	//db.Unscoped().Where("id = ? ", 1).Delete(dhcporm.Userd{})
	//log.Println("after delete user=1")

	query := db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.Subnetv4{})

	if query.Error != nil {
		return fmt.Errorf("delete subnet error, subnet id: " + id)
	}

	return nil
}
