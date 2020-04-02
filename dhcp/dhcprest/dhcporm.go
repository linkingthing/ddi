package dhcprest

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dhcp/agent/dhcpv4agent"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	dhcpgrpc "github.com/linkingthing/ddi/dhcp/grpc"
	"github.com/linkingthing/ddi/ipam"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils/arp"
	"github.com/paulstuart/ping"
	"net"
	"time"
)

const Dhcpv4Ver string = "4"

const CRDBAddr = "postgresql://maxroach@localhost:26257/ddi?ssl=true&sslmode=require&sslrootcert=/root/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/cockroach-v19.2.0/certs/client.maxroach.crt"

const checkPeriod = 10

var PGDBConn *PGDB

type PGDB struct {
	db     *gorm.DB
	ticker *time.Ticker
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
	p.db.AutoMigrate(&dhcporm.OrmOptionName{})

	p.db.AutoMigrate(&dhcporm.OrmSubnetv6{})
	p.db.AutoMigrate(&dhcporm.Reservationv6{})
	p.db.AutoMigrate(&dhcporm.Poolv6{})
	p.db.AutoMigrate(&dhcporm.AliveAddress{})
	p.ticker = time.NewTicker(checkPeriod * time.Second)
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
		//log.Println("k: ", k, ", v: ", v)
		//log.Println("in Subnetv4List, v.ID: ", v.ID)

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
func (handler *PGDB) CreateSubnetv4(name string, subnet string, validLifetime string) (dhcporm.OrmSubnetv4, error) {
	log.Println("into CreateSubnetv4, name, subnet, validLifetime: ", name, subnet, validLifetime)
	var s4 = dhcporm.OrmSubnetv4{
		Dhcpv4ConfId:  1,
		Name:          name,
		Subnet:        subnet,
		ValidLifetime: validLifetime,
		//DhcpVer:       Dhcpv4Ver,
	}

	query := handler.db.Create(&s4)

	if query.Error != nil {
		return s4, fmt.Errorf("create subnet error, subnet name: " + name)
	}
	var last dhcporm.OrmSubnetv4
	query.Last(&last)
	log.Println("query.value: ", query.Value, ", id: ", last.ID)

	//send msg to kafka queue, which is read by dhcp server
	req := pb.CreateSubnetv4Req{
		Subnet:        subnet,
		Id:            strconv.Itoa(int(last.ID)),
		ValidLifetime: validLifetime,
	}
	log.Println("pb.CreateSubnetv4Req req: ", req)

	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	dhcp.SendDhcpCmd(data, dhcpv4agent.CreateSubnetv4)

	log.Println(" in CreateSubnetv4, last: ", last)
	return last, nil
}

func (handler *PGDB) OrmUpdateSubnetv4(subnetv4 *RestSubnetv4) error {
	log.Println("into dhcporm, OrmUpdateSubnetv4, Subnet: ", subnetv4.Subnet)

	dbS4 := dhcporm.OrmSubnetv4{}
	//dbS4.SubnetId = subnetv4.ID
	dbS4.Subnet = subnetv4.Subnet
	dbS4.Name = subnetv4.Name
	dbS4.ValidLifetime = subnetv4.ValidLifetime
	id, err := strconv.Atoi(subnetv4.ID)
	if err != nil {
		log.Println("subnetv4.ID error, id: ", subnetv4.ID)
		return err
	}
	dbS4.ID = uint(id)
	//if subnet.SubnetId == "" {
	//	subnet.SubnetId = strconv.Itoa(int(subnet.ID))
	//}

	log.Println("begin to save db, dbS4.ID: ", dbS4.ID)
	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&dbS4).Error; err != nil {
		return err
	}

	//todo send kafka msg
	req := pb.UpdateSubnetv4Req{Id: subnetv4.ID, Subnet: subnetv4.Subnet, ValidLifetime: subnetv4.ValidLifetime}
	data, err := proto.Marshal(&req)
	if err != nil {
		log.Println("proto.Marshal error, ", err)
		return err
	}
	log.Println("begin to call SendDhcpCmd, update subnetv4")
	if err := dhcp.SendDhcpCmd(data, dhcpv4agent.UpdateSubnetv4); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return err
	}

	//if err := restfulapi.SendCmdDhcpv4(data, dhcpv4agent.UpdateSubnetv4); err != nil { //
	//}
	//end of todo
	//db.Model(subnet).Update(ormS4)

	tx.Commit()
	return nil
}
func (handler *PGDB) DeleteSubnetv4(id string) error {
	log.Println("into dhcprest DeleteSubnetv4, id ", id)

	//dbId := ConvertStringToUint(id)
	//query := handler.db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.OrmSubnetv4{})
	var ormS4 dhcporm.OrmSubnetv4

	tx := handler.db.Begin()
	defer tx.Rollback()

	if err := tx.First(&ormS4, id).Error; err != nil {
		return fmt.Errorf("unknown subnetv4 with ID %s, %w", id, err)
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	ormS4.ID = uint(num)

	if err := tx.Unscoped().Delete(&ormS4).Error; err != nil {
		return err
	}
	req := pb.DeleteSubnetv4Req{Id: id, Subnet: ormS4.Subnet}
	log.Println("DeleteSubnetv4() req: ", req)
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := dhcp.SendDhcpCmd(data, dhcpv4agent.DeleteSubnetv4); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return err
	}
	tx.Commit()

	//s4 := handler.GetSubnetv4ById(id)
	//err := db.Unscoped().Delete(s4).Error
	//if err != nil {
	//	log.Println("删除子网出错: ", err)
	//	return err
	//}
	//query := db.Unscoped().Where("id = ? ", dbId).Delete(dhcporm.OrmSubnetv4{})
	//aCLDB.ID = uint(dbId)
	//if err := tx.Unscoped().Delete(&aCLDB).Error; err != nil {
	//    return err
	//}

	return nil
}

//return (new inserted id, error)
func (handler *PGDB) OrmSplitSubnetv4(s4 *dhcporm.OrmSubnetv4, newMask int) ([]*dhcporm.OrmSubnetv4, error) {
	log.Println("into OrmSplitSubnetv4, s4.subnet: ", s4.Subnet)

	var ormS4s []*dhcporm.OrmSubnetv4
	var err error

	// compute how many new subnets should be created
	newSubs := getSegs(s4.Subnet, newMask)
	for _, v := range newSubs {
		log.Println("in for loop, v: ", v)

		var newS4 dhcporm.OrmSubnetv4
		newS4, err = handler.CreateSubnetv4(v, v, "0")
		if err != nil {
			log.Println("create subnetv4 error, ", err)
			return ormS4s, err
		}
		ormS4s = append(ormS4s, &newS4)
	}
	log.Println("in OrmSplitSubnetv4, ormS4s: ", ormS4s)
	//todo delte ormSubnet4
	s4ID := strconv.Itoa(int(s4.ID))
	if err := handler.DeleteSubnetv4(s4ID); err != nil {
		log.Println("delete subnetv4 error, ", err)
		return ormS4s, err
	}
	log.Println("in OrmSplitSubnetv4, after delete ormS4s: ", ormS4s)
	return ormS4s, nil
	//todo

	//var last dhcporm.OrmSubnetv4
	//query.Last(&last)
	//log.Println("query.value: ", query.Value, ", id: ", last.ID)
	//
	////send msg to kafka queue, which is read by dhcp server
	//req := pb.CreateSubnetv4Req{
	//	Subnet:        subnet,
	//	Id:            strconv.Itoa(int(last.ID)),
	//	ValidLifetime: validLifetime,
	//}
	//log.Println("pb.CreateSubnetv4Req req: ", req)
	//
	//data, err := proto.Marshal(&req)
	//if err != nil {
	//	return last, err
	//}
	//dhcp.SendDhcpCmd(data, dhcpv4agent.CreateSubnetv4)

	//return restS4s, nil
}

/*
 * param: s4s, some subnet ids
 * param: newSubnet, new subnet cidr
 */
func (handler *PGDB) OrmMergeSubnetv4(s4IDs []string, newSubnet string) (*dhcporm.OrmSubnetv4, error) {
	log.Println("into OrmMergeSubnetv4, newSubnet: ", newSubnet, ", s4IDs: ", s4IDs)
	var s4Objs []*dhcporm.OrmSubnetv4
	var ormS4 dhcporm.OrmSubnetv4
	var err error

	//get subnets which will be merged
	for _, s4ID := range s4IDs {
		s4Obj := handler.GetSubnetv4ById(s4ID)
		s4Objs = append(s4Objs, s4Obj)

		// 1 delete every subnet which will be merged
		if err = handler.DeleteSubnetv4(s4ID); err != nil {
			log.Println("delete subnetv4 error, error: ", err)
			return &ormS4, err
		}
		log.Println("delete subnetv4 ok, s4id: ", s4ID)
	}
	log.Println("-- s4Objs: ", s4Objs)

	// 2 create new subnet with subnet: newSubnet, if some properties will be heritated further, fill them
	ormS4, err = handler.CreateSubnetv4(newSubnet, newSubnet, "0")
	if err != nil {
		log.Println("create subnetv4 error, ", err)
		return &ormS4, err
	}
	log.Println("create subnetv4 ok, newSubnet: ", newSubnet)
	return &ormS4, err
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

func (handler *PGDB) OrmCreateReservation(id string, r *RestReservation) (dhcporm.Reservation, error) {
	log.Println("into OrmCreateReservation")
	log.Println("in OrmCreateReservation, r.BootFileName: ", r.BootFileName)
	log.Println("in OrmCreateReservation, r.Subnetv4ID: ", r.IpAddress)
	log.Println("in OrmCreateReservation, r.BootFileName: ", r.ID)
	log.Println("in OrmCreateReservation, r.BootFileName: ", r.OptionData)
	log.Println("in OrmCreateReservation, r.BootFileName: ", r.BootFileName)
	var rsv = dhcporm.Reservation{
		//Duid:         r.Duid,
		BootFileName: r.BootFileName,
		Subnetv4ID:   ConvertStringToUint(id),
		Hostname:     r.Hostname,
		IpAddress:    r.IpAddress,
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
	//ormRsv.Duid = r.Duid
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

func (handler *PGDB) OrmOptionNameList() []*dhcporm.OrmOptionName {
	log.Println("in dhcprest, OrmOptionNameList,  ")
	var optionNames []*dhcporm.OrmOptionName
	var ps []dhcporm.OrmOptionName

	//subnetIdUint := ConvertStringToUint(subnetId)
	if err := handler.db.Find(&ps).Error; err != nil {
		return nil
	}

	for _, p := range ps {
		p2 := p
		p2.ID = p.ID
		p2.OptionName = p.OptionName
		p2.OptionType = p.OptionType
		p2.OptionId = p.OptionId
		p2.OptionVer = p.OptionVer

		optionNames = append(optionNames, &p2)
	}

	log.Println("in OrmOptionNameList, optionNames: ", optionNames)
	return optionNames
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

func (handler *PGDB) OrmGetPool(subnetId string, pool_id string) *dhcporm.Pool {
	log.Println("into rest OrmGetPool, subnetId: ", subnetId, "pool_id: ", pool_id)
	dbRsvId := ConvertStringToUint(pool_id)

	pool := dhcporm.Pool{}
	if err := handler.db.First(&pool, int(dbRsvId)).Error; err != nil {
		//fmt.Errorf("get reservation error, subnetId: ", subnetId, " reservation id: ", rsv_id)
		return nil
	}

	return &pool
}

func (handler *PGDB) OrmCreatePool(subnetv4_id string, r *RestPool) (dhcporm.Pool, error) {
	log.Println("into OrmCreatePool, r: ", r, ", subnetv4_id: ", subnetv4_id)

	sid, err := strconv.Atoi(subnetv4_id)
	if err != nil {
		log.Println("OrmCreatePool, sid error: ", subnetv4_id)
	}
	var ormPool dhcporm.Pool
	ormPool = dhcporm.Pool{
		Subnetv4ID:       uint(sid),
		BeginAddress:     r.BeginAddress,
		EndAddress:       r.EndAddress,
		OptionData:       []dhcporm.Option{},
		ValidLifetime:    r.ValidLifetime,
		MaxValidLifetime: r.MaxValidLifetime,
	}

	var pool = pb.Pools{
		Pool:             r.BeginAddress + "-" + r.EndAddress,
		Options:          []*pb.Option{},
		ValidLifetime:    strconv.Itoa(r.ValidLifetime),
		MaxValidLifetime: strconv.Itoa(r.MaxValidLifetime),

		//DhcpVer:       Dhcpv4Ver,
	}

	//get subnet by subnetv4_id
	ormSubnetv4 := handler.GetSubnetv4ById(subnetv4_id)
	s4Subnet := ormSubnetv4.Subnet

	//todo: post kafka msg to dhcp agent
	pools := []*pb.Pools{}
	pools = append(pools, &pool)
	req := pb.CreateSubnetv4PoolReq{
		Id:               subnetv4_id,
		Subnet:           s4Subnet,
		Pool:             pools,
		ValidLifetime:    pool.ValidLifetime,
		MaxValidLifetime: pool.MaxValidLifetime,
	}
	log.Println("OrmCreatePool, req: ", req)
	data, err := proto.Marshal(&req)
	if err != nil {
		return ormPool, err
	}
	if err := dhcp.SendDhcpCmd(data, dhcpv4agent.CreateSubnetv4Pool); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return ormPool, err
	}
	//end of todo

	query := handler.db.Create(&ormPool)
	if query.Error != nil {
		return dhcporm.Pool{}, fmt.Errorf("CreatePool error, begin address: " +
			r.BeginAddress + ", end adderss: " + r.EndAddress)
	}

	return ormPool, nil
}

func (handler *PGDB) OrmUpdatePool(subnetv4_id string, r *RestPool) error {

	log.Println("into dhcporm, OrmUpdatePool, id: ", r.GetID())

	//get subnetv4 name
	s4 := handler.GetSubnetv4ById(subnetv4_id)
	subnetName := s4.Subnet

	//oldPoolName := r.BeginAddress + "-" + r.EndAddress
	//search subnet, if not exist, return error
	oldPoolObj := handler.OrmGetPool(subnetv4_id, r.GetID())
	if oldPoolObj == nil {
		return fmt.Errorf("Pool not exists, return")
	}
	oldPoolName := oldPoolObj.BeginAddress + "-" + oldPoolObj.EndAddress

	ormPool := dhcporm.Pool{}
	ormPool.ID = ConvertStringToUint(r.GetID())
	ormPool.BeginAddress = r.BeginAddress
	ormPool.EndAddress = r.EndAddress
	ormPool.Subnetv4ID = ConvertStringToUint(subnetv4_id)
	ormPool.ValidLifetime = r.ValidLifetime
	ormPool.MaxValidLifetime = r.MaxValidLifetime

	log.Println("begin to save db, pool.ID: ", r.GetID(), ", pool.subnetv4id: ", ormPool.Subnetv4ID)

	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&ormPool).Error; err != nil {
		return err
	}
	//todo send kafka msg
	req := pb.UpdateSubnetv4PoolReq{
		Oldpool:          oldPoolName,
		Subnet:           subnetName,
		Pool:             ormPool.BeginAddress + "-" + ormPool.EndAddress,
		Options:          []*pb.Option{},
		ValidLifetime:    strconv.Itoa(ormPool.ValidLifetime),
		MaxValidLifetime: strconv.Itoa(ormPool.MaxValidLifetime),
	}
	data, err := proto.Marshal(&req)
	if err != nil {
		log.Println("proto.Marshal error, ", err)
		return err
	}
	log.Println("begin to call SendDhcpCmd, update subnetv4 pool, req: ", req)
	if err := dhcp.SendDhcpCmd(data, dhcpv4agent.UpdateSubnetv4Pool); err != nil {
		log.Println("SendDhcpCmd error, ", err)
		return err
	}

	//if err := restfulapi.SendCmdDhcpv4(data, dhcpv4agent.UpdateSubnetv4); err != nil { //
	//}
	//end of todo
	//db.Model(subnet).Update(ormS4)

	tx.Commit()
	return nil
	//db.Model(&ormPool).Updates(&ormPool)

	return nil
}

func (handler *PGDB) OrmDeletePool(id string) error {
	log.Println("into dhcprest OrmDeletePool, id ", id)

	var ormSubnetv4 dhcporm.OrmSubnetv4
	var ormPool dhcporm.Pool

	tx := handler.db.Begin()
	defer tx.Rollback()

	if err := tx.First(&ormPool, id).Error; err != nil {
		return fmt.Errorf("unknown subnetv4pool with ID %s, %w", id, err)
	}
	log.Println("subnetv4 id: ", ormPool.Subnetv4ID)

	if err := tx.First(&ormSubnetv4, ormPool.Subnetv4ID).Error; err != nil {
		return fmt.Errorf("unknown subnetv4 with ID %s, %w", ormPool.Subnetv4ID, err)
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	ormPool.ID = uint(num)

	if err := tx.Unscoped().Delete(&ormPool).Error; err != nil {
		return err
	}
	req := pb.DeleteSubnetv4PoolReq{
		Subnet: ormSubnetv4.Subnet,
		Pool:   ormPool.BeginAddress + "-" + ormPool.EndAddress,
	}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := dhcp.SendDhcpCmd(data, dhcpv4agent.DeleteSubnetv4Pool); err != nil {
		log.Println("SendDhcpCmd error, ", err)
		return err
	}
	tx.Commit()

	return nil
}

func (handler *PGDB) GetDividedAddress(subNetID string) (*ipam.DividedAddress, error) {
	log.Println("into dhcporm GetDividedAddress, subNetID ", subNetID)
	one := ipam.DividedAddress{}
	one.SetID(subNetID)
	//get the reservation address
	reservData := handler.OrmReservationList(subNetID)
	one.Data = make(map[string]ipam.StatusAddress, 255)
	allData := make(map[string]ipam.StatusAddress, 255)
	for _, a := range reservData {
		if a.ReservType == "hw-address" || a.ReservType == "client-id" {
			//get the stable address
			tmp := ipam.StatusAddress{AddressType: "stable"}
			allData[a.IpAddress] = tmp
		} else {
			tmp := ipam.StatusAddress{AddressType: "reserved"}
			allData[a.IpAddress] = tmp
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
		beginPart = beginNums[0] + "." + beginNums[1] + "." + beginNums[2] + "."
		for i := begin; i <= end; i++ {
			dynamicAddress = append(dynamicAddress, beginPart+strconv.Itoa(i))
		}
	}
	found := false
	for _, ip := range dynamicAddress {
		found = false
		for _, a := range reservData {
			if ip == a.IpAddress {
				found = true
				break
			}
		}
		if !found {
			tmp := ipam.StatusAddress{AddressType: "dynamic"}
			allData[ip] = tmp
		}
	}
	//get manual address
	var manuals []dhcporm.ManualAddress
	if err := handler.db.Where("subnetv4_id = ?", subNetID).Find(&manuals).Error; err != nil {
		return nil, err
	}
	for _, v := range manuals {
		tmp := ipam.StatusAddress{AddressType: "manual"}
		allData[v.IpAddress] = tmp
	}
	//get the release address for the subnet
	leases := dhcpgrpc.GetLeases(subNetID)
	for _, l := range leases {
		var macAddr string
		for i := 0; i < len(l.HwAddress); i++ {
			tmp := fmt.Sprintf("%d", l.HwAddress[i])
			macAddr += tmp
		}

		tmp := ipam.StatusAddress{MacAddress: macAddr, AddressType: "lease", LeaseStartTime: l.Expire - int64(l.ValidLifetime), LeaseEndTime: l.Expire}
		allData[l.IpAddress] = tmp
	}
	beginNums := strings.Split(pools[0].BeginAddress, ".")
	prefix := beginNums[0] + "." + beginNums[1] + "." + beginNums[2] + "."
	for i := 1; i < 256; i++ {
		if allData[prefix+strconv.Itoa(i)].AddressType == "" {
			tmp := ipam.StatusAddress{AddressType: "unused"}
			one.Data[prefix+strconv.Itoa(i)] = tmp
		} else {
			one.Data[prefix+strconv.Itoa(i)] = allData[prefix+strconv.Itoa(i)]
		}
	}
	return &one, nil
}

func (handler *PGDB) GetScanAddress(id string) (*ipam.ScanAddress, error) {
	leases := dhcpgrpc.GetLeases(id)
	var retData ipam.ScanAddress
	retData.SetID(id)
	retData.Data = make(map[string]ipam.StatusAddress, 255)
	var subnet dhcporm.OrmSubnetv4
	if err := handler.db.First(&subnet, id).Error; err != nil {
		return nil, err
	}
	originalReservData := handler.OrmReservationList(id)
	var notUsedIP []string
	var allIP []string
	//for not used IP adderss
	s := strings.Split(subnet.Subnet, "/")
	if len(s) == 2 && s[1] == "24" {
		nums := strings.Split(s[0], ".")
		if len(nums) >= 3 {
			for i := 1; i <= 255; i++ {
				allIP = append(allIP, nums[0]+"."+nums[1]+"."+nums[2]+"."+strconv.Itoa(i))
			}
		}
	}
	found := false
	for _, ip := range allIP {
		found = false
		for _, reserve := range originalReservData {
			if ip == reserve.IpAddress {
				found = true
				break
			}
		}
		if !found {
			for _, lease := range leases {
				if ip == lease.IpAddress {
					found = true
					break
				}
			}
		}
		if !found {
			notUsedIP = append(notUsedIP, ip)
		}
	}
	for _, ip := range notUsedIP {
		if _, err := arp.Arp("ens37", ip, 1); err != nil {
			log.Println(err)
			continue
		}
		tmp := ipam.StatusAddress{AddressType: "collision"}
		retData.Data[ip] = tmp
	}
	//for used ip addresses
	usedIP := map[string]string{}
	var reservs []dhcporm.Reservation
	if err := handler.db.Where("subnetv4_id = ?", id).Find(&reservs).Error; err != nil {
		return nil, err
	}
	for _, r := range reservs {
		if r.ReservType == "hw-address" {
			usedIP[r.IpAddress] = r.ReservValue
		}
	}
	for _, l := range leases {
		var macAddr string
		for i := 0; i < len(l.HwAddress); i++ {
			tmp := fmt.Sprintf("%d", l.HwAddress[i])
			macAddr += tmp
		}
		usedIP[l.IpAddress] = macAddr
	}
	var retMac *net.HardwareAddr
	var err error
	for ip, mac := range usedIP {
		retMac = nil
		if retMac, err = arp.Arp("ens37", ip, 1); err != nil {
			log.Println(err)
			continue
		}
		if retMac.String() != mac {
			tmp := ipam.StatusAddress{AddressType: "collision"}
			retData.Data[ip] = tmp
		}
	}
	//get the dead ip
	var alives []dhcporm.AliveAddress
	if err := handler.db.Where("subnetv4_id = ?", id).Find(&alives).Error; err != nil {
		return nil, err
	}
	for _, a := range alives {
		if time.Now().Unix()-a.LastAliveTime > 60*60*24 {
			tmp := ipam.StatusAddress{AddressType: "dead", ScanTime: a.ScanTime, LastAliveTime: a.LastAliveTime}
			retData.Data[a.IPAddress] = tmp
		}
	}
	return &retData, nil
}

func (handler *PGDB) KeepDetectAlive() {
	for {
		select {
		case <-handler.ticker.C:
			if err := handler.DetectAliveAddress(); err != nil {
				continue
			}
		}
	}
}

func (handler *PGDB) DetectAliveAddress() error {
	//get all the resevation address where reserv_type equal "hw-address" or "client-id"
	var reservs []dhcporm.Reservation
	if err := handler.db.Find(&reservs).Error; err != nil {
		return err
	}
	type stable struct {
		IP         string
		Subnetv4ID uint
	}
	var stables []stable
	for _, r := range reservs {
		if r.ReservType == "hw-address" || r.ReservType == "client-id" {
			tmp := stable{IP: r.IpAddress, Subnetv4ID: r.Subnetv4ID}
			stables = append(stables, tmp)
		}
	}
	/*type alive struct {
		IP         string
		ScanTime   int64
		Subnetv4ID uint
	}*/
	//var alives []alive
	var alives []dhcporm.AliveAddress
	for _, s := range stables {
		if ping.Ping(s.IP, 2) {
			tmp := dhcporm.AliveAddress{ScanTime: time.Now().Unix(), LastAliveTime: time.Now().Unix(), IPAddress: s.IP, Subnetv4ID: s.Subnetv4ID}
			alives = append(alives, tmp)
		} else {
			tmp := dhcporm.AliveAddress{ScanTime: time.Now().Unix(), LastAliveTime: 0, IPAddress: s.IP, Subnetv4ID: s.Subnetv4ID}
			alives = append(alives, tmp)
		}
	}
	tx := handler.db.Begin()
	defer tx.Rollback()
	for _, a := range alives {
		if a.LastAliveTime == 0 {
			tmp := dhcporm.AliveAddress{IPAddress: a.IPAddress}
			if err := tx.First(&tmp).Error; err != nil {
				if err := tx.Save(&a).Error; err != nil {
					return err
				}
			} else {
				a.LastAliveTime = tmp.LastAliveTime
				if err := tx.Save(&a).Error; err != nil {
					return err
				}
			}
		} else {
			if err := tx.Save(&a).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (handler *PGDB) OrmCreateOptionName(r *RestOptionName) (dhcporm.OrmOptionName, error) {
	log.Println("into OrmCreatePool, r: ", r)

	var ormOpName dhcporm.OrmOptionName
	ormOpName = dhcporm.OrmOptionName{
		OptionName: r.OptionName,
		OptionId:   r.OptionId,
		OptionVer:  r.OptionVer,
		OptionType: r.OptionType,
	}

	o := handler.getOptionNamebyNameVer(r)
	if o.ID > 0 {
		return ormOpName, fmt.Errorf("error, Option exists")
	}
	//get subnet by subnetv4_id
	//ormSubnetv4 := handler.GetSubnetv4ById(subnetv4_id)
	//s4Subnet := ormSubnetv4.Subnet

	query := handler.db.Create(&ormOpName)
	if query.Error != nil {
		return dhcporm.OrmOptionName{}, fmt.Errorf("CreateOptionName error: ")
	}

	return ormOpName, nil
}

// get option name by name and ver
func (handler *PGDB) getOptionNamebyNameVer(r *RestOptionName) *dhcporm.OrmOptionName {
	log.Println("in getOptionNamebyNameVer, OptionName: ", r.OptionName, ", ver: ", r.OptionVer)

	var ormOpName dhcporm.OrmOptionName
	handler.db.Where(&dhcporm.OrmOptionName{OptionName: r.OptionName, OptionVer: r.OptionVer}).Find(&ormOpName)

	return &ormOpName
}
