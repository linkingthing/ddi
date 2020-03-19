package dhcprest

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dhcp/agent/dhcpv4agent"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"github.com/linkingthing/ddi/pb"
	"log"
	"strconv"
)

const Dhcpv4Ver string = "4"

const CRDBAddr = "postgresql://maxroach@localhost:26257/ddi?ssl=true&sslmode=require&sslrootcert=/root/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/cockroach-v19.2.0/certs/client.maxroach.crt"

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
	p.DB.AutoMigrate(&dhcporm.Pool{})

	p.DB.AutoMigrate(&dhcporm.OrmSubnetv6{})
	p.DB.AutoMigrate(&dhcporm.Reservationv6{})
	p.DB.AutoMigrate(&dhcporm.Poolv6{})

	return p
}

func (handler *PGDB) Close() {
	handler.DB.Close()
}

//func GetDhcpv4Conf(db *gorm.DB) interface{} {
//
//	return nil
//}

func (handler *PGDB) Subnetv4List(db *gorm.DB) []dhcporm.OrmSubnetv4 {
	var subnetv4s []dhcporm.OrmSubnetv4
	query := db.Find(&subnetv4s)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	for k, v := range subnetv4s {

		subnetv4s[k].SubnetId = strconv.Itoa(int(v.ID))
		rsv := []*dhcporm.Reservation{}
		if err := db.Where("subnetv4_id = ?", strconv.Itoa(int(v.ID))).Find(&rsv).Error; err != nil {
			log.Print(err)
		}
		subnetv4s[k].Reservations = rsv
	}

	return subnetv4s
}

func (handler *PGDB) getSubnetv4BySubnet(db *gorm.DB, subnet string) *dhcporm.OrmSubnetv4 {
	log.Println("in getSubnetv4BySubnet, subnet: ", subnet)

	var subnetv4 dhcporm.OrmSubnetv4
	db.Where(&dhcporm.OrmSubnetv4{Subnet: subnet}).Find(&subnetv4)

	return &subnetv4
}

func (handler *PGDB) GetSubnetv4ById(db *gorm.DB, id string) *dhcporm.OrmSubnetv4 {
	log.Println("in dhcp/dhcprest/GetSubnetv4ById, id: ", id)
	dbId := ConvertStringToUint(id)

	subnetv4 := dhcporm.OrmSubnetv4{}
	subnetv4.ID = dbId
	db.Preload("Reservations").First(&subnetv4)

	return &subnetv4
}

//return (new inserted id, error)
func (handler *PGDB) CreateSubnetv4(db *gorm.DB, name string, subnet string, validLifetime string) (string, error) {
	var s4 = dhcporm.OrmSubnetv4{
		Dhcpv4ConfId:  1,
		Name:          name,
		Subnet:        subnet,
		SubnetId:      "0",
		ValidLifetime: validLifetime,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := db.Create(&s4)

	if query.Error != nil {
		return "", fmt.Errorf("create subnet error, subnet name: " + name)
	}
	var last dhcporm.OrmSubnetv4
	query.Last(&last)
	log.Println("query.value: ", query.Value, ", id: ", last.ID)

	//send msg to kafka queue, which is read by dhcp server
	req := pb.CreateSubnetv4Req{
		Subnet: subnet,
		Id:     strconv.Itoa(int(last.ID)),
	}
	log.Println("pb.CreateSubnetv4Req req: ", req)

	data, err := proto.Marshal(&req)
	if err != nil {
		return "", err
	}
	dhcp.SendDhcpCmd(data, dhcpv4agent.CreateSubnetv4)

	return strconv.Itoa(int(last.ID)), nil
}

func (handler *PGDB) UpdateSubnetv4(db *gorm.DB, ormS4 dhcporm.OrmSubnetv4) error {

	log.Println("into dhcporm, UpdateSubnetv4, name: ", ormS4.Name)
	//search subnet, if not exist, return error
	subnet := handler.getSubnetv4BySubnet(db, ormS4.Subnet)
	if subnet == nil {
		return fmt.Errorf(ormS4.Subnet + " not exists, return")
	}
	//if subnet.SubnetId == "" {
	//	subnet.SubnetId = strconv.Itoa(int(subnet.ID))
	//}

	db.Model(&subnet).Update(ormS4)

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

func (handler *PGDB) OrmPoolList(db *gorm.DB, subnetId string) []*dhcporm.Pool {
	log.Println("in dhcprest, OrmPoolList, subnetId: ", subnetId)
	var pools []*dhcporm.Pool
	var ps []dhcporm.Pool

	subnetIdUint := ConvertStringToUint(subnetId)
	if err := db.Where("subnetv4_id = ?", subnetIdUint).Find(&ps).Error; err != nil {
		return nil
	}

	for _, p := range ps {
		p2 := p
		p2.ID = p.ID
		p2.Subnetv4ID = subnetIdUint

		pools = append(pools, &p2)
	}

	return pools
}

func (handler *PGDB) OrmGetPool(db *gorm.DB, subnetId string, rsv_id string) *dhcporm.Pool {
	log.Println("into rest OrmGetPool, subnetId: ", subnetId, "rsv_id: ", rsv_id)
	dbRsvId := ConvertStringToUint(rsv_id)

	rsv := dhcporm.Pool{}
	if err := db.First(&rsv, int(dbRsvId)).Error; err != nil {
		//fmt.Errorf("get reservation error, subnetId: ", subnetId, " reservation id: ", rsv_id)
		return nil
	}

	return &rsv
}

func (handler *PGDB) OrmCreatePool(db *gorm.DB, subnetv4_id string, r *RestPool) (dhcporm.Pool, error) {
	log.Println("into OrmCreatePool")
	var rsv = dhcporm.Pool{
		Pool: r.Pool,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := db.Create(&rsv)
	if query.Error != nil {
		return dhcporm.Pool{}, fmt.Errorf("CreatePool error, pool: " + r.Pool)
	}

	return rsv, nil
}

func (handler *PGDB) OrmUpdatePool(db *gorm.DB, subnetv4_id string, r *RestPool) error {

	log.Println("into dhcporm, OrmUpdatePool, id: ", r.GetID())

	//search subnet, if not exist, return error
	//subnet := handler.OrmGetReservation(db, subnetv4_id, r.GetID())
	//if subnet == nil {
	//	return fmt.Errorf(name + " not exists, return")
	//}

	ormRsv := dhcporm.Pool{}
	ormRsv.ID = ConvertStringToUint(r.GetID())
	ormRsv.Pool = r.Pool

	db.Model(&ormRsv).Updates(ormRsv)

	return nil
}

func (handler *PGDB) OrmDeletePool(db *gorm.DB, id string) error {
	log.Println("into dhcprest OrmDeletePool, id ", id)
	dbId := ConvertStringToUint(id)

	query := db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.Pool{})

	if query.Error != nil {
		return fmt.Errorf("delete subnet Pool error, Reservation id: " + id)
	}

	return nil
}
