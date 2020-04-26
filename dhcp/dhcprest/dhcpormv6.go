package dhcprest

import (
	"fmt"
	"log"
	"strconv"

	"github.com/linkingthing/ddi/utils"

	dnsapi "github.com/linkingthing/ddi/dns/restfulapi"

	"github.com/linkingthing/ddi/dhcp/agent/dhcpv6agent"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/pb"

	"github.com/jinzhu/gorm"

	"github.com/linkingthing/ddi/dhcp/dhcporm"
)

func (handler *PGDB) GetSubnetv6ByName(db *gorm.DB, name string) *dhcporm.OrmSubnetv6 {
	log.Println("in GetSubnetv6ByName, name: ", name)

	var subnetv6 dhcporm.OrmSubnetv6
	db.Where(&dhcporm.OrmSubnetv6{Subnet: name}).Find(&subnetv6)

	return &subnetv6
}

func (handler *PGDB) Subnetv6List(search *SubnetSearch) []dhcporm.OrmSubnetv6 {
	var subnetv6s []dhcporm.OrmSubnetv6
	if search != nil && search.DhcpVer != "" {
		subnet := handler.getOrmSubnetv6BySubnet(search.Subnet)
		subnetv6s = append(subnetv6s, subnet)
		log.Println("in Subnetv6List, search ret subnet: ", subnet)
	} else {
		query := handler.db.Find(&subnetv6s)
		if query.Error != nil {
			log.Print(query.Error.Error())
		}
	}

	for k, v := range subnetv6s {
		//log.Println("k: ", k, ", v: ", v)
		//log.Println("in Subnetv4List, v.ID: ", v.ID)
		if len(v.Name) > 0 && len(v.ZoneName) == 0 {
			subnetv6s[k].ZoneName = v.Name
		}
		rsv6 := []*dhcporm.OrmReservationv6{}
		if err := handler.db.Where("subnetv6_id = ?", strconv.Itoa(int(v.ID))).Find(&rsv6).Error; err != nil {
			log.Print(err)
		}
		subnetv6s[k].Reservationv6s = rsv6
	}

	return subnetv6s
}

func (handler *PGDB) getOrmSubnetv6BySubnet(subnet string) dhcporm.OrmSubnetv6 {
	log.Println("in getOrmSubnetv6BySubnet, subnet: ", subnet)

	var subnetv6 dhcporm.OrmSubnetv6
	handler.db.Where(&dhcporm.OrmSubnetv6{Subnet: subnet}).Find(&subnetv6)

	return subnetv6
}

func (handler *PGDB) GetSubnetv6ById(id string) *dhcporm.OrmSubnetv6 {
	log.Println("in dhcp/dhcprest/GetSubnetv6ById, id: ", id)
	dbId := ConvertStringToUint(id)

	subnetv6 := dhcporm.OrmSubnetv6{}
	subnetv6.ID = dbId
	handler.db.Preload("Reservationv6s").First(&subnetv6)

	return &subnetv6
}

//get Currently maxId from Kea config
func (handler *PGDB) GetSubnetv6MaxId() uint32 {
	var maxId uint32

	row := handler.db.Table("subnetv6s").Select("MAX(subnetId)").Row()
	row.Scan(&maxId)
	log.Println("in GetSubnetMaxId, maxId: ", maxId)
	log.Println("in GetSubnetMaxId, utils.Subnetv6MaxId: ", utils.Subnetv6MaxId)
	if utils.Subnetv6MaxId <= maxId {
		utils.Subnetv6MaxId = maxId
	}
	return maxId
}

//return (new inserted id, error)
func (handler *PGDB) CreateSubnetv6(s *RestSubnetv6) (dhcporm.OrmSubnetv6, error) {
	log.Println("into CreateSubnetv6, name, subnet, validLifetime: ")
	var s6 = dhcporm.OrmSubnetv6{
		Dhcpv6ConfId:     1,
		Name:             s.Name,
		Subnet:           s.Subnet,
		ZoneName:         s.Name,
		DhcpEnable:       1,
		ValidLifetime:    s.ValidLifetime,
		MaxValidLifetime: s.MaxValidLifetime,
		DnsServer:        s.DnsServer,
	}
	maxId := handler.GetSubnetv6MaxId()
	s6.SubnetId = maxId + 1

	if len(s6.Name) > 0 && len(s6.ZoneName) == 0 {
		s6.ZoneName = s6.Name
	}
	query := handler.db.Create(&s6)

	if query.Error != nil {
		return s6, fmt.Errorf("create subnet error, subnet name: ")
	}
	var last dhcporm.OrmSubnetv6
	query.Last(&last)
	log.Println("query.value: ", query.Value, ", id: ", last.ID)

	//send msg to kafka queue, which is read by dhcp server
	req := pb.CreateSubnetv6Req{
		Subnet:        s6.Subnet,
		Id:            s6.SubnetId,
		ValidLifetime: s6.ValidLifetime,
		DnsServer:     s6.DnsServer,
	}
	log.Println("pb.CreateSubnetv6Req req: ", req)

	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	utils.SendDhcpv6Cmd(data, dhcpv6agent.CreateSubnetv6)

	log.Println(" in CreateSubnetv6, last: ", last)
	return last, nil
}

func (handler *PGDB) OrmUpdateSubnetv6(subnetv6 *RestSubnetv6) error {
	log.Println("into dhcporm, OrmUpdateSubnetv6, Subnet: ", subnetv6.Subnet)

	dbS6 := dhcporm.OrmSubnetv6{}
	dbS6.Name = subnetv6.Name
	dbS6.ValidLifetime = subnetv6.ValidLifetime
	id, err := strconv.Atoi(subnetv6.ID)
	if err != nil {
		log.Println("subnetv6.ID error, id: ", subnetv6.ID)
		return err
	}
	dbS6.ID = uint(id)
	dbS6.DhcpEnable = subnetv6.DhcpEnable
	dbS6.ZoneName = subnetv6.ZoneName
	if len(dbS6.Name) > 0 && len(dbS6.ZoneName) == 0 {
		dbS6.ZoneName = dbS6.Name
	}
	dbS6.DnsEnable = subnetv6.DnsEnable
	dbS6.Note = subnetv6.Note
	dbS6.DnsServer = subnetv6.DnsServer

	//get subnet name from db
	getOrmS6 := handler.GetSubnetv6ById(subnetv6.ID)
	dbS6.Subnet = getOrmS6.Subnet

	if subnetv6.DnsEnable > 0 {
		if len(subnetv6.ViewId) == 0 {
			log.Println("Error subnetv6 viewId is null, return")
			//return fmt.Errorf("zone is enabled, viewId is null")
		}
		zone := dnsapi.Zone{Name: subnetv6.ZoneName, ZoneType: "master"}

		dnsapi.DBCon.CreateZone(&zone, subnetv6.ViewId)
	}

	log.Println("begin to save db, dbS6.ID: ", dbS6.ID)
	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&dbS6).Error; err != nil {
		return err
	}

	//todo send kafka msg
	req := pb.UpdateSubnetv6Req{Id: subnetv6.ID, Subnet: subnetv6.Subnet, ValidLifetime: subnetv6.ValidLifetime}
	data, err := proto.Marshal(&req)
	if err != nil {
		log.Println("proto.Marshal error, ", err)
		return err
	}
	log.Println("begin to call SendDhcpv6Cmd, update subnetv6")
	if err := utils.SendDhcpv6Cmd(data, dhcpv6agent.UpdateSubnetv6); err != nil {
		log.Println("SendCmdDhcpv6 error, ", err)
		return err
	}

	tx.Commit()
	return nil
}

func (handler *PGDB) DeleteSubnetv6(id string) error {

	var ormS6 dhcporm.OrmSubnetv6

	tx := handler.db.Begin()
	defer tx.Rollback()

	if err := tx.First(&ormS6, id).Error; err != nil {
		return fmt.Errorf("unknown subnetv6 with ID %s, %w", id, err)
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	ormS6.ID = uint(num)

	if err := tx.Unscoped().Delete(&ormS6).Error; err != nil {
		return err
	}
	req := pb.DeleteSubnetv4Req{Id: id, Subnet: ormS6.Subnet}

	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := utils.SendDhcpv6Cmd(data, dhcpv6agent.DeleteSubnetv6); err != nil {
		log.Println("SendCmdDhcpv6 error, ", err)
		return err
	}
	tx.Commit()

	return nil
}

// --- old
func (handler *PGDB) GetSubnetv6(db *gorm.DB, id string) *dhcporm.OrmSubnetv6 {
	dbId := ConvertStringToUint(id)

	subnetv6 := dhcporm.OrmSubnetv6{}
	subnetv6.ID = dbId
	db.Preload("Reservations").First(&subnetv6)

	return &subnetv6
}

//
//func (handler *PGDB) CreateSubnetv6Old(db *gorm.DB, name string, validLifetime string) error {
//	var subnet = dhcporm.OrmSubnetv6{
//		Dhcpv6ConfId:  1,
//		Subnet:        name,
//		ValidLifetime: validLifetime,
//		//DhcpVer:       Dhcpv4Ver,
//	}
//
//	query := db.Create(&subnet)
//
//	if query.Error != nil {
//		return fmt.Errorf("create subnet error, subnet name: " + name)
//	}
//
//	return nil
//}

func (handler *PGDB) OrmPoolv6List(subnetId string) []*dhcporm.Poolv6 {
	log.Println("in dhcprest, OrmPoolv6List, subnetId: ", subnetId)
	var poolv6s []*dhcporm.Poolv6
	var ps []dhcporm.Poolv6

	subnetIdUint := ConvertStringToUint(subnetId)
	if err := handler.db.Where("subnetv6_id = ?", subnetIdUint).Find(&ps).Error; err != nil {
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
		p2.Subnetv6ID = subnetIdUint
		p2.BeginAddress = p.BeginAddress
		p2.EndAddress = p.EndAddress

		poolv6s = append(poolv6s, &p2)
	}

	return poolv6s
}
func (handler *PGDB) OrmGetPoolv6(subnetId string, pool_id string) *dhcporm.Poolv6 {
	log.Println("into rest OrmGetPoolv6, subnetId: ", subnetId, "pool_id: ", pool_id)
	dbRsvId := ConvertStringToUint(pool_id)

	poolv6 := dhcporm.Poolv6{}
	if err := handler.db.First(&poolv6, int(dbRsvId)).Error; err != nil {
		//fmt.Errorf("get reservation error, subnetId: ", subnetId, " reservation id: ", rsv_id)
		return nil
	}

	return &poolv6
}

func (handler *PGDB) OrmCreatePoolv6(subnetv6_id string, r *RestPoolv6) (dhcporm.Poolv6, error) {
	log.Println("into OrmCreatePoolv6, r: ", r, ", subnetv6_id: ", subnetv6_id)

	sid, err := strconv.Atoi(subnetv6_id)
	if err != nil {
		log.Println("OrmCreatePoolv6, sid error: ", subnetv6_id)
	}
	var ormPoolv6 dhcporm.Poolv6
	ormPoolv6 = dhcporm.Poolv6{
		Subnetv6ID:       uint(sid),
		BeginAddress:     r.BeginAddress,
		EndAddress:       r.EndAddress,
		OptionData:       []dhcporm.Option{},
		ValidLifetime:    ConvertStringToInt(r.ValidLifetime),
		MaxValidLifetime: ConvertStringToInt(r.MaxValidLifetime),
	}

	var pool = pb.Pools{
		Pool:             r.BeginAddress + "-" + r.EndAddress,
		Options:          []*pb.Option{},
		ValidLifetime:    r.ValidLifetime,
		MaxValidLifetime: r.MaxValidLifetime,

		//DhcpVer:       Dhcpv4Ver,
	}

	//get subnet by subnetv4_id
	ormSubnetv6 := handler.GetSubnetv6ById(subnetv6_id)
	s6Subnet := ormSubnetv6.Subnet

	//todo: post kafka msg to dhcp agent
	pools := []*pb.Pools{}
	pools = append(pools, &pool)
	req := pb.CreateSubnetv6PoolReq{
		Id:               subnetv6_id,
		Subnet:           s6Subnet,
		Pool:             pools,
		ValidLifetime:    pool.ValidLifetime,
		MaxValidLifetime: pool.MaxValidLifetime,
	}
	log.Println("OrmCreatePool, req: ", req)
	data, err := proto.Marshal(&req)
	if err != nil {
		return ormPoolv6, err
	}
	if err := utils.SendDhcpv6Cmd(data, dhcpv6agent.CreateSubnetv6Pool); err != nil {
		log.Println("SendCmdDhcpv6 error, ", err)
		return ormPoolv6, err
	}
	//end of todo

	query := handler.db.Create(&ormPoolv6)
	if query.Error != nil {
		return dhcporm.Poolv6{}, fmt.Errorf("CreatePool error, begin address: " +
			r.BeginAddress + ", end adderss: " + r.EndAddress)
	}

	return ormPoolv6, nil
}

func (handler *PGDB) OrmUpdatePoolv6(subnetv6_id string, r *RestPoolv6) error {

	log.Println("into dhcporm, OrmUpdatePool, id: ", r.GetID())

	//get subnetv4 name
	s6 := handler.GetSubnetv6ById(subnetv6_id)
	subnetName := s6.Subnet

	//oldPoolName := r.BeginAddress + "-" + r.EndAddress
	//search subnet, if not exist, return error
	oldPoolObj := handler.OrmGetPoolv6(subnetv6_id, r.GetID())
	if oldPoolObj == nil {
		return fmt.Errorf("Pool not exists, return")
	}
	oldPoolName := oldPoolObj.BeginAddress + "-" + oldPoolObj.EndAddress

	ormPool := dhcporm.Poolv6{}
	ormPool.ID = ConvertStringToUint(r.GetID())
	ormPool.BeginAddress = r.BeginAddress
	ormPool.EndAddress = r.EndAddress
	ormPool.Subnetv6ID = ConvertStringToUint(subnetv6_id)
	ormPool.ValidLifetime = ConvertStringToInt(r.ValidLifetime)
	ormPool.MaxValidLifetime = ConvertStringToInt(r.MaxValidLifetime)

	log.Println("begin to save db, pool.ID: ", r.GetID(), ", pool.subnetv6id: ", ormPool.Subnetv6ID)

	tx := handler.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&ormPool).Error; err != nil {
		return err
	}
	//todo send kafka msg
	req := pb.UpdateSubnetv6PoolReq{
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
	log.Println("begin to call SendDhcpv6Cmd, update subnetv4 pool, req: ", req)
	if err := utils.SendDhcpv6Cmd(data, dhcpv6agent.UpdateSubnetv6Pool); err != nil {
		log.Println("SendDhcpv6Cmd error, ", err)
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

func (handler *PGDB) OrmDeletePoolv6(id string) error {
	log.Println("into dhcprest OrmDeletePoolv6, id ", id)

	var ormSubnetv6 dhcporm.OrmSubnetv6
	var ormPool dhcporm.Poolv6

	tx := handler.db.Begin()
	defer tx.Rollback()

	if err := tx.First(&ormPool, id).Error; err != nil {
		return fmt.Errorf("unknown subnetv6pool with ID %s, %w", id, err)
	}
	log.Println("subnetv6 id: ", ormPool.Subnetv6ID)

	if err := tx.First(&ormSubnetv6, ormPool.Subnetv6ID).Error; err != nil {
		return fmt.Errorf("unknown subnetv6 with ID %s, %w", ormPool.Subnetv6ID, err)
	}
	num, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	ormPool.ID = uint(num)

	if err := tx.Unscoped().Delete(&ormPool).Error; err != nil {
		return err
	}
	req := pb.DeleteSubnetv6PoolReq{
		Subnet: ormSubnetv6.Subnet,
		Pool:   ormPool.BeginAddress + "-" + ormPool.EndAddress,
	}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := utils.SendDhcpv6Cmd(data, dhcpv6agent.DeleteSubnetv6Pool); err != nil {
		log.Println("SendDhcpv6Cmd error, ", err)
		return err
	}
	tx.Commit()

	return nil
}
