package restfulapi

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm"
	tb "github.com/linkingthing/ddi/dns/cockroachtables"
	"github.com/linkingthing/ddi/pb"
	myaes "github.com/linkingthing/ddi/utils/aes"
	kfkcli "github.com/linkingthing/ddi/utils/kafkaclient"
	"strconv"
)

const (
	STARTDNS                  = "StartDNS"
	STOPDNS                   = "StopDNS"
	CREATEACL                 = "CreateACL"
	DELETEACL                 = "DeleteACL"
	UPDATEACL                 = "UpdateACL"
	CREATEVIEW                = "CreateView"
	UPDATEVIEW                = "UpdateView"
	DELETEVIEW                = "DeleteView"
	CREATEZONE                = "CreateZone"
	DELETEZONE                = "DeleteZone"
	CREATERR                  = "CreateRR"
	UPDATERR                  = "UpdateRR"
	DELETERR                  = "DeleteRR"
	UPDATEDEFAULTFORWARD      = "UpdateDefaultForward"
	DELETEDEFAULTFORWARD      = "DeleteDefaultForward"
	UPDATEFORWARD             = "UpdateForward"
	DELETEFORWARD             = "DeleteForward"
	CREATEREDIRECTION         = "CreateRedirection"
	UPDATEREDIRECTION         = "UpdateRedirection"
	DELETEREDIRECTION         = "DeleteRedirection"
	CREATEDEFAULTDNS64        = "CreateDefaultDNS64"
	UPDATEDEFAULTDNS64        = "UpdateDefaultDNS64"
	DELETEDEFAULTDNS64        = "DeleteDefaultDNS64"
	CREATEDNS64               = "CreateDNS64"
	UPDATEDNS64               = "UpdateDNS64"
	DELETEDNS64               = "DeleteDNS64"
	CREATEIPBLACKHOLE         = "CreateIPBlackHole"
	UPDATEIPBLACKHOLE         = "UpdateIPBlackHole"
	DELETEIPBLACKHOLE         = "DeleteIPBlackHole"
	UPDATERECURSIVECONCURRENT = "UpdateRecursiveConcurrent"
	CREATESORTLIST            = "CreateSortList"
	UPDATESORTLIST            = "UpdateSortList"
	DELETESORTLIST            = "DeleteSortList"
)

var DBCon *DBController

type DBController struct {
	db     *gorm.DB
	aesKey []byte
}

func NewDBController(db *gorm.DB) *DBController {
	one := &DBController{}
	one.db = db
	tx := one.db.Begin()
	defer tx.Rollback()
	if err := tx.AutoMigrate(&tb.View{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.Zone{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.RR{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.ACL{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.IP{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.Forwarder{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.DefaultForward{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.DefaultForwarder{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.Redirection{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.DefaultDNS64{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.DNS64{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.IPBlackHole{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.RecursiveConcurrent{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.DDIUserPWD{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.SortListElement{}).Error; err != nil {
		panic(err)
	}
	one.aesKey = []byte("linkingthing.com")
	//if the user "admin"is not exists create one
	var userPWD tb.DDIUserPWD
	if err := tx.Where("name = ?", "admin").Find(&userPWD).Error; err != nil {
		newOne := tb.DDIUserPWD{}
		newOne.Name = "admin"
		encryptText, err := myaes.Encrypt(one.aesKey, "admin")
		if err != nil {
			panic(err)
		}
		fmt.Println(encryptText)
		newOne.PWD = encryptText
		if err := tx.Create(&newOne).Error; err != nil {
			panic(err)
		}
	}
	any := tb.ACL{}
	any.ID = 1
	if err := tx.Find(&any).Error; err != nil {
		any.Name = "any"
		if err := tx.Create(&any).Error; err != nil {
			panic(err)
		}
	}
	none := tb.ACL{}
	none.ID = 2
	if err := tx.Find(&none).Error; err != nil {
		none.Name = "none"
		if err := tx.Create(&none).Error; err != nil {
			panic(err)
		}
	}
	viewDefault := tb.View{}
	viewDefault.ID = 1000000
	var many []tb.View
	if err := tx.Find(&many).Error; err != nil {
		panic(err)
	}
	if err := tx.Find(&viewDefault).Error; err != nil {
		viewDefault.Name = "default"
		viewDefault.Priority = len(many) + 1
		viewDefault.ACLs = append(viewDefault.ACLs, any)
		if err := tx.Create(&viewDefault).Error; err != nil {
			panic(err)
		}
	}
	tx.Commit()
	return one
}
func (controller *DBController) Close() {
	controller.db.Close()
}

func (controller *DBController) CreateACL(aCL *ACL) (tb.ACL, error) {
	//create new data in the database
	var one tb.ACL
	one.Name = aCL.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	var dbACLs []tb.ACL
	if err := tx.Where("name = ?", aCL.Name).Find(&dbACLs).Error; err != nil {
		return tb.ACL{}, err
	}
	if len(dbACLs) > 0 {
		return tb.ACL{}, fmt.Errorf("the name %s of acl exists", aCL.Name)
	}
	for _, iP := range aCL.IPs {
		ip := tb.IP{IP: iP}
		one.IPs = append(one.IPs, ip)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.ACL{}, err
	}
	var last tb.ACL
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
	if err := kfkcli.KafkaClient.SendCmd(data, CREATEACL); err != nil {
		return last, err
	}
	tx.Commit()
	return last, nil
}

func (controller *DBController) DeleteACL(id string) error {
	var err error
	if id == "1" || id == "2" {
		return fmt.Errorf("It's not allow to delete the default any or none acl!")
	}
	one := tb.ACL{}
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.First(&one, id).Error; err != nil {
		return fmt.Errorf("unknown ACL with ID %s, %w", id, err)
	}
	var views []tb.View
	if err := tx.Model(&one).Related(&views, "Views").Error; err != nil {
		return err
	}
	if len(views) > 0 {
		return fmt.Errorf("the ACL with ID %s is used by view or views, can not be deleted!", id)
	}
	var sortLists []tb.SortListElement
	if err := tx.Model(&one).Related(&sortLists, "SortListElements").Error; err != nil {
		return err
	}
	if len(sortLists) > 0 {
		return fmt.Errorf("the ACL with ID %s is used by sortList, can not be deleted!", id)
	}
	//delete the ips and acl from the database
	var aCLDB tb.ACL
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
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEACL); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetACL(id string) (*ACL, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var a tb.ACL
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
	var iPs []tb.IP
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
	var one tb.ACL
	var num int
	var err error
	if aCL.ID == "1" || aCL.ID == "2" {
		fmt.Errorf("It's not allow to modify the default any or none acl!")
	}
	if num, err = strconv.Atoi(aCL.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = aCL.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	//check the name of the acl is not exists except itself.
	tmp := tb.ACL{}
	if err := tx.First(&tmp, one.ID).Error; err != nil {
		return err
	}
	if tmp.Name != aCL.Name {
		acls := []tb.ACL{}
		if err := tx.Where("name = ?", aCL.Name).Find(&acls).Error; err != nil {
			return err
		}
		if len(acls) >= 1 {
			return fmt.Errorf("the name of the acl: %s is exists!", aCL.Name)
		}
	}
	//delete the old ips data.
	if err := tx.Where("acl_id = ?", aCL.GetID()).Delete(&tb.IP{}).Error; err != nil {
		return err
	}
	//add new ips to the acl
	for _, iP := range aCL.IPs {
		ip := tb.IP{IP: iP}
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
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEACL); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (controller *DBController) GetACLs() []*ACL {
	var aCLs []*ACL
	var aCLDBs []tb.ACL
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
		var iPDBs []tb.IP
		if err := tx.Where("acl_id = ?", aCLDB.ID).Find(&iPDBs).Error; err == nil {
			for _, iP := range iPDBs {
				aCL.IPs = append(aCL.IPs, iP.IP)
			}
		}
		aCLs = append(aCLs, &aCL)
	}
	return aCLs
}

func (controller *DBController) CreateView(view *View) (tb.View, error) {
	//create new data in the database
	var one tb.View
	one.Name = view.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	//adjust the priority
	var allView []tb.View
	if err := tx.Find(&allView).Error; err != nil {
		return tb.View{}, err
	}
	one.Priority = view.Priority
	if view.Priority > len(allView) {
		one.Priority = len(allView)
	} else if view.Priority < 0 {
		one.Priority = 1
	}
	var tmpView []tb.View
	for i, viewDB := range allView {
		if viewDB.Priority >= one.Priority {
			allView[i].Priority++
		}
		tmpView = append(tmpView, allView[i])
	}
	for _, viewDB := range tmpView {
		if err := tx.Model(&viewDB).UpdateColumn("priority", viewDB.Priority).Error; err != nil {
			return tb.View{}, err
		}
	}
	var dbViews []tb.View
	//check wether the view is exists.
	if err := tx.Where("name = ?", view.Name).Find(&dbViews).Error; err != nil {
		return tb.View{}, err
	}
	if len(dbViews) > 0 {
		return tb.View{}, fmt.Errorf("the name %s of view has exists!", view.Name)
	}
	//add the acls to the view
	var err error
	var aclids []string
	for _, id := range view.ACLIDs {
		aclids = append(aclids, id)
		tmp := tb.ACL{}
		var index int
		if index, err = strconv.Atoi(id); err != nil {
			return tb.View{}, err
		}
		tmp.ID = uint(index)
		//check wether the acl is valid
		var tmpACL tb.ACL
		if err := tx.First(&tmpACL, id).Error; err != nil {
			return tb.View{}, fmt.Errorf("id %s of acl not exists, %w", id, err)
		}
		tmp.Name = tmpACL.Name
		one.ACLs = append(one.ACLs, tmp)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.View{}, err
	}
	var last tb.View
	tx.Last(&last)
	req := pb.CreateViewReq{ViewName: view.Name, ViewID: strconv.Itoa(int(last.ID)), Priority: int32(view.Priority), ACLIDs: aclids}
	data, err := proto.Marshal(&req)
	if err != nil {
		return tb.View{}, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATEVIEW); err != nil {
		return tb.View{}, err
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
	if view.Name == "default" {
		return fmt.Errorf("view default can't not be deleted")
	}
	var viewDB tb.View
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
	var allView []tb.View
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
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEVIEW); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) UpdateView(view *View) error {
	var one tb.View
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
	var dbView tb.View
	if err := tx.First(&dbView, view.GetID()).Error; err != nil {
		return fmt.Errorf("the id %s of the view is not exists!%w", view.GetID(), err)
	}
	//adjust view priority
	var allDBView []tb.View
	var newDBViews []tb.View
	if err := tx.Find(&allDBView).Error; err != nil {
		return err
	}
	origin := dbView.Priority
	dest := view.Priority
	if dest >= len(allDBView) {
		dest = len(allDBView) - 1
	}
	if origin < dest {
		for k, v := range allDBView {
			if v.Priority > origin && v.Priority <= dest {
				allDBView[k].Priority--
				newDBViews = append(newDBViews, allDBView[k])
			} else if v.Priority == origin {
				allDBView[k].Priority = dest
				newDBViews = append(newDBViews, allDBView[k])
			}
		}
	} else if origin > dest {
		for k, v := range allDBView {
			if v.Priority >= dest && v.Priority < origin {
				allDBView[k].Priority++
				newDBViews = append(newDBViews, allDBView[k])
			} else if v.Priority == origin {
				allDBView[k].Priority = dest
				newDBViews = append(newDBViews, allDBView[k])
			}
		}
	}
	//update the priority in the database.
	for _, viewDB := range newDBViews {
		if err := tx.Model(&viewDB).UpdateColumn("priority", viewDB.Priority).Error; err != nil {
			return err
		}
	}
	//delete the relationship between view and acl
	var delete_aclids []string
	var add_aclids []string
	add_aclids = view.ACLIDs
	var delete_acls []tb.ACL
	if err := tx.Model(&one).Related(&delete_acls, "ACLs").Error; err != nil {
		return err
	}
	if err := tx.Model(&one).Association("ACLs").Clear().Error; err != nil {
		return err
	}
	// add the relationship between view and acl
	var dbACLs []tb.ACL
	for _, id := range view.ACLIDs {
		var tmp tb.ACL
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
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEVIEW); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetView(id string) (*View, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var a tb.View
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
	var acls []tb.ACL
	if err := tx.Model(&a).Related(&acls, "ACLs").Error; err != nil {
		return nil, err
	}
	var ids []string
	for _, acl := range acls {
		ids = append(ids, strconv.Itoa(int(acl.ID)))
		var tmp ACL
		tmp.ID = strconv.Itoa(int(acl.ID))
		tmp.Name = acl.Name
		view.ACLs = append(view.ACLs, &tmp)
	}
	view.ACLIDs = ids
	var zones []tb.Zone
	if err := tx.Where("view_id = ?", id).Find(&zones).Error; err != nil {
		return nil, err
	}
	for _, dbZone := range zones {
		var zone Zone
		zone.SetID(strconv.Itoa(int(dbZone.ID)))
		zone.Name = dbZone.Name
		//zone.ZoneFile = dbZone.ZoneFile
		view.Zones = append(view.Zones, &zone)
	}
	view.ZoneSize = len(zones)
	var redirections []tb.Redirection
	if err := tx.Where("view_id = ?", id).Where("redirect_type = ?", "rpz").Find(&redirections).Error; err != nil {
		return nil, err
	}
	view.RPZSize = len(redirections)
	if err := tx.Where("view_id = ?", id).Where("redirect_type = ?", "redirect").Find(&redirections).Error; err != nil {
		return nil, err
	}
	view.RedirectSize = len(redirections)
	var dns64s []tb.DNS64
	if err := tx.Where("view_id = ?", id).Find(&dns64s).Error; err != nil {
		return nil, err
	}
	view.DNS64Size = len(dns64s)
	return &view, nil
}

func (controller *DBController) GetViews() []*View {
	var views []*View
	var viewDBs []tb.View
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Order("priority").Find(&viewDBs).Error; err != nil {
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

func (controller *DBController) CreateZone(zone *Zone, viewID string) (tb.Zone, error) {
	//create new data in the database
	var one tb.Zone
	one.Name = zone.Name
	var num int
	var err error
	if num, err = strconv.Atoi(viewID); err != nil {
		return tb.Zone{}, err
	}
	one.ViewID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	//set zone file name
	var dbview tb.View
	if err := tx.First(&dbview, num).Error; err != nil {
		return tb.Zone{}, err
	}
	if zone.ZoneType == "master" {
		one.ZoneFile = zone.Name + dbview.Name + ".zone"
	}
	one.ZoneType = zone.ZoneType
	var dbZones []tb.Zone
	if err := tx.Where("name = ?", zone.Name).Where("view_id = ?", viewID).Where("zone_type = ?", zone.ZoneType).Find(&dbZones).Error; err != nil {
		return tb.Zone{}, err
	}
	if len(dbZones) > 0 {
		return tb.Zone{}, fmt.Errorf("the name %s of zone has exists!", zone.Name)
	}
	if err := tx.Create(&one).Error; err != nil {
		return tb.Zone{}, err
	}
	var last tb.Zone
	tx.Last(&last)
	req := pb.CreateZoneReq{ViewID: viewID, ZoneName: zone.Name, ZoneID: strconv.Itoa(int(last.ID)), ZoneFileName: one.ZoneFile}
	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATEZONE); err != nil {
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
	var zoneDB tb.Zone
	var num int
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	zoneDB.ID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Unscoped().Delete(&zoneDB).Error; err != nil {
		return err
	}
	req := pb.DeleteZoneReq{ViewID: viewID, ZoneID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEZONE); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetZone(viewID string, id string) (*Zone, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var a tb.Zone
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
	var rrs []tb.RR
	if err := tx.Where("zone_id = ?", id).Find(&rrs).Error; err != nil {
		return nil, err
	}
	for _, dbRR := range rrs {
		rr := &RR{}
		rr.Name = dbRR.Name
		rr.DataType = dbRR.DataType
		rr.TTL = dbRR.TTL
		rr.Value = dbRR.Value
		zone.rRs = append(zone.rRs, rr)
	}
	zone.Type = "zone"
	zone.SetCreationTimestamp(a.CreatedAt)
	zone.RRSize = len(rrs)
	var forwarders []tb.Forwarder
	if err := tx.Where("zone_id = ?", zone.ID).Find(&forwarders).Error; err != nil {
		return nil, err
	}
	zone.ForwarderSize = len(forwarders)
	return &zone, nil
}

/*func (controller *DBController) UpdateZone(zone *Zone) error {
	var one tb.Zone
	var num int
	var err error
	if num, err = strconv.Atoi(zone.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = zone.Name
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Where("acl_id = ?", zone.GetID()).Delete(&tb.IP{}).Error; err != nil {
		return err
	}
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	var iPDB tb.IP
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

func (controller *DBController) GetZones(viewID string, zoneType string) []*Zone {
	var zones []*Zone
	var zoneDBs []tb.Zone
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Where("view_id = ?", viewID).Where("zone_type = ?", zoneType).Find(&zoneDBs).Error; err != nil {
		return nil
	}
	var err error
	for _, zoneDB := range zoneDBs {
		var zone *Zone
		if zone, err = controller.GetZone(viewID, strconv.Itoa(int(zoneDB.ID))); err != nil {
			return nil
		}
		zones = append(zones, zone)
	}
	return zones
}

func (controller *DBController) CreateRR(rr *RR, zoneID string, viewID string) (tb.RR, error) {
	//create new data in the database
	var one tb.RR
	var num int
	var err error
	if num, err = strconv.Atoi(zoneID); err != nil {
		return tb.RR{}, err
	}
	one.ZoneID = uint(num)
	one.Name = rr.Name
	one.DataType = rr.DataType
	one.TTL = rr.TTL
	one.Value = rr.Value
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Create(&one).Error; err != nil {
		return tb.RR{}, err
	}
	var last tb.RR
	tx.Last(&last)
	req := pb.CreateRRReq{ViewID: viewID, ZoneID: zoneID, RRID: strconv.Itoa(int(last.ID)), Name: rr.Name, Type: rr.DataType, TTL: strconv.Itoa(int(rr.TTL)), Value: rr.Value}
	data, err := proto.Marshal(&req)
	if err != nil {
		return last, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATERR); err != nil {
		return last, err
	}
	tx.Commit()
	return last, nil
}

func (controller *DBController) DeleteRR(id string, zoneID string, viewID string) error {
	var err error
	//delete the rr and rrs from the database
	var rrDB tb.RR
	var num int
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	rrDB.ID = uint(num)
	tx := controller.db.Begin()
	defer tx.Rollback()
	//check the relationship between rr,zone and view
	var tmpID int
	var v tb.View
	if tmpID, err = strconv.Atoi(viewID); err != nil {
		return err
	}
	if err := tx.First(&v, tmpID).Error; err != nil {
		return err
	}
	if tmpID, err = strconv.Atoi(zoneID); err != nil {
		return err
	}
	var z tb.Zone
	if err := tx.Where("view_id = ?", viewID).First(&z, tmpID).Error; err != nil {
		return err
	}
	var rr tb.RR
	if err := tx.Where("zone_id = ?", zoneID).First(&rr, num).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Delete(&rrDB).Error; err != nil {
		return err
	}
	req := pb.DeleteRRReq{ViewID: viewID, ZoneID: zoneID, RRID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETERR); err != nil {
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
	var v tb.View
	if err := tx.First(&v, index).Error; err != nil {
		return nil, err
	}
	var z tb.Zone
	if index, err = strconv.Atoi(zoneID); err != nil {
		return nil, err
	}
	if err := tx.Where("view_id = ?", viewID).First(&z, index).Error; err != nil {
		return nil, err
	}
	var dbRR tb.RR
	if index, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	if err := tx.Where("zone_id = ?", zoneID).First(&dbRR, index).Error; err != nil {
		return nil, err
	}
	rr := RR{}
	rr.SetID(id)
	rr.Name = dbRR.Name
	rr.DataType = dbRR.DataType
	rr.TTL = dbRR.TTL
	rr.Value = dbRR.Value
	rr.Type = "rr"
	rr.SetCreationTimestamp(dbRR.CreatedAt)
	return &rr, nil
}

func (controller *DBController) UpdateRR(rr *RR, zoneID string, viewID string) error {
	var one tb.RR
	var num int
	var err error
	if num, err = strconv.Atoi(rr.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = rr.Name
	one.DataType = rr.DataType
	one.TTL = rr.TTL
	one.Value = rr.Value
	var id int
	if id, err = strconv.Atoi(zoneID); err != nil {
		return err
	}
	one.ZoneID = uint(id)
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	req := pb.UpdateRRReq{ViewID: viewID, ZoneID: zoneID, RRID: rr.ID, Name: rr.Name, Type: rr.DataType, TTL: strconv.Itoa(int(rr.TTL)), Value: rr.Value}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATERR); err != nil {
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
	var v tb.View
	if err := tx.First(&v, index).Error; err != nil {
		return nil, err
	}
	var z tb.Zone
	if index, err = strconv.Atoi(zoneID); err != nil {
		return nil, err
	}
	if err := tx.Where("view_id = ?", viewID).First(&z, index).Error; err != nil {
		return nil, err
	}
	//get all RR
	var dbRRs []tb.RR
	if err := tx.Where("zone_id = ?", zoneID).Find(&dbRRs).Error; err != nil {
		return nil, err
	}
	var rrs []*RR
	for _, dbRR := range dbRRs {
		one := &RR{}
		one.SetID(strconv.Itoa(int(dbRR.ID)))
		one.Name = dbRR.Name
		one.DataType = dbRR.DataType
		one.TTL = dbRR.TTL
		one.Value = dbRR.Value
		one.SetCreationTimestamp(dbRR.CreatedAt)
		rrs = append(rrs, one)
	}
	defer tx.Rollback()
	return rrs, nil
}

func (controller *DBController) CreateDefaultForward(fw *Forward) (tb.DefaultForward, error) {
	//check wether the default forward had been exists.
	many := []tb.DefaultForward{}
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Find(&many).Error; err != nil {
		return tb.DefaultForward{}, err
	}
	if len(many) >= 1 {
		return tb.DefaultForward{}, fmt.Errorf("Just allow one default forward configuration!")
	}
	defaultfw := tb.DefaultForward{ForwardType: fw.ForwardType}
	for _, ip := range fw.IPs {
		tmp := tb.DefaultForwarder{IP: ip}
		defaultfw.Forwarders = append(defaultfw.Forwarders, tmp)
	}
	if err := tx.Create(&defaultfw).Error; err != nil {
		return tb.DefaultForward{}, err
	}
	tx.Commit()
	return defaultfw, nil
}

func (controller *DBController) DeleteDefaultForward(id string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	defaultfw := tb.DefaultForward{}
	if err := tx.Unscoped().Delete(&defaultfw).Error; err != nil {
		return err
	}
	req := pb.DeleteDefaultForwardReq{}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEDEFAULTFORWARD); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) UpdateDefaultForward(forward *Forward) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var many []tb.DefaultForward
	if err := tx.Find(&many).Error; err != nil {
		return err
	}
	if len(many) == 0 {
		forward.ID = ""
		fw, err := controller.CreateDefaultForward(forward)
		if err != nil {
			return err
		}
		forward.ID = strconv.Itoa(int(fw.ID))
	} else {
		fw := tb.DefaultForward{}
		fw.ID = many[0].ID
		fw.ForwardType = forward.ForwardType
		for _, ip := range forward.IPs {
			tmp := tb.DefaultForwarder{IP: ip}
			fw.Forwarders = append(fw.Forwarders, tmp)
		}
		//delete the old ips
		if err := tx.Unscoped().Where("forward_id = ?", many[0].ID).Delete(&tb.DefaultForwarder{}).Error; err != nil {
			return err
		}
		//update new data
		if err := tx.Save(&fw).Error; err != nil {
			return err
		}
	}
	req := pb.UpdateDefaultForwardReq{Type: forward.ForwardType, IPs: forward.IPs}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEDEFAULTFORWARD); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetDefaultForward(id string) (*Forward, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	fw := tb.DefaultForward{}
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	fw.ID = uint(num)
	if err := tx.First(&fw).Error; err != nil {
		return nil, err
	}
	if err := tx.Model(&fw).Association("Forwarders").Find(&fw.Forwarders).Error; err != nil {
		return nil, err
	}
	one := Forward{}
	one.ID = id
	one.ForwardType = fw.ForwardType
	for _, v := range fw.Forwarders {
		one.IPs = append(one.IPs, v.IP)
	}
	one.Type = "forward"
	one.SetCreationTimestamp(fw.CreatedAt)
	return &one, nil
}

func (controller *DBController) GetDefaultForwards() ([]*Forward, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	fws := []*Forward{}
	tbFWs := []tb.DefaultForward{}
	if err := tx.Find(&tbFWs).Error; err != nil {
		return nil, err
	}
	var err error
	for _, one := range tbFWs {
		tmp := &Forward{}
		if tmp, err = controller.GetDefaultForward(strconv.Itoa(int(one.ID))); err != nil {
			return nil, err
		}
		fws = append(fws, tmp)
	}
	if len(fws) == 0 {
		tmp := &Forward{}
		tmp.ID = "0"
		fws = append(fws, tmp)
		return fws, nil
	}
	return fws, nil
}

func (controller *DBController) GetForward(id string) (*ForwardData, error) {
	var fws []tb.Forwarder
	tx := controller.db.Begin()
	defer tx.Rollback()
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	var zone tb.Zone
	zone.ID = uint(num)
	if err := tx.First(&zone).Error; err != nil {
		return nil, err
	}
	if zone.IsForward == 0 {
		tmp := ForwardData{}
		tmp.ID = "0"
		return &tmp, nil
	}
	if err := tx.Model(&zone).Association("Forwarders").Find(&fws).Error; err != nil {
		return nil, err
	}
	tmp := ForwardData{}
	tmp.ID = "0"
	tmp.ForwardType = zone.ForwardType
	for _, fw := range fws {
		tmp.ID = strconv.Itoa(int(fw.ID))
		tmp.IPs = append(tmp.IPs, fw.IP)
	}
	return &tmp, nil
}

func (controller *DBController) UpdateForward(forward *ForwardData, zoneID string, viewID string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var zone tb.Zone
	var num int
	var err error
	if num, err = strconv.Atoi(zoneID); err != nil {
		return err
	}
	zone.ID = uint(num)
	if err := tx.First(&zone).Error; err != nil {
		return err
	}
	//delete all the old data in the Forwarders;
	if err := tx.Unscoped().Where("zone_id = ?", zoneID).Delete(&tb.Forwarder{}).Error; err != nil {
		return err
	}
	//add new data to the zone and forwarders.
	zone.IsForward = 1
	zone.ForwardType = forward.ForwardType
	for _, ip := range forward.IPs {
		tmp := tb.Forwarder{}
		tmp.IP = ip
		zone.Forwarders = append(zone.Forwarders, tmp)
	}
	if err := tx.Save(&zone).Error; err != nil {
		return err
	}
	req := pb.UpdateForwardReq{ViewID: viewID, ZoneID: zoneID, Type: forward.ForwardType, IPs: forward.IPs}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEFORWARD); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) DeleteForward(id string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	//delete the old ips
	if err := tx.Unscoped().Where("zone_id = ?", id).Delete(&tb.Forwarder{}).Error; err != nil {
		return err
	}
	zone := tb.Zone{}
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	zone.ID = uint(num)
	if err := tx.Model(&zone).UpdateColumn("is_forward", 0).Error; err != nil {
		return err
	}
	req := pb.DeleteForwardReq{}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEFORWARD); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

////////////////
func (controller *DBController) CreateRedirection(rd *Redirection, viewID string) (*tb.Redirection, error) {
	var view tb.View
	tx := controller.db.Begin()
	defer tx.Rollback()
	var num int
	var err error
	if num, err = strconv.Atoi(viewID); err != nil {
		return nil, err
	}
	view.ID = uint(num)
	if err := tx.First(&view).Error; err != nil {
		return nil, fmt.Errorf("id %s of view not exists, %w", viewID, err)
	}
	tbrd := tb.Redirection{Name: rd.Name, TTL: rd.TTL, DataType: rd.DataType, RedirectType: rd.RedirectType, Value: rd.Value}
	view.Redirections = append(view.Redirections, tbrd)
	if err := tx.Save(&view).Error; err != nil {
		return nil, err
	}
	if err := tx.Last(&tbrd).Error; err != nil {
		return nil, err
	}
	req := pb.CreateRedirectionReq{ID: strconv.Itoa(int(tbrd.ID)), ViewID: viewID, Name: rd.Name, TTL: strconv.Itoa(int(rd.TTL)), DataType: rd.DataType, RedirectType: rd.RedirectType, Value: rd.Value}
	data, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATEREDIRECTION); err != nil {
		return nil, err
	}
	tx.Commit()
	return &tbrd, nil
}

func (controller *DBController) DeleteRedirection(id string, viewID string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	rd := tb.Redirection{}
	rd.ID = uint(num)
	if err := tx.First(&rd).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Delete(&rd).Error; err != nil {
		return err
	}
	req := pb.DeleteRedirectionReq{ID: id, ViewID: viewID, RedirectType: rd.RedirectType}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEREDIRECTION); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) UpdateRedirection(rd *Redirection, viewID string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var view tb.View
	var num int
	var err error
	if num, err = strconv.Atoi(viewID); err != nil {
		return err
	}
	view.ID = uint(num)
	if err := tx.First(&view).Error; err != nil {
		return fmt.Errorf("the id %s of view does not exists!")
	}
	var one tb.Redirection
	if num, err = strconv.Atoi(rd.ID); err != nil {
		return err
	}
	one.ID = uint(num)
	one.Name = rd.Name
	one.TTL = rd.TTL
	one.DataType = rd.DataType
	one.RedirectType = rd.RedirectType
	one.Value = rd.Value
	view.Redirections = append(view.Redirections, one)
	if err := tx.Save(&view).Error; err != nil {
		return err
	}
	req := pb.UpdateRedirectionReq{ID: rd.ID, ViewID: viewID, Name: rd.Name, TTL: strconv.Itoa(int(rd.TTL)), DataType: rd.DataType, RedirectType: rd.RedirectType, IP: rd.Value}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEREDIRECTION); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetRedirection(id string) (*Redirection, error) {
	var rd tb.Redirection
	tx := controller.db.Begin()
	defer tx.Rollback()
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return nil, err
	}
	rd.ID = uint(num)
	if err := tx.First(&rd).Error; err != nil {
		return nil, err
	}
	var tmp Redirection
	tmp.ID = id
	tmp.Name = rd.Name
	tmp.TTL = rd.TTL
	tmp.DataType = rd.DataType
	tmp.RedirectType = rd.RedirectType
	tmp.Value = rd.Value
	return &tmp, nil
}

func (controller *DBController) GetRedirections(viewID string) ([]*Redirection, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	rds := []*Redirection{}
	var tbrds []tb.Redirection
	if err := tx.Where("view_id = ?", viewID).Find(&tbrds).Error; err != nil {
		return nil, err
	}
	var err error
	for _, one := range tbrds {
		tmp := &Redirection{}
		if tmp, err = controller.GetRedirection(strconv.Itoa(int(one.ID))); err != nil {
			return nil, err
		}
		rds = append(rds, tmp)
	}
	return rds, nil
}

/////
func (controller *DBController) CreateDefaultDNS64(dns64 *DefaultDNS64) (*tb.DefaultDNS64, error) {
	//check whether the acl is exists.
	if err := controller.CheckACL(dns64.ClientACL); err != nil {
		return nil, err
	}
	if err := controller.CheckACL(dns64.AAddress); err != nil {
		return nil, err
	}
	//create the default dns64 data.There is not any relationship between default dns64 table and the view table.
	tx := controller.db.Begin()
	defer tx.Rollback()
	tbdns64 := tb.DefaultDNS64{}
	tbdns64.Prefix = dns64.Prefix
	var num int
	var err error
	if num, err = strconv.Atoi(dns64.ClientACL); err != nil {
		return nil, err
	}
	tbdns64.ClientACL = uint(num)
	if num, err = strconv.Atoi(dns64.AAddress); err != nil {
		return nil, err
	}
	tbdns64.AAddress = uint(num)
	if err := tx.Create(&tbdns64).Error; err != nil {
		return nil, err
	}
	req := pb.CreateDefaultDNS64Req{ID: strconv.Itoa(int(tbdns64.ID)), Prefix: dns64.Prefix, ClientACL: dns64.ClientACL, AAddress: dns64.AAddress}
	data, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATEDEFAULTDNS64); err != nil {
		return nil, err
	}
	tx.Commit()
	return &tbdns64, nil
}

func (controller *DBController) DeleteDefaultDNS64(id string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	dns64 := tb.DefaultDNS64{}
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	dns64.ID = uint(num)
	if err := tx.Unscoped().Delete(&dns64).Error; err != nil {
		return err
	}
	req := pb.DeleteDefaultDNS64Req{ID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEDEFAULTDNS64); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) UpdateDefaultDNS64(dns64 *DefaultDNS64) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	tbdns64 := tb.DefaultDNS64{}
	var num int
	var err error
	if num, err = strconv.Atoi(dns64.ID); err != nil {
		return err
	}
	tbdns64.ID = uint(num)
	if err := controller.CheckACL(dns64.ClientACL); err != nil {
		return err
	}
	if err := controller.CheckACL(dns64.AAddress); err != nil {
		return err
	}
	tbdns64.Prefix = dns64.Prefix
	if num, err = strconv.Atoi(dns64.ClientACL); err != nil {
		return err
	}
	tbdns64.ClientACL = uint(num)
	if num, err = strconv.Atoi(dns64.AAddress); err != nil {
		return err
	}
	tbdns64.AAddress = uint(num)
	if err := tx.Save(&tbdns64).Error; err != nil {
		return err
	}
	req := pb.UpdateDefaultDNS64Req{ID: strconv.Itoa(int(tbdns64.ID)), Prefix: dns64.Prefix, ClientACL: dns64.ClientACL, AAddress: dns64.AAddress}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEDEFAULTDNS64); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) CheckACL(id string) error {
	if err := controller.db.First(&tb.ACL{}, id).Error; err != nil {
		return fmt.Errorf("id %s of acl not exists, %w", id, err)
	}
	return nil
}

func (controller *DBController) GetDefaultDNS64(id string) (*DefaultDNS64, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	dns64 := tb.DefaultDNS64{}
	if err := tx.First(&dns64, id).Error; err != nil {
		return nil, err
	}
	one := DefaultDNS64{}
	one.ID = id
	one.Prefix = dns64.Prefix
	one.ClientACL = strconv.Itoa(int(dns64.ClientACL))
	var acl tb.ACL
	if err := tx.First(&acl, dns64.ClientACL).Error; err != nil {
		return nil, err
	}
	one.ClientACLName = acl.Name
	acl.ID = 0
	one.AAddress = strconv.Itoa(int(dns64.AAddress))
	acl.ID = 0
	if err := tx.First(&acl, dns64.AAddress).Error; err != nil {
		return nil, err
	}
	one.AddressName = acl.Name
	return &one, nil
}

func (controller *DBController) GetDefaultDNS64s() ([]*DefaultDNS64, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	dns64s := []*DefaultDNS64{}
	tbDNS64s := []tb.DefaultDNS64{}
	if err := tx.Find(&tbDNS64s).Error; err != nil {
		return nil, err
	}
	var err error
	for _, one := range tbDNS64s {
		tmp := &DefaultDNS64{}
		if tmp, err = controller.GetDefaultDNS64(strconv.Itoa(int(one.ID))); err != nil {
			return nil, err
		}
		dns64s = append(dns64s, tmp)
	}
	return dns64s, nil
}

func (controller *DBController) CreateDNS64(dns64 *DNS64, viewID string) (*tb.DNS64, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	view := tb.View{}
	var num int
	var err error
	if num, err = strconv.Atoi(viewID); err != nil {
		return nil, err
	}
	view.ID = uint(num)
	if err := tx.First(&view).Error; err != nil {
		return nil, err
	}
	tbdns64 := tb.DNS64{}
	tbdns64.Prefix = dns64.Prefix
	if num, err = strconv.Atoi(dns64.ClientACL); err != nil {
		return nil, err
	}
	tbdns64.ClientACL = uint(num)
	if num, err = strconv.Atoi(dns64.AAddress); err != nil {
		return nil, err
	}
	tbdns64.AAddress = uint(num)
	view.DNS64s = append(view.DNS64s, tbdns64)
	if err := tx.Save(&view).Error; err != nil {
		return nil, err
	}
	if err := tx.Last(&tbdns64).Error; err != nil {
		return nil, err
	}
	req := pb.CreateDNS64Req{ID: strconv.Itoa(int(tbdns64.ID)), ViewID: viewID, Prefix: dns64.Prefix, ClientACL: dns64.ClientACL, AAddress: dns64.AAddress}
	data, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATEDNS64); err != nil {
		return nil, err
	}
	tx.Commit()
	return &tbdns64, nil
}

func (controller *DBController) DeleteDNS64(id string, viewID string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	dns64 := tb.DNS64{}
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	dns64.ID = uint(num)
	if err := tx.Unscoped().Delete(&dns64).Error; err != nil {
		return err
	}
	req := pb.DeleteDNS64Req{ID: id, ViewID: viewID}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEDNS64); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) UpdateDNS64(dns64 *DNS64, viewID string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	tbdns64 := tb.DNS64{}
	var num int
	var err error
	if num, err = strconv.Atoi(dns64.ID); err != nil {
		return err
	}
	tbdns64.ID = uint(num)
	if err := controller.CheckACL(dns64.ClientACL); err != nil {
		return err
	}
	if err := controller.CheckACL(dns64.AAddress); err != nil {
		return err
	}
	if err := tx.Find(&tbdns64).Error; err != nil {
		return err
	}
	tbdns64.Prefix = dns64.Prefix
	if num, err = strconv.Atoi(dns64.ClientACL); err != nil {
		return err
	}
	tbdns64.ClientACL = uint(num)
	if num, err = strconv.Atoi(dns64.AAddress); err != nil {
		return err
	}
	tbdns64.AAddress = uint(num)
	if err := tx.Save(&tbdns64).Error; err != nil {
		return err
	}
	req := pb.UpdateDNS64Req{ID: strconv.Itoa(int(tbdns64.ID)), ViewID: viewID, Prefix: dns64.Prefix, ClientACL: dns64.ClientACL, AAddress: dns64.AAddress}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEDNS64); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetDNS64(id string) (*DNS64, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	dns64 := tb.DNS64{}
	if err := tx.First(&dns64, id).Error; err != nil {
		return nil, err
	}
	one := DNS64{}
	one.ID = id
	one.Prefix = dns64.Prefix
	one.ClientACL = strconv.Itoa(int(dns64.ClientACL))
	var acl tb.ACL
	if err := tx.First(&acl, dns64.ClientACL).Error; err != nil {
		return nil, err
	}
	one.ClientACLName = acl.Name
	acl.ID = 0
	one.AAddress = strconv.Itoa(int(dns64.AAddress))
	acl.ID = 0
	if err := tx.First(&acl, dns64.AAddress).Error; err != nil {
		return nil, err
	}
	one.AddressName = acl.Name
	return &one, nil
}

func (controller *DBController) GetDNS64s(viewID string) ([]*DNS64, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	dns64s := []*DNS64{}
	tbDNS64s := []tb.DNS64{}
	if err := tx.Where("view_id = ?", viewID).Find(&tbDNS64s).Error; err != nil {
		return nil, err
	}
	var err error
	for _, one := range tbDNS64s {
		tmp := &DNS64{}
		if tmp, err = controller.GetDNS64(strconv.Itoa(int(one.ID))); err != nil {
			return nil, err
		}
		dns64s = append(dns64s, tmp)
	}
	return dns64s, nil
}

func (controller *DBController) CreateIPBlackHole(blackHole *IPBlackHole) (*tb.IPBlackHole, error) {
	//create the blackHole data.
	tx := controller.db.Begin()
	defer tx.Rollback()
	tbblackHole := tb.IPBlackHole{}
	var num int
	var err error
	if num, err = strconv.Atoi(blackHole.ACLID); err != nil {
		return nil, err
	}
	tbblackHole.ACLID = uint(num)
	if err := tx.Create(&tbblackHole).Error; err != nil {
		return nil, err
	}
	req := pb.CreateIPBlackHoleReq{ID: strconv.Itoa(int(tbblackHole.ID)), ACLID: blackHole.ACLID}
	data, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATEIPBLACKHOLE); err != nil {
		return nil, err
	}
	tx.Commit()
	return &tbblackHole, nil
}

func (controller *DBController) DeleteIPBlackHole(id string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	blackHole := tb.IPBlackHole{}
	var num int
	var err error
	if num, err = strconv.Atoi(id); err != nil {
		return err
	}
	blackHole.ID = uint(num)
	//delete the blackhole data.
	if err := tx.Unscoped().Delete(&blackHole).Error; err != nil {
		return err
	}
	req := pb.DeleteIPBlackHoleReq{ID: id}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETEIPBLACKHOLE); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) UpdateIPBlackHole(blackHole *IPBlackHole) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	tbblackHole := tb.IPBlackHole{}
	var num int
	var err error
	if num, err = strconv.Atoi(blackHole.ID); err != nil {
		return err
	}
	tbblackHole.ID = uint(num)
	if num, err = strconv.Atoi(blackHole.ACLID); err != nil {
		return err
	}
	tbblackHole.ACLID = uint(num)
	if err := tx.Save(&tbblackHole).Error; err != nil {
		return err
	}
	req := pb.UpdateIPBlackHoleReq{ID: blackHole.ID, ACLID: blackHole.ACLID}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATEIPBLACKHOLE); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetIPBlackHole(id string) (*IPBlackHole, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	blackHole := tb.IPBlackHole{}
	if err := tx.First(&blackHole, id).Error; err != nil {
		return nil, err
	}
	one := IPBlackHole{}
	var acl tb.ACL
	if err := tx.First(&acl, blackHole.ACLID).Error; err != nil {
		return nil, fmt.Errorf("the id %d of acl is not exists!", blackHole.ACLID)
	}
	one.ACLID = strconv.Itoa(int(blackHole.ACLID))
	one.ACLName = acl.Name
	one.ID = id
	return &one, nil
}

func (controller *DBController) GetIPBlackHoles() ([]*IPBlackHole, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	blackHoles := []*IPBlackHole{}
	tbBlackHoles := []tb.IPBlackHole{}
	if err := tx.Find(&tbBlackHoles).Error; err != nil {
		return nil, err
	}
	var err error
	for _, one := range tbBlackHoles {
		tmp := &IPBlackHole{}
		if tmp, err = controller.GetIPBlackHole(strconv.Itoa(int(one.ID))); err != nil {
			return nil, err
		}
		blackHoles = append(blackHoles, tmp)
	}
	return blackHoles, nil
}

func (controller *DBController) GetRecursiveConcurrents() ([]*RecursiveConcurrent, error) {
	var con tb.RecursiveConcurrent
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.First(&con).Error; err != nil {
		con.RecursiveClients = 1000
		con.FetchesPerZone = 0
		if err := tx.Create(&con).Error; err != nil {
			return nil, err
		}
	}
	var many []*RecursiveConcurrent
	var tmp RecursiveConcurrent
	tmp.RecursiveClients = int(con.RecursiveClients)
	tmp.FetchesPerZone = int(con.FetchesPerZone)
	many = append(many, &tmp)
	tx.Commit()
	return many, nil
}

func (controller *DBController) UpdateRecursiveConcurrent(max *RecursiveConcurrent) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var con tb.RecursiveConcurrent
	if err := tx.First(&con).Error; err != nil {
		con.RecursiveClients = 1000
		con.FetchesPerZone = 0
		if err := tx.Create(&con).Error; err != nil {
			return err
		}
	}
	con.RecursiveClients = uint(max.RecursiveClients)
	con.FetchesPerZone = uint(max.FetchesPerZone)
	if err := tx.Save(&con).Error; err != nil {
		return err
	}
	req := pb.UpdateRecurConcuReq{RecursiveClients: strconv.Itoa(int(max.RecursiveClients)), FetchesPerZone: strconv.Itoa(int(max.FetchesPerZone))}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATERECURSIVECONCURRENT); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) UpdatePWD(user string, pwd string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	encryptText, err := myaes.Encrypt(controller.aesKey, pwd)
	if err != nil {
		return err
	}
	one := tb.DDIUserPWD{Name: user}
	if err := tx.Find(&one).Error; err != nil {
		return err
	}
	if err := tx.Model(&one).UpdateColumn("pwd", encryptText).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (controller *DBController) GetUserPWD(user string) (*string, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	one := tb.DDIUserPWD{Name: user}
	if err := tx.Find(&one).Error; err != nil {
		return nil, err
	}
	//rawText, err := myaes.Decrypt(controller.aesKey, one.PWD)
	fmt.Println("get user pwd", one.PWD)
	rawText, err := myaes.Decrypt([]byte("linkingthing.com"), one.PWD)
	if err != nil {
		return nil, err
	}
	return &rawText, nil
}

func (controller *DBController) CreateSortList(sorts *SortList) (*tb.SortListElement, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	one := tb.SortListElement{}
	//add the acls to the sortlist
	for k, id := range sorts.ACLIDs {
		//check wether the acl is valid
		var tmp tb.ACL
		if err := tx.First(&tmp, id).Error; err != nil {
			return nil, fmt.Errorf("id %s of acl not exists, %w", id, err)
		}
		one.ACLID = tmp.ID
		if k == 0 {
			one.PrevACLID = "0"
		} else {
			one.PrevACLID = sorts.ACLIDs[k-1]
		}
		if k < len(sorts.ACLIDs)-1 {
			one.NextACLID = sorts.ACLIDs[k+1]
		} else {
			one.NextACLID = "0"
		}
		one.ID = 0
		if err := tx.Create(&one).Error; err != nil {
			return nil, err
		}
	}
	req := pb.CreateSortListReq{ACLIDs: sorts.ACLIDs}
	data, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, CREATESORTLIST); err != nil {
		return nil, err
	}
	tx.Commit()
	return &one, nil
}

func (controller *DBController) DeleteSortList() error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	//delete all the sortlistelement
	var tmp []tb.SortListElement
	if err := tx.Unscoped().Delete(&tmp).Error; err != nil {
		return err
	}
	req := pb.DeleteSortListReq{}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, DELETESORTLIST); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (controller *DBController) GetSortList() (*SortList, error) {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var one tb.SortListElement
	if err := tx.Where("prev_acl_id=?", 0).Find(&one).Error; err != nil {
		return nil, err
	}
	var sortList SortList
	var err error
	var num int
	if num, err = strconv.Atoi(one.NextACLID); err != nil {
		return nil, err
	}
	currentid := uint(num)
	//query the acl info
	var acl tb.ACL
	if err := tx.First(&acl, one.ACLID).Error; err != nil {
		return nil, err
	}
	sortList.ACLIDs = append(sortList.ACLIDs, strconv.Itoa(int(one.ACLID)))
	var acltmp ACL
	acltmp.ID = strconv.Itoa(int(acl.ID))
	acltmp.Name = acl.Name
	sortList.ACLs = append(sortList.ACLs, &acltmp)
	for currentid != 0 {
		var one tb.SortListElement
		if err := tx.Where("acl_id = ?", currentid).Find(&one).Error; err != nil {
			return nil, err
		}
		sortList.ACLIDs = append(sortList.ACLIDs, strconv.Itoa(int(currentid)))
		var acl tb.ACL
		if err := tx.First(&acl, one.ACLID).Error; err != nil {
			return nil, err
		}
		var acltmp ACL
		acltmp.ID = strconv.Itoa(int(acl.ID))
		acltmp.Name = acl.Name
		sortList.ACLs = append(sortList.ACLs, &acltmp)
		var num int
		if num, err = strconv.Atoi(one.NextACLID); err != nil {
			return nil, err
		}
		currentid = uint(num)
	}
	sortList.SetID("1")
	return &sortList, nil
}

func (controller *DBController) UpdateSortList(sortList *SortList) error {
	if err := controller.DeleteSortList(); err != nil {
		return err
	}
	if _, err := controller.CreateSortList(sortList); err != nil {
		return err
	}
	req := pb.UpdateSortListReq{ACLIDs: sortList.ACLIDs}
	data, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	if err := kfkcli.KafkaClient.SendCmd(data, UPDATESORTLIST); err != nil {
		return err
	}

	return nil
}
