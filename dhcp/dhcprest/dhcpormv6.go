package dhcprest

import (
	"fmt"
	"log"
	"strconv"

	"github.com/linkingthing/ddi/dhcp/agent/dhcpv6agent"

	"github.com/golang/protobuf/proto"
	"github.com/linkingthing/ddi/dhcp"
	"github.com/linkingthing/ddi/pb"

	"github.com/jinzhu/gorm"

	"github.com/linkingthing/ddi/dhcp/dhcporm"
)

func (handler *PGDB) Subnetv6List() []dhcporm.OrmSubnetv6 {
	var subnetv6s []dhcporm.OrmSubnetv6

	query := handler.db.Find(&subnetv6s)
	if query.Error != nil {
		log.Print(query.Error.Error())
	}

	for k, v := range subnetv6s {
		//log.Println("k: ", k, ", v: ", v)
		//log.Println("in Subnetv4List, v.ID: ", v.ID)

		rsv6 := []*dhcporm.OrmReservationv6{}
		if err := handler.db.Where("subnetv6_id = ?", strconv.Itoa(int(v.ID))).Find(&rsv6).Error; err != nil {
			log.Print(err)
		}
		subnetv6s[k].Reservationv6s = rsv6
	}

	return subnetv6s
}

func (handler *PGDB) getSubnetv6BySubnet(subnet string) *dhcporm.OrmSubnetv6 {
	log.Println("in getSubnetv6BySubnet, subnet: ", subnet)

	var subnetv6 dhcporm.OrmSubnetv6
	handler.db.Where(&dhcporm.OrmSubnetv6{Subnet: subnet}).Find(&subnetv6)

	return &subnetv6
}

func (handler *PGDB) GetSubnetv6ById(id string) *dhcporm.OrmSubnetv6 {
	log.Println("in dhcp/dhcprest/GetSubnetv6ById, id: ", id)
	dbId := ConvertStringToUint(id)

	subnetv6 := dhcporm.OrmSubnetv6{}
	subnetv6.ID = dbId
	handler.db.Preload("Reservationv6s").First(&subnetv6)

	return &subnetv6
}

//return (new inserted id, error)
func (handler *PGDB) CreateSubnetv6(s *RestSubnetv6) (dhcporm.OrmSubnetv6, error) {
	log.Println("into CreateSubnetv6, name, subnet, validLifetime: ")
	var s6 = dhcporm.OrmSubnetv6{
		Dhcpv6ConfId:  1,
		Name:          s.Name,
		Subnet:        s.Subnet,
		ValidLifetime: s.ValidLifetime,
		//Gateway:       s.Gateway,
		//DhcpVer:       Dhcpv4Ver,
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
		Subnet:        s.Subnet,
		Id:            strconv.Itoa(int(last.ID)),
		ValidLifetime: s.ValidLifetime,
		//Gateway:       s.Gateway,
	}
	log.Println("pb.CreateSubnetv6Req req: ", req)

	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	dhcp.SendDhcpv6Cmd(data, dhcpv6agent.CreateSubnetv6)

	log.Println(" in CreateSubnetv6, last: ", last)
	return last, nil
}

func (handler *PGDB) OrmUpdateSubnetv6(subnetv6 *RestSubnetv6) error {
	log.Println("into dhcporm, OrmUpdateSubnetv6, Subnet: ", subnetv6.Subnet)

	dbS6 := dhcporm.OrmSubnetv6{}
	//dbS4.SubnetId = subnetv4.ID
	dbS6.Subnet = subnetv6.Subnet
	dbS6.Name = subnetv6.Name
	dbS6.ValidLifetime = subnetv6.ValidLifetime
	id, err := strconv.Atoi(subnetv6.ID)
	if err != nil {
		log.Println("subnetv6.ID error, id: ", subnetv6.ID)
		return err
	}
	dbS6.ID = uint(id)

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
	if err := dhcp.SendDhcpv6Cmd(data, dhcpv6agent.UpdateSubnetv6); err != nil {
		log.Println("SendCmdDhcpv6 error, ", err)
		return err
	}

	tx.Commit()
	return nil
}

func (handler *PGDB) DeleteSubnetv6(id string) error {
	log.Println("into dhcprest DeleteSubnetv6, id ", id)

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
	log.Println("DeleteSubnetv6() req: ", req)
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := dhcp.SendDhcpv6Cmd(data, dhcpv6agent.DeleteSubnetv6); err != nil {
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
