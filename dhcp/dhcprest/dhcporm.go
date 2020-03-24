package dhcprest

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dhcp/agent/dhcpv4agent"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	dhcpgrpc "github.com/linkingthing/ddi/dhcp/grpc"
	"github.com/linkingthing/ddi/ipam"
	"github.com/linkingthing/ddi/pb"
	"log"
	"strconv"
	"strings"
)

const Dhcpv4Ver string = "4"

const CRDBAddr = "postgresql://maxroach@localhost:26257/ddi?ssl=true&sslmode=require&sslrootcert=/root/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/cockroach-v19.2.0/certs/client.maxroach.crt"

var PGDBConn *PGDB

type PGDB struct {
	db *gorm.DB
}

/*func init() {
	PGDBConn = NewPGDB()
}*/

func NewPGDB(db *gorm.DB) *PGDB {
	p := &PGDB{}
	//var err error
	/*p.db, err = gorm.Open("postgres", CRDBAddr)
	if err != nil {
		log.Fatal(err)
	}*/
	p.db = db

	p.db.AutoMigrate(&dhcporm.OrmSubnetv4{})
	p.db.AutoMigrate(&dhcporm.Reservation{})
	p.db.AutoMigrate(&dhcporm.Option{})
	p.db.AutoMigrate(&dhcporm.Pool{})
	p.db.AutoMigrate(&dhcporm.ManualAddress{})

	p.db.AutoMigrate(&dhcporm.OrmSubnetv6{})
	p.db.AutoMigrate(&dhcporm.Reservationv6{})
	p.db.AutoMigrate(&dhcporm.Poolv6{})

	return p
}

func (handler *PGDB) Close() {
	handler.db.Close()
}

//func GetDhcpv4Conf(db *gorm.DB) interface{} {
//
//	return nil
//}

func (handler *PGDB) Subnetv4List() []dhcporm.OrmSubnetv4 {
	var subnetv4s []dhcporm.OrmSubnetv4
	query := handler.db.Find(&subnetv4s)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	for k, v := range subnetv4s {
		rsv := []dhcporm.Reservation{}
		if err := handler.db.Where("subnetv4_id = ?", strconv.Itoa(int(v.ID))).Find(&rsv).Error; err != nil {
			log.Print(err)
		}
		subnetv4s[k].Reservations = rsv
	}

	return subnetv4s
}

func (handler *PGDB) getSubnetv4BySubnet(subnet string) *dhcporm.OrmSubnetv4 {
	log.Println("in getSubnetv4BySubnet, subnet: ", subnet)

	var subnetv4 dhcporm.OrmSubnetv4
	handler.db.Where(&dhcporm.OrmSubnetv4{Subnet: subnet}).Find(&subnetv4)

	return &subnetv4
}

func (handler *PGDB) GetSubnetv4ById(id string) *dhcporm.OrmSubnetv4 {
	log.Println("in dhcp/dhcprest/GetSubnetv4ById, id: ", id)
	dbId := ConvertStringToUint(id)

	subnetv4 := dhcporm.OrmSubnetv4{}
	subnetv4.ID = dbId
	handler.db.Preload("Reservations").First(&subnetv4)

	return &subnetv4
}

//return (new inserted id, error)
func (handler *PGDB) CreateSubnetv4(name string, subnet string, validLifetime string) (string, error) {
	var s4 = dhcporm.OrmSubnetv4{
		Dhcpv4ConfId:  1,
		Name:          name,
		Subnet:        subnet,
		SubnetId:      "0",
		ValidLifetime: validLifetime,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := handler.db.Create(&s4)

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

func (handler *PGDB) UpdateSubnetv4(ormS4 dhcporm.OrmSubnetv4) error {
	log.Println("into dhcporm, UpdateSubnetv4, Subnet: ", ormS4.Subnet)
	//search subnet, if not exist, return error
	subnet := handler.getSubnetv4BySubnet(ormS4.Subnet)
	if subnet == nil {
		return fmt.Errorf(ormS4.Subnet + " not exists, return")
	}

	log.Println("subnet.id: ", subnet.ID)
	log.Println("subnet.name: ", subnet.Name)
	log.Println("subnet.subnet: ", subnet.Subnet)
	log.Println("subnet.subnet_id: ", subnet.SubnetId)
	log.Println("subnet.ValidLifetime: ", subnet.ValidLifetime)
	//if subnet.SubnetId == "" {
	//	subnet.SubnetId = strconv.Itoa(int(subnet.ID))
	//}

	db.Model(&subnet).Update(ormS4)

	return nil
}
func (handler *PGDB) DeleteSubnetv4(id string) error {
	log.Println("into dhcprest DeleteSubnetv4, id ", id)

	dbId := ConvertStringToUint(id)

	query := handler.db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.OrmSubnetv4{})

	s4 := handler.GetSubnetv4ById(db, id)
	err := db.Unscoped().Delete(s4).Error
	if err != nil {
		log.Println("删除子网出错: ", err)
		return err
	}
	//query := db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.OrmSubnetv4{})
	//aCLDB.ID = uint(dbId)
	//if err := tx.Unscoped().Delete(&aCLDB).Error; err != nil {
	//    return err
	//}

	return nil
}

func (handler *PGDB) OrmReservationList(subnetId string) []dhcporm.Reservation {
	log.Println("in dhcprest, OrmReservationList, subnetId: ", subnetId)
	var rsvs []dhcporm.Reservation

	subnetIdUint := ConvertStringToUint(subnetId)
	if err := handler.db.Where("subnetv4_id = ?", subnetIdUint).Find(&rsvs).Error; err != nil {
		panic(err)
		return nil
	}

	/*for _, rsv := range rsvs {
		rsv2 := rsv
		rsv2.ID = rsv.ID
		rsv2.ReservType = rsv.ReservType
		rsv2.IpAddress = rsv.IpAddress
		rsv2.Hostname = rsv.Hostname
		rsv2.NextServer = rsv.NextServer
		rsv2.ServerHostname = rsv.ServerHostname
		rsv2.BootFileName = rsv.BootFileName
		rsv2.Subnetv4ID = subnetIdUint
		var optionDatas []dhcporm.Option
		if err := handler.db.Where("reservation_id = ?", rsv.ID).Find(&optionDatas).Error; err != nil {
			return nil
		}
		for _, v := range optionDatas {
			rsv2.OptionData = append(rsv2.OptionData, v)
		}

		reservations = append(reservations, &rsv2)
	}*/

	return rsvs
}

func (handler *PGDB) OrmGetReservation(subnetId string, rsv_id string) *dhcporm.Reservation {
	log.Println("into rest OrmGetReservation, subnetId: ", subnetId, "rsv_id: ", rsv_id)
	dbRsvId := ConvertStringToUint(rsv_id)

	rsv := dhcporm.Reservation{}
	if err := handler.db.First(&rsv, int(dbRsvId)).Error; err != nil {
		//fmt.Errorf("get reservation error, subnetId: ", subnetId, " reservation id: ", rsv_id)
		return nil
	}

	return &rsv
}

func (handler *PGDB) OrmCreateReservation(subnetv4_id string, r *RestReservation) (dhcporm.Reservation, error) {
	log.Println("into OrmCreateReservation")
	var rsv = dhcporm.Reservation{
		Duid:         r.Duid,
		BootFileName: r.BootFileName,
		Subnetv4ID:   ConvertStringToUint(subnetv4_id),
		Hostname:     r.Hostname,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := handler.db.Create(&rsv)
	if query.Error != nil {
		return dhcporm.Reservation{}, fmt.Errorf("CreateReservation error, duid: " + r.Duid)
	}

	return rsv, nil
}

func (handler *PGDB) OrmUpdateReservation(subnetv4_id string, r *RestReservation) error {

	log.Println("into dhcporm, OrmUpdateReservation, id: ", r.GetID())

	//search subnet, if not exist, return error
	//subnet := handler.OrmGetReservation(subnetv4_id, r.GetID())
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

func (handler *PGDB) OrmDeleteReservation(id string) error {
	log.Println("into dhcprest OrmDeleteReservation, id ", id)
	dbId := ConvertStringToUint(id)

	query := handler.db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.Reservation{})

	if query.Error != nil {
		return fmt.Errorf("delete subnet Reservation error, Reservation id: " + id)
	}

	return nil
}

func (handler *PGDB) OrmPoolList(subnetId string) []*dhcporm.Pool {
	log.Println("in dhcprest, OrmPoolList, subnetId: ", subnetId)
	var pools []*dhcporm.Pool
	var ps []dhcporm.Pool

	subnetIdUint := ConvertStringToUint(subnetId)
	if err := handler.db.Where("subnetv4_id = ?", subnetIdUint).Find(&ps).Error; err != nil {
		return nil
	}

	for _, p := range ps {
		p2 := p
		p2.ID = p.ID
		p2.Subnetv4ID = subnetIdUint
		p2.BeginAddress = p.BeginAddress
		p2.EndAddress = p.EndAddress

		pools = append(pools, &p2)
	}

	return pools
}

func (handler *PGDB) OrmGetPool(subnetId string, rsv_id string) *dhcporm.Pool {
	log.Println("into rest OrmGetPool, subnetId: ", subnetId, "rsv_id: ", rsv_id)
	dbRsvId := ConvertStringToUint(rsv_id)

	rsv := dhcporm.Pool{}
	if err := handler.db.First(&rsv, int(dbRsvId)).Error; err != nil {
		//fmt.Errorf("get reservation error, subnetId: ", subnetId, " reservation id: ", rsv_id)
		return nil
	}

	return &rsv
}

func (handler *PGDB) OrmCreatePool(subnetv4_id string, r *RestPool) (dhcporm.Pool, error) {
	log.Println("into OrmCreatePool")
	var rsv = dhcporm.Pool{
		BeginAddress: r.BeginAddress,
		EndAddress:   r.EndAddress,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := handler.db.Create(&rsv)
	if query.Error != nil {
		return dhcporm.Pool{}, fmt.Errorf("CreatePool error, begin address: " + r.BeginAddress + ", end adderss: " + r.EndAddress)
	}

	return rsv, nil
}

func (handler *PGDB) OrmUpdatePool(subnetv4_id string, r *RestPool) error {

	log.Println("into dhcporm, OrmUpdatePool, id: ", r.GetID())

	//search subnet, if not exist, return error
	//subnet := handler.OrmGetReservation(subnetv4_id, r.GetID())
	//if subnet == nil {
	//	return fmt.Errorf(name + " not exists, return")
	//}

	ormRsv := dhcporm.Pool{}
	ormRsv.ID = ConvertStringToUint(r.GetID())
	ormRsv.BeginAddress = r.BeginAddress
	ormRsv.EndAddress = r.EndAddress

	db.Model(&ormRsv).Updates(ormRsv)

	return nil
}

func (handler *PGDB) OrmDeletePool(id string) error {
	log.Println("into dhcprest OrmDeletePool, id ", id)
	dbId := ConvertStringToUint(id)

	query := handler.db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.Pool{})

	if query.Error != nil {
		return fmt.Errorf("delete subnet Pool error, Reservation id: " + id)
	}

	return nil
}

func (handler *PGDB) GetDividedAddress(subNetID string) (*ipam.DividedAddress, error) {
	log.Println("into dhcporm GetDividedAddress, subNetID ", subNetID)
	one := ipam.DividedAddress{}
	one.SetID(subNetID)
	//get the reservation address
	data := handler.OrmReservationList(subNetID)
	for _, a := range data {
		if a.ReservType == "hw-address" || a.ReservType == "client-id" {
			//get the stable address
			one.Stable = append(one.Reserved, a.IpAddress)
		} else {
			one.Reserved = append(one.Reserved, a.IpAddress)
		}
	}
	//get the pools under the subnet
	pools := handler.OrmPoolList(subNetID)
	var dynamicAddress []string
	for _, pool := range pools {
		beginNums := strings.Split(pool.BeginAddress, ".")
		endNums := strings.Split(pool.EndAddress, ".")
		var err error
		var begin int
		var end int
		if begin, err = strconv.Atoi(string(beginNums[3])); err != nil {
			break
		}
		if end, err = strconv.Atoi(string(endNums[3])); err != nil {
			break
		}
		var beginPart string
		beginPart = beginNums[0] + "." + beginNums[1] + "." + beginNums[2]
		for i := begin; i <= end; i++ {
			dynamicAddress = append(dynamicAddress, beginPart+strconv.Itoa(i))
		}
	}
	found := false
	for _, ip := range dynamicAddress {
		found = false
		for _, a := range data {
			if ip == a.IpAddress {
				found = true
				break
			}
		}
		if !found {
			one.Dynamic = append(one.Dynamic, ip)
		}
	}
	//get manual address
	var manuals []dhcporm.ManualAddress
	if err := handler.db.Where("subnetv4_id = ?", subNetID).Find(&manuals).Error; err != nil {
		return nil, err
	}
	for _, v := range manuals {
		one.Manual = append(one.Manual, v.IpAddress)
	}
	//get the release address for the subnet
	leases := dhcpgrpc.GetLeases(subNetID)
	one.Lease = leases
	return &one, nil
}

func (handler *PGDB) GetScanAddress(id string) (*ipam.ScanAddress, error) {
	return nil, nil
}
