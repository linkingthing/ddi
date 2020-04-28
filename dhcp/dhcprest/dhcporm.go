package dhcprest

import (
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/dhcp/agent/dhcpv4agent"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	dhcpgrpc "github.com/linkingthing/ddi/dhcp/grpc"
	dnsapi "github.com/linkingthing/ddi/dns/restfulapi"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
)

const Dhcpv4Ver string = "4"
const CRDBAddr = "postgresql://maxroach@localhost:26257/ddi?ssl=true&sslmode=require&sslrootcert=/root/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/cockroach-v19.2.0/certs/client.maxroach.crt"
const checkPeriod = 10

var PGDBConn *PGDB

type PGDB struct {
	db     *gorm.DB
	ticker *time.Ticker
}

func NewPGDB(db *gorm.DB) *PGDB {
	p := &PGDB{}
	p.db = db
	p.db.AutoMigrate(&dhcporm.OrmSubnetv4{})
	p.db.AutoMigrate(&dhcporm.OrmReservation{})
	p.db.AutoMigrate(&dhcporm.Option{})
	p.db.AutoMigrate(&dhcporm.Pool{})
	p.db.AutoMigrate(&dhcporm.ManualAddress{})
	p.db.AutoMigrate(&dhcporm.OrmOptionName{})

	p.db.AutoMigrate(&dhcporm.OrmSubnetv6{})
	p.db.AutoMigrate(&dhcporm.OrmReservationv6{})
	p.db.AutoMigrate(&dhcporm.Poolv6{})
	//p.db.AutoMigrate(&dhcporm.AliveAddress{})
	if err := p.db.AutoMigrate(&dhcporm.IPAddress{}).Error; err != nil {
		panic(err)
	}
	p.ticker = time.NewTicker(checkPeriod * time.Second)
	return p
}

func (handler *PGDB) Close() {
	handler.db.Close()
}

func (handler *PGDB) Subnetv4List(search *SubnetSearch) []dhcporm.OrmSubnetv4 {
	var subnetv4s []dhcporm.OrmSubnetv4

	if search != nil && search.DhcpVer != "" {
		subnet := handler.getOrmSubnetv4BySubnet(search.Subnet)
		subnetv4s = append(subnetv4s, subnet)
	} else {
		query := handler.db.Find(&subnetv4s)
		if query.Error != nil {
			log.Print(query.Error.Error())
		}
	}

	for k, v := range subnetv4s {

		if len(v.Name) > 0 && len(v.ZoneName) == 0 {
			subnetv4s[k].ZoneName = v.Name
		}
		rsv := []dhcporm.OrmReservation{}
		if err := handler.db.Where("subnetv4_id = ?", strconv.Itoa(int(v.ID))).Find(&rsv).Error; err != nil {
			log.Print(err)
		}
		subnetv4s[k].Reservations = rsv
	}

	return subnetv4s
}

func (handler *PGDB) getOrmSubnetv4BySubnet(subnet string) dhcporm.OrmSubnetv4 {

	var subnetv4 dhcporm.OrmSubnetv4
	handler.db.Where(&dhcporm.OrmSubnetv4{Subnet: subnet}).Find(&subnetv4)
	return subnetv4
}

func (handler *PGDB) GetSubnetv4ById(id string) *dhcporm.OrmSubnetv4 {
	dbId := ConvertStringToUint(id)

	subnetv4 := dhcporm.OrmSubnetv4{}
	subnetv4.ID = dbId
	handler.db.Preload("Reservations").First(&subnetv4)

	return &subnetv4
}

//get Currently maxId from Kea config
func (handler *PGDB) GetSubnetMaxId() uint32 {
	var maxId uint32

	row := handler.db.Table("subnetv4s").Select("MAX(subnet_id)").Row()
	row.Scan(&maxId)
	log.Println("in GetSubnetMaxId, maxId: ", maxId)
	log.Println("in GetSubnetMaxId, utils.Subnetv4MaxId: ", utils.Subnetv4MaxId)
	if maxId < 100 {
		log.Println("in GetSubnetMaxId, set maxId to 100")
		maxId = 100
	}
	if utils.Subnetv4MaxId <= maxId {
		utils.Subnetv4MaxId = maxId
	}
	return maxId
}

//return (new inserted id, error)
func (handler *PGDB) CreateSubnetv4(restSubnetv4 *RestSubnetv4) (dhcporm.OrmSubnetv4, error) {
	log.Println("into CreateSubnetv4, name, subnet, validLifetime: ")

	var s4 = dhcporm.OrmSubnetv4{
		Name:             restSubnetv4.Name,
		SubnetId:         0,
		Subnet:           restSubnetv4.Subnet,
		ValidLifetime:    restSubnetv4.ValidLifetime,
		MaxValidLifetime: restSubnetv4.MaxValidLifetime,
		Gateway:          restSubnetv4.Gateway,
		DnsServer:        restSubnetv4.DnsServer,
		DhcpEnable:       restSubnetv4.DhcpEnable,
		ZoneName:         restSubnetv4.ZoneName,
		//DhcpVer:       Dhcpv4Ver,
	}
	maxId := handler.GetSubnetMaxId()
	s4.SubnetId = maxId + 1

	if len(s4.Name) > 0 && len(s4.ZoneName) == 0 {
		s4.ZoneName = s4.Name
	}
	query := handler.db.Create(&s4)

	if query.Error != nil {
		return s4, fmt.Errorf("create subnet error, subnet name: " + restSubnetv4.Name)
	}
	var last dhcporm.OrmSubnetv4
	query.Last(&last)
	log.Println("query.value: ", query.Value, ", id: ", last.ID)
	restSubnetv4.ID = strconv.Itoa(int(last.ID))

	//send msg to kafka queue, which is read by dhcp server
	req := pb.CreateSubnetv4Req{
		Subnet:        restSubnetv4.Subnet,
		Id:            s4.SubnetId,
		ValidLifetime: restSubnetv4.ValidLifetime,
		Gateway:       restSubnetv4.Gateway,
		DnsServer:     restSubnetv4.DnsServer,
	}
	log.Println("pb.CreateSubnetv4Req req.id: ", req.Id)
	log.Println("pb.CreateSubnetv4Req req.Subnet: ", req.Subnet)

	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	utils.SendDhcpCmd(data, dhcpv4agent.CreateSubnetv4)

	log.Println(" in CreateSubnetv4, last: ", last)
	return last, nil
}

func (handler *PGDB) OrmUpdateSubnetv4(subnetv4 *RestSubnetv4) error {
	log.Println("into dhcporm, OrmUpdateSubnetv4, Subnet: ", subnetv4.Subnet)

	dbS4 := dhcporm.OrmSubnetv4{}
	dbS4.Name = subnetv4.Name
	dbS4.ValidLifetime = subnetv4.ValidLifetime
	dbS4.MaxValidLifetime = subnetv4.MaxValidLifetime
	dbS4.ID = ConvertStringToUint(subnetv4.ID)
	dbS4.DhcpEnable = subnetv4.DhcpEnable
	dbS4.ZoneName = subnetv4.ZoneName
	dbS4.DnsEnable = subnetv4.DnsEnable
	dbS4.Note = subnetv4.Note
	dbS4.Gateway = subnetv4.Gateway
	dbS4.DnsServer = subnetv4.DnsServer
	if len(dbS4.Name) > 0 && len(dbS4.ZoneName) == 0 {
		dbS4.ZoneName = dbS4.Name
	}

	//get subnet name from db
	getOrmS4 := handler.GetSubnetv4ById(subnetv4.ID)
	dbS4.Subnet = getOrmS4.Subnet
	subnetv4.Subnet = getOrmS4.Subnet

	//added for new zone handler
	if subnetv4.DnsEnable > 0 {
		if len(subnetv4.ViewId) == 0 {
			log.Println("Error viewId is null, return")
			//return fmt.Errorf("zone is enabled, viewId is null")
		}
		zone := dnsapi.Zone{Name: subnetv4.ZoneName, ZoneType: "master"}
		log.Println("to create zone, name:", zone.Name)
		dnsapi.DBCon.CreateZone(&zone, subnetv4.ViewId)

	}

	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&dbS4).Error; err != nil {
		return err
	}

	//send msg to kafka queue, which is read by dhcp server
	req := pb.UpdateSubnetv4Req{
		Subnet:           subnetv4.Subnet,
		Id:               subnetv4.ID,
		ValidLifetime:    subnetv4.ValidLifetime,
		MaxValidLifetime: subnetv4.MaxValidLifetime,
		Gateway:          subnetv4.Gateway,
		DnsServer:        subnetv4.DnsServer,
	}
	data, err := proto.Marshal(&req)
	if err != nil {
		log.Println("proto.Marshal error, ", err)
		return err
	}

	if err := utils.SendDhcpCmd(data, dhcpv4agent.UpdateSubnetv4); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return err
	}

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
	if err := utils.SendDhcpCmd(data, dhcpv4agent.DeleteSubnetv4); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return err
	}
	tx.Commit()

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
		restS4 := RestSubnetv4{}
		var newS4 dhcporm.OrmSubnetv4
		restS4.Name = v
		restS4.Subnet = v
		newS4, err = handler.CreateSubnetv4(&restS4)
		if err != nil {
			log.Println("create subnetv4 error, ", err)
			return ormS4s, err
		}
		ormS4s = append(ormS4s, &newS4)
	}
	s4ID := strconv.Itoa(int(s4.ID))
	if err := handler.DeleteSubnetv4(s4ID); err != nil {
		log.Println("delete subnetv4 error, ", err)
		return ormS4s, err
	}

	return ormS4s, nil
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

	// 2 create new subnet with subnet: newSubnet, if some properties will be heritated further, fill them
	restS4 := RestSubnetv4{}
	restS4.Name = newSubnet
	restS4.Subnet = newSubnet
	ormS4, err = handler.CreateSubnetv4(&restS4)
	if err != nil {
		log.Println("create subnetv4 error, ", err)
		return &ormS4, err
	}
	log.Println("create subnetv4 ok, newSubnet: ", newSubnet)
	return &ormS4, err
}

func (handler *PGDB) OrmReservationList(subnetId string) []dhcporm.OrmReservation {
	log.Println("in dhcprest, OrmReservationList, subnetId: ", subnetId)
	var rsvs []dhcporm.OrmReservation

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

func (handler *PGDB) OrmGetReservation(subnetId string, rsv_id string) *dhcporm.OrmReservation {
	log.Println("into rest OrmGetReservation, subnetId: ", subnetId, "rsv_id: ", rsv_id)
	dbRsvId := ConvertStringToUint(rsv_id)

	rsv := dhcporm.OrmReservation{}
	if err := handler.db.First(&rsv, int(dbRsvId)).Error; err != nil {
		//fmt.Errorf("get reservation error, subnetId: ", subnetId, " reservation id: ", rsv_id)
		return nil
	}

	return &rsv
}

func (handler *PGDB) OrmCreateReservation(subnetv4_id string, r *RestReservation) (dhcporm.OrmReservation, error) {
	log.Println("into OrmCreateReservation, r: ", r, ", subnetv4_id: ", subnetv4_id)

	ormRsv := dhcporm.OrmReservation{
		//BootFileName: r.BootFileName,
		Subnetv4ID: ConvertStringToUint(subnetv4_id),
		Hostname:   r.Hostname,
		IpAddress:  r.IpAddress,
		HwAddress:  r.HwAddress,
		ClientId:   r.ClientId,
		CircuitId:  r.CircuitId,
		NextServer: r.NextServer,
		ReservType: r.ResvType,
		//DhcpVer:       Dhcpv4Ver,
	}
	pbRsv := pb.Reservation{
		Duid:        r.Duid,
		Hostname:    r.Hostname,
		IpAddresses: r.IpAddress,
		NextServer:  r.NextServer,
		HwAddress:   r.HwAddress,
		ClientId:    r.ClientId,
		CircuitId:   r.CircuitId,
		ResvType:    r.ResvType,
		//OptData:     r.OptionData,
	}

	//check whether subnet id exists
	s4Obj := handler.GetSubnetv4ById(subnetv4_id)
	if s4Obj.Subnet == "" {
		log.Println("subnet not exist")
		return ormRsv, fmt.Errorf("subnet not exist")
	}

	log.Println("begin to save db, ormRsv.IpAddress: ", ormRsv.IpAddress)
	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&ormRsv).Error; err != nil {
		log.Println("save ormRsv error: ", err)
		return ormRsv, err
	}

	//todo: post kafka msg to dhcp agent
	rsvs := []*pb.Reservation{}
	rsvs = append(rsvs, &pbRsv)
	req := pb.CreateSubnetv4ReservationReq{
		Subnet:     s4Obj.Subnet,
		IpAddr:     pbRsv.IpAddresses,
		Duid:       pbRsv.Duid,
		Hostname:   pbRsv.Hostname,
		HwAddress:  pbRsv.HwAddress,
		CircuitId:  pbRsv.CircuitId,
		ClientId:   pbRsv.ClientId,
		NextServer: pbRsv.NextServer,
		ResvType:   pbRsv.ResvType,
	}
	log.Println("OrmCreateReservation, req: ", req)
	data, err := proto.Marshal(&req)
	if err != nil {
		return ormRsv, err
	}
	if err := utils.SendDhcpCmd(data, dhcpv4agent.CreateSubnetv4Reservation); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return ormRsv, err
	}
	//end of todo

	tx.Commit()

	return ormRsv, nil
}

func (handler *PGDB) OrmUpdateReservation(subnetv4_id string, r *RestReservation) error {

	ormRsv := dhcporm.OrmReservation{
		//Duid:         r.Duid,
		BootFileName: r.BootFileName,
		Subnetv4ID:   ConvertStringToUint(subnetv4_id),
		Hostname:     r.Hostname,
		IpAddress:    r.IpAddress,
		//DhcpVer:       Dhcpv4Ver,
	}
	pbRsv := pb.Reservation{
		//Duid:        r.Duid,
		Hostname:    r.Hostname,
		IpAddresses: r.IpAddress,
		NextServer:  r.NextServer,
	}
	if len(r.Duid) > 0 {
		ormRsv.Duid = r.Duid
		pbRsv.Duid = r.Duid
	}

	//get curr rsv by id
	ormRsvID := r.GetID()
	oldOrmRsvObj := handler.OrmGetReservation(subnetv4_id, ormRsvID)
	log.Println("get old orm rsv obj: ", oldOrmRsvObj)

	//check whether subnet id exists
	s4Obj := handler.GetSubnetv4ById(subnetv4_id)
	if s4Obj.Subnet == "" {
		log.Println("subnet not exist")
		return fmt.Errorf("subnet not exist")
	}

	log.Println("begin to save db, ormRsv.IpAddress: ", ormRsv.IpAddress, ", oldIP: ", oldOrmRsvObj.IpAddress)
	tx := handler.db.Begin()
	defer tx.Rollback()
	handler.db.Model(&ormRsv).Updates(ormRsv)

	//todo: post kafka msg to dhcp agent
	rsvs := []*pb.Reservation{}
	rsvs = append(rsvs, &pbRsv)
	req := pb.UpdateSubnetv4ReservationReq{
		Subnet:     s4Obj.Subnet,
		IpAddr:     pbRsv.IpAddresses,
		Duid:       pbRsv.Duid,
		Hostname:   pbRsv.Hostname,
		NextServer: pbRsv.NextServer,
		OldRsvIP:   oldOrmRsvObj.IpAddress,
	}
	log.Println("OrmUpdateReservation, req: ", req)
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := utils.SendDhcpCmd(data, dhcpv4agent.UpdateSubnetv4Reservation); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return err
	}
	//end of todo

	tx.Commit()

	return nil
}

func (handler *PGDB) OrmDeleteReservation(id string) error {
	log.Println("into dhcprest OrmDeleteReservation, id ", id)

	var ormSubnetv4 dhcporm.OrmSubnetv4
	var ormRsv dhcporm.OrmReservation

	tx := handler.db.Begin()
	defer tx.Rollback()

	if err := tx.First(&ormRsv, id).Error; err != nil {
		return fmt.Errorf("unknown subnetv4rsv with ID %s, %w", id, err)
	}
	log.Println("subnetv4 id: ", ormRsv.Subnetv4ID)

	if err := tx.First(&ormSubnetv4, ormRsv.Subnetv4ID).Error; err != nil {
		return fmt.Errorf("unknown subnetv4 with ID %s, %w", ormRsv.Subnetv4ID, err)
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	ormRsv.ID = uint(num)

	if err := tx.Unscoped().Delete(&ormRsv).Error; err != nil {
		return err
	}
	req := pb.DeleteSubnetv4ReservationReq{
		Subnet: ormSubnetv4.Subnet,
		IpAddr: ormRsv.IpAddress,
	}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := utils.SendDhcpCmd(data, dhcpv4agent.DeleteSubnetv4Reservation); err != nil {
		log.Println("SendDhcpCmd error, ", err)
		return err
	}
	tx.Commit()

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

func (handler *PGDB) OrmUpdateOptionName(opName *RestOptionName) error {
	log.Println("into dhcporm, OrmUpdateOptionName, OptionName: ", opName.OptionName)

	dbOpName := dhcporm.OrmOptionName{}
	dbOpName.OptionName = opName.OptionName
	dbOpName.OptionVer = opName.OptionVer
	dbOpName.OptionId = opName.OptionId
	dbOpName.OptionType = opName.OptionType
	dbOpName.ID = ConvertStringToUint(opName.ID)

	log.Println("begin to save db, dbOpName.ID: ", dbOpName.ID)
	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&dbOpName).Error; err != nil {
		log.Println("update option name error: ", err)
		return err
	}

	tx.Commit()

	return nil
}

func (handler *PGDB) OrmDeleteOptionName(id string) error {
	log.Println("into dhcprest OrmDeleteOptionName, id ", id)

	var ormOpName dhcporm.OrmOptionName

	tx := handler.db.Begin()
	defer tx.Rollback()

	if err := tx.First(&ormOpName, id).Error; err != nil {
		return fmt.Errorf("unknown OptionName with ID %s, %w", id, err)
	}

	num, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	ormOpName.ID = uint(num)

	if err := tx.Unscoped().Delete(&ormOpName).Error; err != nil {
		return err
	}

	tx.Commit()

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

	////get the release address for the subnet
	//leases := dhcpgrpc.GetLeases(subnetId)
	//log.Println("in OrmPoolList, leases: ", leases)
	//for _, l := range leases {
	//	var macAddr string
	//	for i := 0; i < len(l.HwAddress); i++ {
	//		tmp := fmt.Sprintf("%d", l.HwAddress[i])
	//		macAddr += tmp
	//	}
	//	log.Println("in OrmPoolList, macAddr: ", macAddr)
	//
	//	tmp := ipam.StatusAddress{MacAddress: macAddr, AddressType: "lease", LeaseStartTime: l.Expire - int64(l.ValidLifetime), LeaseEndTime: l.Expire}
	//	//allData[l.IpAddress] = tmp
	//	log.Println("in OrmPoolList, tmp: ", tmp)
	//}

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

func (handler *PGDB) OrmCreatePool(subnetv4_id string, r *RestPool) (*dhcporm.Pool, error) {
	log.Println("into OrmCreatePool, r: ", r, ", subnetv4_id: ", subnetv4_id)

	ormPool := ConvertRestPool2OrmPool(r)

	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&ormPool).Error; err != nil {
		return ormPool, err
	}

	//get subnet by subnetv4_id
	ormSubnetv4 := handler.GetSubnetv4ById(subnetv4_id)
	s4Subnet := ormSubnetv4.Subnet

	// post kafka msg to dhcp agent
	var options []*dhcp.Option
	var err error
	options, err = dhcp.CreateOptionsFromPb(r.Gateway, r.DnsServer)
	if err != nil {
		log.Println("error to CreateOptionsFromPb, gateway: ", r.Gateway)
	}
	pbOptions := dhcp.CreatePbOptions(options)

	req := pb.CreateSubnetv4PoolReq{
		Id:               subnetv4_id,
		Subnet:           s4Subnet,
		Pool:             r.BeginAddress + "-" + r.EndAddress,
		Options:          pbOptions,
		ValidLifetime:    r.ValidLifetime,
		MaxValidLifetime: r.MaxValidLifetime,
	}
	log.Println("OrmCreatePool, req: ", req)
	data, err := proto.Marshal(&req)
	if err != nil {
		return ormPool, err
	}
	if err := utils.SendDhcpCmd(data, dhcpv4agent.CreateSubnetv4Pool); err != nil {
		log.Println("SendCmdDhcpv4 error, ", err)
		return ormPool, err
	}
	//end of kafka

	tx.Commit()
	return ormPool, nil
}

func (handler *PGDB) OrmUpdatePool(subnetv4_id string, r *RestPool) error {

	//get subnetv4 name
	s4 := handler.GetSubnetv4ById(subnetv4_id)
	subnetName := s4.Subnet

	oldPoolObj := handler.OrmGetPool(subnetv4_id, r.GetID())
	if len(oldPoolObj.BeginAddress) == 0 {
		return fmt.Errorf("Pool not exists, return")
	}
	oldPoolName := oldPoolObj.BeginAddress + "-" + oldPoolObj.EndAddress

	ormPool := dhcporm.Pool{
		Subnetv4ID:       ConvertStringToUint(r.GetParent().GetID()),
		BeginAddress:     r.BeginAddress,
		EndAddress:       r.EndAddress,
		ValidLifetime:    ConvertStringToInt(r.ValidLifetime),
		MaxValidLifetime: ConvertStringToInt(r.MaxValidLifetime),
		Gateway:          r.Gateway,
		DnsServer:        r.DnsServer,
	}

	log.Println(" *** begin to save db, pool.ID: ", r.GetID(), ", pool.subnetv4id: ", ormPool.Subnetv4ID)
	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&ormPool).Error; err != nil {
		return err
	}
	//send kafka msg
	req := pb.UpdateSubnetv4PoolReq{
		Oldpool:          oldPoolName,
		Subnet:           subnetName,
		Pool:             ormPool.BeginAddress + "-" + ormPool.EndAddress,
		Options:          []*pb.Option{},
		ValidLifetime:    strconv.Itoa(ormPool.ValidLifetime),
		MaxValidLifetime: strconv.Itoa(ormPool.MaxValidLifetime),
		Gateway:          r.Gateway,
		DnsServer:        r.DnsServer,
	}
	data, err := proto.Marshal(&req)
	if err != nil {
		log.Println("proto.Marshal error, ", err)
		return err
	}
	log.Println("begin to call SendDhcpCmd, update subnetv4 pool, req: ", req)
	if err := utils.SendDhcpCmd(data, dhcpv4agent.UpdateSubnetv4Pool); err != nil {
		log.Println("SendDhcpCmd error, ", err)
		return err
	}

	tx.Commit()
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
	if err := utils.SendDhcpCmd(data, dhcpv4agent.DeleteSubnetv4Pool); err != nil {
		log.Println("SendDhcpCmd error, ", err)
		return err
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

// get option name by name and ver
func (handler *PGDB) getOptionNamebyID(id string) *dhcporm.OrmOptionName {
	log.Println("in getOptionNamebyNameVer, id: ", id)

	var ormOpName dhcporm.OrmOptionName
	ormOpName.ID = ConvertStringToUint(id)
	handler.db.First(&ormOpName)

	log.Println("ormOpName: ", ormOpName)
	return &ormOpName
}

type BaseJsonOptionName struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    []*RestOptionNameConfig `json:"data"`
}

// get statistics group by v4/v6
func (handler *PGDB) GetOptionNameStatistics() *BaseJsonOptionName {
	//log.Println("in getOptionNameStatistics")

	rows, err := handler.db.Table("orm_option_names").Select("option_ver, count(*) as total").Group("option_ver").Rows()
	if err != nil {
		log.Println("group by error: ", err)
	}
	//var retArr OptionNameStatisticsRet
	var newRet BaseJsonOptionName
	//var retArr2 []*RestOptionNameConfig

	newRet.Status = "200"
	newRet.Message = "ok"

	for rows.Next() {
		//log.Println("--- rows : ", rows)

		var ret OptionNameStatistics
		if err := rows.Scan(&ret.OptionVer, &ret.Total); err != nil {
			log.Println("get from db error, err: ", err)
		}
		//log.Println("ret OptionVer: ", ret.OptionVer, ", total: ", ret.Total)
		var restOpName RestOptionNameConfig
		if ret.OptionVer == "v4" {
			//retArr.V4Num = ret.Total

			restOpName.OptionName = "DHCP"
			restOpName.OptionNum = ret.Total
			restOpName.OptionType = "IPv4"
			restOpName.OptionNotes = ""

			newRet.Data = append(newRet.Data, &restOpName)
		}
		if ret.OptionVer == "v6" {
			//retArr.V6Num = ret.Total

			restOpName.OptionName = "DHCPv6"
			restOpName.OptionNum = ret.Total
			restOpName.OptionType = "IPv6"
			restOpName.OptionNotes = ""

			newRet.Data = append(newRet.Data, &restOpName)
		}

	}
	//newRet.Data = retArr2

	log.Println("newRet.Data: ", newRet.Data)
	return &newRet
}

func (handler *PGDB) GetIPAddresses(subNetID string, ip string, hostName string, mac string) ([]*IPAddress, error) {
	log.Println("into dhcptb GetIPAddress, subNetID ", subNetID)
	//get the reservation address
	reservData := PGDBConn.OrmReservationList(subNetID)
	allData := make(map[string]dhcporm.IPAddress, 255)
	for _, a := range reservData {
		if a.ReservType == "hw-address" || a.ReservType == "client-id" {
			//get the stable address
			tmp := dhcporm.IPAddress{IP: a.IpAddress, AddressType: "stable"}
			allData[a.IpAddress] = tmp
		} else {
			tmp := dhcporm.IPAddress{IP: a.IpAddress, AddressType: "reserved"}
			allData[a.IpAddress] = tmp
		}
	}
	//get the pools under the subnet
	pools := PGDBConn.OrmPoolList(subNetID)
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
			tmp := dhcporm.IPAddress{IP: ip, AddressType: "dynamic"}
			allData[ip] = tmp
		}
	}
	//get manual address
	var manuals []dhcporm.ManualAddress
	if err := handler.db.Where("subnetv4_id = ?", subNetID).Find(&manuals).Error; err != nil {
		return nil, err
	}
	for _, v := range manuals {
		tmp := dhcporm.IPAddress{IP: v.IpAddress, AddressType: "manual"}
		allData[v.IpAddress] = tmp
	}
	//get the lease address for the subnet
	//first get the corespoding SubnetId in the table OrmSubnetv4
	var subnetv4 dhcporm.OrmSubnetv4
	if err:=handler.db.First(&subnetv4,subNetID).Error;err!=nil{
		return nil,err
	}
	leases,err := dhcpgrpc.GetLeases(strconv.Itoa(int(subnetv4.SubnetId)))
	if err!=nil{
		return nil,err
	}
	for _, l := range leases {
		var macAddr string
		for i := 0; i < len(l.HwAddress); i++ {
			var tmp string
			if i==0{
				tmp = fmt.Sprintf("%X", l.HwAddress[i])
			}else{
				tmp = fmt.Sprintf(":%X", l.HwAddress[i])
			}
			
			macAddr += tmp
		}
		tmp := dhcporm.IPAddress{IP: l.IpAddress, MacAddress: macAddr, AddressType: "lease", LeaseStartTime: l.Expire - int64(l.ValidLifetime), LeaseEndTime: l.Expire}
		allData[l.IpAddress] = tmp
	}
	subnet := PGDBConn.GetSubnetv4ById(subNetID)
	parts := strings.Split(subnet.Subnet, "/")
	beginNums := strings.Split(parts[0], ".")
	prefix := beginNums[0] + "." + beginNums[1] + "." + beginNums[2] + "."
	allSubnetAddress,err := handler.GetSubnetAllAddresses(subnet.Subnet)
	if err!= nil{
		return nil,err
	}
	var num int
	if num,err = strconv.Atoi(subNetID);err!= nil{
		return nil,err
	}	
	for i,v := range allSubnetAddress{
		if allData[v].AddressType == "" {
			tmp := dhcporm.IPAddress{IP: v, AddressType: "unused",Subnetv4ID:uint(num)}
			allData[prefix+strconv.Itoa(i)] = tmp
		}
	}
	var data []*IPAddress
	var filterData []*IPAddress
	var input []dhcporm.IPAddress
	for _, v := range allData {
		v.Subnetv4ID = uint(num)
		input = append(input, v)
	}
	if data, err = handler.UpdateIPAddresses(input); err != nil {
		return nil, err
	}
	if ip != "" {
		for _, v := range data {
			if v.IP == ip {
				filterData = append(filterData, v)
			}
		}
		data = filterData
		filterData = filterData[0:0]
	}

	if hostName != "" {
		for _, v := range data {
			if v.HostName == hostName {
				if v.IP == ip {
					filterData = append(filterData, v)
				}
			}
		}
		data = filterData
		filterData = filterData[0:0]
	}
	if mac != "" {
		for _, v := range data {
			if v.MacAddress == mac {
				if v.IP == ip {
					filterData = append(filterData, v)
				}
			}
		}
		data = filterData

	}

	return data, nil
}

func (handler *PGDB) GetSubnetAllAddresses(subnet string) ([]string, error) {
	var all []string
	var length int
	splits := strings.Split(subnet, "/")
	var ip net.IP
	if len(splits) < 2 {
		panic("wrong")
	}
	ip = net.ParseIP(string(splits[0]))
	if ip == nil {
		panic("wrong")
	}
	var err error
	if length, err = strconv.Atoi(string(splits[1])); err != nil {
		panic(err)
	}
	a := ip.To4()
	var max int
	max = int(math.Pow(2, float64(8-length%8)) - 1)
	if 0 == length/8 {
		for i := 0; i <= max; i++ {
			for j := 1; j <= 255; j++ {
				for k := 1; k <= 255; k++ {
					for m := 1; m <= 255; m++ {
						all = append(all, net.IPv4(a[0]+byte(i), a[1]+byte(j), a[2]+byte(k), a[3]+byte(m)).String())
					}
				}
			}
		}
	}
	if 1 == length/8 {
		for j := 0; j <= max; j++ {
			for k := 1; k <= 255; k++ {
				for m := 1; m <= 255; m++ {
					all = append(all, net.IPv4(a[0], a[1]+byte(j), a[2]+byte(k), a[3]+byte(m)).String())
				}
			}
		}
	}
	if 2 == length/8 {
		for k := 0; k <= max; k++ {
			for m := 1; m <= 255; m++ {
				all = append(all, net.IPv4(a[0], a[1], a[2]+byte(k), a[3]+byte(m)).String())
			}
		}
	}
	if 3 == length/8 {
		for m := 0; m <= max; m++ {
			all = append(all, net.IPv4(a[0], a[1], a[2], a[3]+byte(m)).String())
		}
	}
	return all, nil
}

func (handler *PGDB) UpdateIPAddresses(ipAddresses []dhcporm.IPAddress) ([]*IPAddress, error) {
	//create new data in the database
	tx := handler.db.Begin()
	defer tx.Rollback()
	for i, one := range ipAddresses {
		var tmp dhcporm.IPAddress
		if err := tx.Where("ip = ?", one.IP).Find(&tmp).Error; err != nil {
			if err := tx.Create(&ipAddresses[i]).Error; err != nil {
				return nil, err
			}
		} else {
			if one.MacAddress != "" {
				if err := tx.Model(&tmp).UpdateColumn("mac_address", one.MacAddress).Error; err != nil {
					return nil, err
				}
				ipAddresses[i].MacAddress = one.MacAddress
			}else{
				ipAddresses[i].MacAddress = tmp.MacAddress
			}
			if one.LeaseStartTime != 0 {
				if err := tx.Model(&tmp).UpdateColumn("lease_start_time", one.LeaseStartTime).Error; err != nil {
					return nil, err
				}
				ipAddresses[i].LeaseStartTime = one.LeaseStartTime
			}else{
				ipAddresses[i].LeaseStartTime = tmp.LeaseStartTime
			}
			if one.LeaseEndTime != 0 {
				if err := tx.Model(&tmp).UpdateColumn("lease_end_time", one.LeaseEndTime).Error; err != nil {
					return nil, err
				}
				ipAddresses[i].LeaseEndTime = one.LeaseEndTime
			}else{
				ipAddresses[i].LeaseEndTime = tmp.LeaseEndTime
			}
			if err := tx.Model(&tmp).UpdateColumn("address_type", one.AddressType).Error; err != nil {
				return nil, err
			}			
			ipAddresses[i].ID = tmp.ID
			ipAddresses[i].MacVender = tmp.MacVender
			ipAddresses[i].AddressType = tmp.AddressType
			ipAddresses[i].OperSystem = tmp.OperSystem
			ipAddresses[i].NetBIOSName = tmp.NetBIOSName
			ipAddresses[i].HostName = tmp.HostName
			ipAddresses[i].InterfaceID = tmp.InterfaceID
			ipAddresses[i].FingerPrint = tmp.FingerPrint
			ipAddresses[i].DeviceTypeFlag = tmp.DeviceTypeFlag
			ipAddresses[i].DeviceType = tmp.DeviceType
			ipAddresses[i].BusinessFlag = tmp.BusinessFlag
			ipAddresses[i].Business = tmp.Business
			ipAddresses[i].ChargePersonFlag = tmp.ChargePersonFlag
			ipAddresses[i].ChargePerson = tmp.ChargePerson
			ipAddresses[i].TelFlag = tmp.TelFlag
			ipAddresses[i].Tel = tmp.Tel
			ipAddresses[i].DepartmentFlag = tmp.DepartmentFlag
			ipAddresses[i].Department = tmp.Department
			ipAddresses[i].PositionFlag = tmp.PositionFlag
			ipAddresses[i].Position = tmp.Position

		}
	}
	tx.Commit()
	var data []*IPAddress
	for _, v := range ipAddresses {
		var tmp IPAddress
		tmp.SetID(strconv.Itoa(int(v.ID)))
		tmp.IP = v.IP
		tmp.MacAddress = v.MacAddress
		tmp.MacVender = v.MacVender
		tmp.AddressType = v.AddressType
		tmp.OperSystem = v.OperSystem
		tmp.NetBIOSName = v.NetBIOSName
		tmp.HostName = v.HostName
		tmp.InterfaceID = v.InterfaceID
		tmp.FingerPrint = v.FingerPrint
		tmp.LeaseStartTime = v.LeaseStartTime
		tmp.LeaseEndTime = v.LeaseEndTime
		tmp.DeviceTypeFlag = v.DeviceTypeFlag
		tmp.DeviceType = v.DeviceType
		tmp.BusinessFlag = v.BusinessFlag
		tmp.Business = v.Business
		tmp.ChargePersonFlag = v.ChargePersonFlag
		tmp.ChargePerson = v.ChargePerson
		tmp.TelFlag = v.TelFlag
		tmp.Tel = v.Tel
		tmp.DepartmentFlag = v.DepartmentFlag
		tmp.Department = v.Department
		tmp.PositionFlag = v.PositionFlag
		tmp.Position = v.Position
		data = append(data, &tmp)
	}
	return data, nil
}

func (handler *PGDB) UpdateIPAddress(one *IPAddress,subnetid string) error {
	var tmp dhcporm.IPAddress
	var err error
	var num int
	tx := handler.db.Begin()
	defer tx.Rollback()	
	if num, err = strconv.Atoi(one.ID); err != nil {
		return err
	}
	tmp.ID = uint(num)
	if err:=tx.First(&tmp, one.GetID()).Error;err!=nil{
		return err
	}
	tmp.IP = one.IP
	tmp.MacAddress = one.MacAddress
	tmp.MacVender = one.MacVender
	tmp.OperSystem = one.OperSystem
	tmp.NetBIOSName = one.NetBIOSName
	tmp.HostName = one.HostName
	tmp.InterfaceID = one.InterfaceID
	tmp.FingerPrint = one.FingerPrint
	tmp.LeaseStartTime = one.LeaseStartTime
	tmp.LeaseEndTime = one.LeaseEndTime
	tmp.DeviceTypeFlag = one.DeviceTypeFlag
	tmp.DeviceType = one.DeviceType
	tmp.BusinessFlag = one.BusinessFlag
	tmp.Business = one.Business
	tmp.ChargePersonFlag = one.ChargePersonFlag
	tmp.ChargePerson = one.ChargePerson
	tmp.TelFlag = one.TelFlag
	tmp.Tel = one.Tel
	tmp.DepartmentFlag = one.DepartmentFlag
	tmp.Department = one.Department
	tmp.PositionFlag = one.PositionFlag
	tmp.Position = one.Position
	if num,err = strconv.Atoi(subnetid);err!=nil{
		return err
	}
	tmp.Subnetv4ID = uint(num)
	if err := tx.Save(&tmp).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}
func (handler *PGDB) UpdateIPAttrAppend(ipAddressID string, attrAppend *IPAttrAppend) error {
	tx := handler.db.Begin()
	defer tx.Rollback()
	var one dhcporm.IPAddress
	if err := tx.First(&one, ipAddressID).Error; err != nil {
		return err
	}
	one.DeviceTypeFlag = attrAppend.DeviceTypeFlag
	one.BusinessFlag = attrAppend.BusinessFlag
	one.ChargePersonFlag = attrAppend.ChargePersonFlag
	one.TelFlag = attrAppend.TelFlag
	one.DepartmentFlag = attrAppend.DepartmentFlag
	one.PositionFlag = attrAppend.PositionFlag
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}
func (handler *PGDB) GetIPAttrAppend(id string) (*IPAttrAppend, error) {
	var tmp dhcporm.IPAddress
	if err := handler.db.First(&tmp, id).Error; err != nil {
		return nil, err
	}
	var one IPAttrAppend
	one.SetID(id)
	one.DeviceTypeFlag = tmp.DeviceTypeFlag
	one.BusinessFlag = tmp.BusinessFlag
	one.ChargePersonFlag = tmp.ChargePersonFlag
	one.TelFlag = tmp.TelFlag
	one.DepartmentFlag = tmp.DepartmentFlag
	one.PositionFlag = tmp.PositionFlag
	return &one, nil
}
