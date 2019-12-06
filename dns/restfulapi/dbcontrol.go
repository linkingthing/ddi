package dnscontroller

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	tb "github.com/linkingthing/ddi/dns/cockroachtables"
	"github.com/linkingthing/ddi/pb"
	"strconv"
)

const (
	STARTDNS   = "StartDNS"
	STOPDNS    = "StopDNS"
	CREATEACL  = "CreateACL"
	DELETEACL  = "DeleteACL"
	CREATEVIEW = "CreateView"
	UPDATEVIEW = "UpdateView"
	DELETEVIEW = "DeleteView"
	CREATEZONE = "CreateZone"
	DELETEZONE = "DeleteZone"
	CREATERR   = "CreateRR"
	UPDATERR   = "UpdateRR"
	DELETERR   = "DeleteRR"
)

var DBCon *DBController

type DBController struct {
	db *gorm.DB
}

func NewDBController() *DBController {
	one := &DBController{}
	const addr = "postgresql://maxroach@localhost:26257/bank?ssl=true&sslmode=require&sslrootcert=/root/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/cockroach-v19.2.0/certs/client.maxroach.crt"
	var err error
	one.db, err = gorm.Open("postgres", addr)
	if err != nil {
		panic(err)
	}
	one.db.AutoMigrate(&tb.DBACL{})
	one.db.AutoMigrate(&tb.DBIP{})
	one.db.AutoMigrate(&tb.DBView{})
	one.db.AutoMigrate(&tb.DBZone{})
	one.db.AutoMigrate(&tb.DBRR{})
	return one
}
func (controller *DBController) Close() {
	controller.db.Close()
}

/*func init() {
	DBCon = NewDBController()
}*/

func (controller *DBController) CreateACL(aCL *ACL) (tb.DBACL, error) {
	//create new data in the database
	var one tb.DBACL
	one.Name = aCL.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	var dbACLs []tb.DBACL
	if err := tx.Where("name = ?", aCL.Name).Find(&dbACLs).Error; err != nil {
		return tb.DBACL{}, err
	}
	if len(dbACLs) > 0 {
		return tb.DBACL{}, fmt.Errorf("the name %s of acl exists", aCL.Name)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.DBACL{}, err
	}
	var last tb.DBACL
	tx.Last(&last)
	var dBIP tb.DBIP
	for _, iP := range aCL.IPs {
		dBIP.IP = iP
		dBIP.ACLID = strconv.Itoa(int(last.ID))
		if err := tx.Create(&dBIP).Error; err != nil {
			return last, nil
		}
	}
	var iPs []string
	for _, iP := range aCL.IPs {
		iPs = append(iPs, iP)
	}
	req := pb.CreateACLReq{ACLName: aCL.Name, ACLID: aCL.GetID(), IPs: iPs}
	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	if err := SendCmd(data, CREATEACL); err != nil {
		return last, err
	}
	tx.Commit()
	return last, nil
}

func (controller *DBController) DeleteACL(id string) error {
	var err error
	if _, err := controller.GetACL(id); err != nil {
		return fmt.Errorf("unknown ACL with ID %s", id)
	}
	//delete the ips and acl from the database
	var aCLDB tb.DBACL
	var num int
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	aCLDB.ID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Delete(&aCLDB).Error; err != nil {
		return err
	}
	if err := tx.Where("acl_id = ?", id).Delete(&tb.DBIP{}).Error; err != nil {
		return err
	}
	req := pb.DeleteACLReq{ACLID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := SendCmd(data, DELETEACL); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetACL(id string) (*ACL, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var a tb.DBACL
	var index int
	var err error
	if index, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	if err := tx.First(&a, index).Error; err != nil {
		return nil, err
	}
	aCL := ACL{}
	aCL.SetID(id)
	aCL.Name = a.Name
	var iPs []tb.DBIP
	if err := tx.Where("acl_id = ?", id).Find(&iPs).Error; err != nil {
		return nil, err
	}
	fmt.Println(iPs)
	for _, dBIP := range iPs {
		aCL.IPs = append(aCL.IPs, dBIP.IP)
	}
	aCL.Type = "acl"
	aCL.SetCreationTimestamp(a.CreatedAt)
	return &aCL, nil
}

func (controller *DBController) UpdateACL(aCL *ACL) error {
	var one tb.DBACL
	var num int
	var err error
	if num, err = strconv.Atoi(aCL.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = aCL.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Where("acl_id = ?", aCL.GetID()).Delete(&tb.DBIP{}).Error; err != nil {
		return err
	}
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	var iPDB tb.DBIP
	for _, ip := range aCL.IPs {
		iPDB.IP = ip
		iPDB.ACLID = aCL.GetID()
		if err := tx.Create(&iPDB).Error; err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetACLs() []*ACL {
	var aCLs []*ACL
	var aCLDBs []tb.DBACL
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Find(&aCLDBs).Error; err != nil {
		return nil
	}
	for _, aCLDB := range aCLDBs {
		var aCL ACL
		aCL.SetID(strconv.Itoa(int(aCLDB.ID)))
		aCL.Name = aCLDB.Name
		aCL.SetCreationTimestamp(aCLDB.CreatedAt)
		var iPDBs []tb.DBIP
		if err := tx.Where("acl_id = ?", aCLDB.ID).Find(&iPDBs).Error; err == nil {
			for _, iP := range iPDBs {
				aCL.IPs = append(aCL.IPs, iP.IP)
			}
		}
		aCLs = append(aCLs, &aCL)
	}
	return aCLs
}

func (controller *DBController) CreateView(view *View) (tb.DBView, error) {
	//create new data in the database
	var one tb.DBView
	one.Name = view.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	//adjust the priority
	var allView []tb.DBView
	if err := tx.Find(&allView).Error; err != nil {
		return tb.DBView{}, err
	}
	one.Priority = view.Priority
	if len(allView)+1 < view.Priority {
		one.Priority = len(allView) + 1
	} else if view.Priority < 0 {
		one.Priority = 1
	}
	for i, viewDB := range allView {
		if viewDB.Priority >= one.Priority {
			allView[i].Priority++
		} else {
			allView = append(allView[:i], allView[i+1:]...)
		}
	}
	for _, viewDB := range allView {
		if err := tx.Model(&viewDB).UpdateColumn("priority", viewDB.Priority).Error; err != nil {
			return tb.DBView{}, err
		}
	}
	var dbViews []tb.DBView
	//check wether the view is exists.
	if err := tx.Where("name = ?", view.Name).Find(&dbViews).Error; err != nil {
		return tb.DBView{}, err
	}
	if len(dbViews) > 0 {
		return tb.DBView{}, fmt.Errorf("the name %s of view has exists!", view.Name)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.DBView{}, err
	}
	var last tb.DBView
	tx.Last(&last)
	var aclids []string
	for _, id := range view.ACLIDs {
		aclids = append(aclids, id)
		//check wether the acl is valid
		var tmpACL tb.DBACL
		if err := tx.Where("view_id = ''").First(&tmpACL, id).Error; err != nil {
			return tb.DBView{}, fmt.Errorf("id %s of acl not exists, %w", id, err)
		}
		//update the acl's view_id
		if err := tx.Model(&tb.DBACL{}).Where("id = ?", id).Update("view_id", strconv.Itoa(int(last.ID))).Error; err != nil {
			return tb.DBView{}, err
		}
	}
	req := pb.CreateViewReq{ViewName: view.Name, ViewID: strconv.Itoa(int(last.ID)), Priority: int32(view.Priority), ACLIDs: aclids}
	data, err := proto.Marshal(&req)
	if err != nil {
		return tb.DBView{}, err
	}
	if err := SendCmd(data, CREATEVIEW); err != nil {
		return tb.DBView{}, err
	}
	tx.Commit()
	return last, nil
}

func (controller *DBController) DeleteView(id string) error {
	var err error
	if _, err := controller.GetView(id); err != nil {
		return fmt.Errorf("unknown View with ID %s", id)
	}
	var viewDB tb.DBView
	var num int
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	viewDB.ID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	var zonesDB []tb.DBZone
	if err := tx.Where("view_id = ?", id).Find(&zonesDB).Error; err != nil {
		return err
	}
	for _, zone := range zonesDB {
		//delete rr
		if err := tx.Where("zone_id = ?", strconv.Itoa(int(zone.ID))).Delete(&tb.DBRR{}).Error; err != nil {
			return err
		}
	}
	//delete zones
	if err := tx.Where("view_id = ?", id).Delete(&tb.DBZone{}).Error; err != nil {
		return err
	}
	var aclsDB []tb.DBACL
	if err := tx.Where("view_id = ?", id).Find(&aclsDB).Error; err != nil {
		return err
	}
	for _, acl := range aclsDB {
		//delete the ips
		if err := tx.Where("acl_id = ?", strconv.Itoa(int(acl.ID))).Delete(&tb.DBIP{}).Error; err != nil {
			return err
		}
	}
	//delete the acls
	if err := tx.Where("view_id = ?", id).Delete(&tb.DBACL{}).Error; err != nil {
		return err
	}
	//delete the view from the database
	if err := tx.Delete(&viewDB).Error; err != nil {
		return err
	}
	req := pb.DeleteViewReq{ViewID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := SendCmd(data, DELETEVIEW); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetView(id string) (*View, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var a tb.DBView
	var index int
	var err error
	if index, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	if err := tx.First(&a, index).Error; err != nil {
		return nil, err
	}
	view := View{}
	view.SetID(id)
	view.Name = a.Name
	view.Priority = a.Priority
	view.Type = "view"
	view.SetCreationTimestamp(a.CreatedAt)
	var acls []tb.DBACL
	if err := tx.Where("view_id = ?", id).Find(&acls).Error; err != nil {
		return nil, err
	}
	var ids []string
	for _, acl := range acls {
		ids = append(ids, strconv.Itoa(int(acl.ID)))
	}
	view.ACLIDs = ids
	var zones []tb.DBZone
	if err := tx.Where("view_id = ?", id).Find(&zones).Error; err != nil {
		return nil, err
	}
	for _, dbZone := range zones {
		var zone Zone
		zone.SetID(strconv.Itoa(int(dbZone.ID)))
		zone.Name = dbZone.Name
		zone.ZoneFile = dbZone.ZoneFile
		view.zones = append(view.zones, &zone)
	}
	return &view, nil
}

func (controller *DBController) GetViews() []*View {
	var views []*View
	var viewDBs []tb.DBView
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Find(&viewDBs).Error; err != nil {
		return nil
	}
	for _, viewDB := range viewDBs {
		var view View
		view.SetID(strconv.Itoa(int(viewDB.ID)))
		view.Name = viewDB.Name
		view.Priority = viewDB.Priority
		view.SetCreationTimestamp(viewDB.CreatedAt)
		var acls []tb.DBACL
		if err := tx.Where("view_id = ?", view.GetID()).Find(&acls).Error; err != nil {
			return nil
		}
		for _, acl := range acls {
			view.ACLIDs = append(view.ACLIDs, strconv.Itoa(int(acl.ID)))
		}
		var zonesDB []tb.DBZone
		if err := tx.Where("view_id = ?", view.GetID()).Find(&zonesDB).Error; err != nil {
			return nil
		}
		for _, zoneDB := range zonesDB {
			zone := &Zone{}
			zone.SetID(strconv.Itoa(int(zoneDB.ID)))
			zone.SetCreationTimestamp(zoneDB.CreatedAt)
			zone.Name = zoneDB.Name
			zone.ZoneFile = zoneDB.ZoneFile
			view.zones = append(view.zones, zone)
		}

		views = append(views, &view)
	}
	return views
}

//////////////

func (controller *DBController) CreateZone(zone *Zone, viewID string) (tb.DBZone, error) {
	//create new data in the database
	var one tb.DBZone
	one.Name = zone.Name
	one.ZoneFile = zone.ZoneFile
	one.ViewID = viewID
	tx := controller.db.Begin()
	defer tx.Rollback()
	var dbZones []tb.DBZone
	if err := tx.Where("name = ?", zone.Name).Find(&dbZones).Error; err != nil {
		return tb.DBZone{}, err
	}
	if len(dbZones) > 0 {
		return tb.DBZone{}, fmt.Errorf("the name %s of zone has exists!", zone.Name)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.DBZone{}, err
	}
	var last tb.DBZone
	tx.Last(&last)
	req := pb.CreateZoneReq{ViewID: viewID, ZoneName: zone.Name, ZoneID: strconv.Itoa(int(last.ID)), ZoneFileName: zone.ZoneFile}
	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	if err := SendCmd(data, CREATEZONE); err != nil {
		return last, err
	}
	tx.Commit()
	return last, nil
}

func (controller *DBController) DeleteZone(id string, viewID string) error {
	var err error
	if _, err := controller.GetZone(viewID, id); err != nil {
		return fmt.Errorf("unknown Zone with ID %s", id)
	}
	//delete the zone and rrs from the database
	var zoneDB tb.DBZone
	var num int
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	zoneDB.ID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Delete(&zoneDB).Error; err != nil {
		return err
	}
	if err := tx.Where("zone_id = ?", id).Delete(&tb.DBRR{}).Error; err != nil {
		return err
	}
	req := pb.DeleteZoneReq{ViewID: viewID, ZoneID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := SendCmd(data, DELETEZONE); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetZone(viewID string, id string) (*Zone, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var a tb.DBZone
	var index int
	var err error
	if index, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	if err := tx.Where("view_id = ?", viewID).First(&a, index).Error; err != nil {
		return nil, err
	}
	zone := Zone{}
	zone.SetID(id)
	zone.Name = a.Name
	var rrs []tb.DBRR
	if err := tx.Where("zone_id = ?", id).Find(&rrs).Error; err != nil {
		return nil, err
	}
	for _, dbRR := range rrs {
		rr := &RR{}
		rr.Data = dbRR.Data
		zone.rRs = append(zone.rRs, rr)
	}
	zone.Type = "zone"
	zone.SetCreationTimestamp(a.CreatedAt)
	return &zone, nil
}

/*func (controller *DBController) UpdateZone(zone *Zone) error {
	var one tb.DBZone
	var num int
	var err error
	if num, err = strconv.Atoi(zone.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = zone.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Where("acl_id = ?", zone.GetID()).Delete(&tb.DBIP{}).Error; err != nil {
		return err
	}
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	var iPDB tb.DBIP
	for _, ip := range zone.IPs {
		iPDB.IP = ip
		iPDB.ZoneID = zone.GetID()
		if err := tx.Create(&iPDB).Error; err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}*/

func (controller *DBController) GetZones(viewID string) []*Zone {
	var zones []*Zone
	var zoneDBs []tb.DBZone
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Where("view_id = ?", viewID).Find(&zoneDBs).Error; err != nil {
		return nil
	}
	for _, zoneDB := range zoneDBs {
		var zone Zone
		zone.SetID(strconv.Itoa(int(zoneDB.ID)))
		zone.Name = zoneDB.Name
		zone.ZoneFile = zoneDB.ZoneFile
		zones = append(zones, &zone)
	}
	return zones
}

func (controller *DBController) CreateRR(rr *RR, zoneID string, viewID string) (tb.DBRR, error) {
	//create new data in the database
	var one tb.DBRR
	one.ZoneID = zoneID
	one.Data = rr.Data
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Create(&one).Error; err != nil {
		return tb.DBRR{}, err
	}
	var last tb.DBRR
	tx.Last(&last)
	req := pb.CreateRRReq{ViewID: viewID, ZoneID: zoneID, RRID: strconv.Itoa(int(last.ID)), RRData: last.Data}
	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	if err := SendCmd(data, CREATERR); err != nil {
		return last, err
	}
	tx.Commit()
	return last, nil
}

func (controller *DBController) DeleteRR(id string, zoneID string, viewID string) error {
	var err error
	//delete the rr and rrs from the database
	var rrDB tb.DBRR
	var num int
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	rrDB.ID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	//check the relationship between rr,zone and view
	var tmpID int
	var v tb.DBView
	if tmpID, err = strconv.Atoi(viewID); err != nil {
		return err
	}
	if err := tx.First(&v, tmpID).Error; err != nil {
		return err
	}
	if tmpID, err = strconv.Atoi(zoneID); err != nil {
		return err
	}
	var z tb.DBZone
	if err := tx.Where("view_id = ?", viewID).First(&z, tmpID).Error; err != nil {
		return err
	}
	var rr tb.DBRR
	if err := tx.Where("zone_id = ?", viewID).First(&rr, num).Error; err != nil {
		return err
	}
	if err := tx.Delete(&rrDB).Error; err != nil {
		return err
	}
	req := pb.DeleteRRReq{ViewID: viewID, ZoneID: zoneID, RRID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := SendCmd(data, DELETEZONE); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetRR(id string, zoneID string, viewID string) (*RR, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var index int
	var err error
	if index, err = strconv.Atoi(viewID); err != nil {
		return nil, err
	}
	var v tb.DBView
	if err := tx.First(&v, index).Error; err != nil {
		return nil, err
	}
	var z tb.DBZone
	if index, err = strconv.Atoi(zoneID); err != nil {
		return nil, err
	}
	if err := tx.Where("view_id = ?", viewID).First(&z, index).Error; err != nil {
		return nil, err
	}
	var dbRR tb.DBRR
	if index, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	if err := tx.Where("zone_id = ?", viewID).First(&dbRR, index).Error; err != nil {
		return nil, err
	}
	rr := RR{}
	rr.SetID(id)
	rr.Data = dbRR.Data
	rr.Type = "rr"
	rr.SetCreationTimestamp(dbRR.CreatedAt)
	return &rr, nil
}

/*func (controller *DBController) UpdateRR(rr *RR) error {
	var one tb.DBRR
	var num int
	var err error
	if num, err = strconv.Atoi(rr.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = rr.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Where("acl_id = ?", rr.GetID()).Delete(&tb.DBIP{}).Error; err != nil {
		return err
	}
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	var iPDB tb.DBIP
	for _, ip := range rr.IPs {
		iPDB.IP = ip
		iPDB.RRID = rr.GetID()
		if err := tx.Create(&iPDB).Error; err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}*/

func (controller *DBController) GetRRs(zoneID string, viewID string) ([]*RR, error) {
	tx := controller.db.Begin()
	var index int
	var err error
	//check the relationship
	if index, err = strconv.Atoi(viewID); err != nil {
		return nil, err
	}
	var v tb.DBView
	if err := tx.First(&v, index).Error; err != nil {
		return nil, err
	}
	var z tb.DBZone
	if index, err = strconv.Atoi(zoneID); err != nil {
		return nil, err
	}
	if err := tx.Where("view_id = ?", viewID).First(&z, index).Error; err != nil {
		return nil, err
	}
	//get all RR
	var dbRRs []tb.DBRR
	if err := tx.Where("zone_id = ?", zoneID).Find(&dbRRs).Error; err != nil {
		return nil, err
	}
	var rrs []*RR
	for _, dbRR := range dbRRs {
		one := &RR{}
		one.SetID(strconv.Itoa(int(dbRR.ID)))
		one.Data = dbRR.Data
		one.SetCreationTimestamp(dbRR.CreatedAt)
		rrs = append(rrs, one)
	}
	defer tx.Rollback()
	return rrs, nil
}
