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
	UPDATEACL  = "UpdateACL"
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
	one.db.AutoMigrate(&tb.DBView{})
	one.db.AutoMigrate(&tb.DBZone{})
	one.db.AutoMigrate(&tb.DBRR{})
	one.db.AutoMigrate(&tb.DBACL{})
	one.db.AutoMigrate(&tb.DBIP{})
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
	for _, iP := range aCL.IPs {
		ip := tb.DBIP{IP: iP}
		one.IPs = append(one.IPs, ip)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.DBACL{}, err
	}
	var last tb.DBACL
	tx.Last(&last)
	var iPs []string
	for _, iP := range aCL.IPs {
		iPs = append(iPs, iP)
	}
	req := pb.CreateACLReq{Name: aCL.Name, ID: strconv.Itoa(int(last.ID)), IPs: iPs}
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
	one := tb.DBACL{}
	var index int
	if index, err = strconv.Atoi(id); err != nil {
		return err
	}
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.First(&one, index).Error; err != nil {
		return fmt.Errorf("unknown ACL with ID %s, %w", id, err)
	}
	//delete the ips and acl from the database
	var aCLDB tb.DBACL
	var num int
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	aCLDB.ID = uint(num)
	if err := tx.Unscoped().Delete(&aCLDB).Error; err != nil {
		return err
	}
	req := pb.DeleteACLReq{ID: id}
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
	//check the name of the acl is not exists except itself.
	tmp := tb.DBACL{}
	if err := tx.First(&tmp, one.ID).Error; err != nil {
		return err
	}
	if tmp.Name != aCL.Name {
		acls := []tb.DBACL{}
		if err := tx.Where("name = ?", aCL.Name).Find(&acls).Error; err != nil {
			return err
		}
		if len(acls) >= 1 {
			return fmt.Errorf("the name of the acl: %s is exists!", aCL.Name)
		}
	}
	//delete the old ips data.
	if err := tx.Where("acl_id = ?", aCL.GetID()).Delete(&tb.DBIP{}).Error; err != nil {
		return err
	}
	//add new ips to the acl
	for _, iP := range aCL.IPs {
		ip := tb.DBIP{IP: iP}
		one.IPs = append(one.IPs, ip)
	}
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	var iPs []string
	for _, iP := range aCL.IPs {
		iPs = append(iPs, iP)
	}
	req := pb.UpdateACLReq{ID: aCL.ID, Name: aCL.Name, NewIPs: iPs}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := SendCmd(data, UPDATEACL); err != nil {
		return err
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
	//add the acls to the view
	var err error
	var aclids []string
	for _, id := range view.ACLIDs {
		aclids = append(aclids, id)
		tmp := tb.DBACL{}
		var index int
		if index, err = strconv.Atoi(id); err != nil {
			return tb.DBView{}, err
		}
		tmp.ID = uint(index)
		//check wether the acl is valid
		var tmpACL tb.DBACL
		if err := tx.First(&tmpACL, id).Error; err != nil {
			return tb.DBView{}, fmt.Errorf("id %s of acl not exists, %w", id, err)
		}
		tmp.Name = tmpACL.Name
		one.ACLs = append(one.ACLs, tmp)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.DBView{}, err
	}
	var last tb.DBView
	tx.Last(&last)
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
	view := &View{}
	if view, err = controller.GetView(id); err != nil {
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
	//delete the relationship between view and acl
	if err := tx.Model(&viewDB).Association("ACLs").Clear().Error; err != nil {
		return err
	}
	priority := view.Priority
	//delete the view from the database
	if err := tx.Unscoped().Delete(&viewDB).Error; err != nil {
		return err
	}
	var allView []tb.DBView
	if err := tx.Find(&allView).Error; err != nil {
		return err
	}
	for i, viewDB := range allView {
		if viewDB.Priority > priority {
			allView[i].Priority--
		}
	}
	for _, viewDB := range allView {
		if err := tx.Model(&viewDB).UpdateColumn("priority", viewDB.Priority).Error; err != nil {
			return err
		}
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

func (controller *DBController) UpdateView(view *View) error {
	var one tb.DBView
	var num int
	var err error
	if num, err = strconv.Atoi(view.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = view.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	//check wether the view is exists
	var dbView tb.DBView
	if err := tx.First(&dbView, view.GetID()).Error; err != nil {
		return fmt.Errorf("the id %s of the view is not exists!%w", view.GetID(), err)
	}
	//adjust view priority
	var allDBView []tb.DBView
	if err := tx.Find(&allDBView).Error; err != nil {
		return err
	}
	origin := dbView.Priority
	dest := view.Priority
	if origin < dest {
		for k, v := range allDBView {
			if v.Priority > origin && v.Priority <= dest {
				allDBView[k].Priority--
			} else if v.Priority == origin {
				allDBView[k].Priority = dest
			} else {
				allDBView = append(allDBView[:k], allDBView[k+1:]...)
			}
		}
	} else if origin > dest {
		for k, v := range allDBView {
			if v.Priority >= dest && v.Priority < origin {
				allDBView[k].Priority++
			} else if v.Priority == origin {
				allDBView[k].Priority = dest
			} else {
				allDBView = append(allDBView[:k], allDBView[k+1:]...)
			}
		}
	}
	//update the priority in the database.
	for _, viewDB := range allDBView {
		if err := tx.Model(&viewDB).UpdateColumn("priority", viewDB.Priority).Error; err != nil {
			return err
		}
	}
	//delete the relationship between view and acl
	var delete_aclids []string
	var add_aclids []string
	add_aclids = view.ACLIDs
	var delete_acls []tb.DBACL
	if err := tx.Model(&one).Related(&delete_acls, "ACLs").Error; err != nil {
		return err
	}
	if err := tx.Model(&one).Association("ACLs").Clear().Error; err != nil {
		return err
	}
	// add the relationship between view and acl
	var dbACLs []tb.DBACL
	for _, id := range view.ACLIDs {
		var tmp tb.DBACL
		var index int
		if index, err = strconv.Atoi(id); err != nil {
			return err
		}
		tmp.ID = uint(index)
		if err := tx.First(&tmp, index).Error; err != nil {
			return fmt.Errorf("the acl of id:%d is not exists!%w", index, err)
		}
		dbACLs = append(dbACLs, tmp)
	}
	if err := tx.Model(&one).Association("ACLs").Append(dbACLs).Error; err != nil {
		return err
	}
	//collect the data for the command.
	for _, acl := range delete_acls {
		delete_aclids = append(delete_aclids, strconv.Itoa(int(acl.ID)))
	}
	for i, del_id := range delete_aclids {
		for k, add_id := range add_aclids {
			if del_id == add_id {
				delete_aclids = append(delete_aclids[:i], delete_aclids[i+1:]...)
				add_aclids = append(add_aclids[:k], add_aclids[k+1:]...)
			}
		}
	}
	req := pb.UpdateViewReq{ViewID: view.ID, Priority: int32(view.Priority), DeleteACLIDs: delete_aclids, AddACLIDs: add_aclids}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := SendCmd(data, UPDATEVIEW); err != nil {
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
	if err := tx.Model(&a).Related(&acls, "ACLs").Error; err != nil {
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
		//zone.ZoneFile = dbZone.ZoneFile
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
	var err error
	for _, viewDB := range viewDBs {
		tmp := &View{}
		if tmp, err = controller.GetView(strconv.Itoa(int(viewDB.ID))); err != nil {
			return nil
		}
		views = append(views, tmp)
	}
	return views
}

//////////////

func (controller *DBController) CreateZone(zone *Zone, viewID string) (tb.DBZone, error) {
	//create new data in the database
	var one tb.DBZone
	one.Name = zone.Name
	var num int
	var err error
	if num, err = strconv.Atoi(viewID); err != nil {
		return tb.DBZone{}, err
	}
	one.ViewID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	//set zone file name
	var dbview tb.DBView
	if err := tx.First(&dbview, num).Error; err != nil {
		return tb.DBZone{}, err
	}
	one.ZoneFile = zone.Name + dbview.Name + ".zone"
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
	req := pb.CreateZoneReq{ViewID: viewID, ZoneName: zone.Name, ZoneID: strconv.Itoa(int(last.ID)), ZoneFileName: one.ZoneFile}
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
		//zone.ZoneFile = zoneDB.ZoneFile
		zones = append(zones, &zone)
	}
	return zones
}

func (controller *DBController) CreateRR(rr *RR, zoneID string, viewID string) (tb.DBRR, error) {
	//create new data in the database
	var one tb.DBRR
	var num int
	var err error
	if num, err = strconv.Atoi(zoneID); err != nil {
		return tb.DBRR{}, err
	}
	one.ZoneID = uint(num)
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
	if err := tx.Where("zone_id = ?", zoneID).First(&rr, num).Error; err != nil {
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

func (controller *DBController) UpdateRR(rr *RR, zoneID string, viewID string) error {
	var one tb.DBRR
	var num int
	var err error
	if num, err = strconv.Atoi(rr.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Data = rr.Data
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Model(&one).UpdateColumn("data", one.Data).Error; err != nil {
		return err
	}
	req := pb.UpdateRRReq{ViewID: viewID, ZoneID: zoneID, RRID: rr.ID, NewRRData: rr.Data}
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
