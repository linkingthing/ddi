package dns

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"text/template"

	"github.com/ben-han-cn/cement/shell"
	kv "github.com/ben-han-cn/kvzoo"
	"github.com/ben-han-cn/kvzoo/backend/bolt"
	"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils/rrupdate"
)

const (
	mainConfName = "named.conf"
	dBName       = "bind.db"
	viewsPath    = "/views/"
	viewsEndPath = "/views"
	zonesPath    = "/zones/"
	zonesEndPath = "/zones"
	aCLsPath     = "/acls/"
	aCLsEndPath  = "/acls"
	rRsEndPath   = "/rrs"
	iPsEndPath   = "/ips"
	namedTpl     = "named.tpl"
	zoneTpl      = "zone.tpl"
	aCLTpl       = "acl.tpl"
	nzfTpl       = "nzf.tpl"
	rndcPort     = "953"
	rrKey        = "key1"
	rrSecret     = "linking_encr"
)

type BindHandler struct {
	tpl         *template.Template
	db          kv.DB
	dnsConfPath string
	dBPath      string
	tplPath     string
}

func NewBindHandler(dnsConfPath string, agentPath string) *BindHandler {
	var tmpDnsPath string
	if dnsConfPath[len(dnsConfPath)-1] != '/' {
		tmpDnsPath = dnsConfPath + "/"
	} else {
		tmpDnsPath = dnsConfPath
	}
	var tmpDBPath string
	if agentPath[len(agentPath)-1] != '/' {
		tmpDBPath = agentPath + "/"
	} else {
		tmpDBPath = agentPath
	}

	instance := &BindHandler{dnsConfPath: tmpDnsPath, dBPath: tmpDBPath, tplPath: tmpDBPath + "templates/"}
	pbolt, err := bolt.New(agentPath + dBName)
	if err != nil {
		return nil
	}
	instance.db = pbolt
	instance.tpl, err = template.ParseFiles(instance.tplPath + namedTpl)
	if err != nil {
		return nil
	}
	instance.tpl, err = instance.tpl.ParseFiles(instance.tplPath + zoneTpl)
	if err != nil {
		return nil
	}
	instance.tpl, err = instance.tpl.ParseFiles(instance.tplPath + aCLTpl)
	if err != nil {
		return nil
	}
	instance.tpl, err = instance.tpl.ParseFiles(instance.tplPath + nzfTpl)
	if err != nil {
		return nil
	}
	return instance
}

type namedData struct {
	ConfigPath string
	Views      []View
}

type View struct {
	Name string
	ACLs []ACL
}

type nzfData struct {
	ViewName string
	Zones    []Zone
}

type Zone struct {
	Name     string
	ZoneFile string
}

type zoneData struct {
	ViewName string
	Name     string
	ZoneFile string
	RRs      []RR
}

type RR struct {
	Data string
}

type ACL struct {
	ID   string
	Name string
	IPs  []string
}

func (handler *BindHandler) StartDNS(req pb.DNSStartReq) error {
	if _, err := os.Stat(handler.dnsConfPath + "named.pid"); err == nil {
		return nil
	}
	if err := handler.rewriteNamedFile(); err != nil {
		return err
	}
	if err := handler.rewriteACLsFile(); err != nil {
		return err
	}

	if err := handler.rewriteZonesFile(); err != nil {
		return err
	}

	if err := handler.rewriteNzfsFile(); err != nil {
		return err
	}

	var param string = "-c" + handler.dnsConfPath + mainConfName
	if _, err := shell.Shell("named", param); err != nil {
		return err
	}
	return nil

}

func (handler *BindHandler) StopDNS() error {
	var err error
	if _, err = shell.Shell("rndc", "halt"); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) CreateACL(req pb.CreateACLReq) error {
	err := handler.addKVs(aCLsPath+req.ACLID, map[string][]byte{"name": []byte(req.ACLName)})
	if err != nil {
		return err
	}
	values := map[string][]byte{}
	for _, ip := range req.IPList {
		values[ip] = []byte("")
	}
	if err := handler.addKVs(aCLsPath+req.ACLID+iPsEndPath, values); err != nil {
		return err
	}
	aCLData := ACL{ID: req.ACLID, Name: req.ACLName, IPs: req.IPList}
	buffer := new(bytes.Buffer)
	if err = handler.tpl.ExecuteTemplate(buffer, aCLTpl, aCLData); err != nil {
		return err
	}
	if err := ioutil.WriteFile(handler.dnsConfPath+req.ACLName+".conf", buffer.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (handler *BindHandler) DeleteACL(req pb.DeleteACLReq) error {
	kvs, err := handler.tableKVs(aCLsPath + req.ACLID)
	if err != nil {
		return err
	}
	if len(kvs) == 0 {
		return nil
	}
	name := kvs["name"]
	if err := os.Remove(handler.dnsConfPath + string(name) + ".conf"); err != nil {
		return err
	}

	handler.db.DeleteTable(kv.TableName(aCLsPath + req.ACLID))
	return nil
}

func (handler *BindHandler) CreateView(req pb.CreateViewReq) error {
	if err := handler.addPriority(int(req.Priority), req.ViewID); err != nil {
		return err
	}
	//create table viewid and put name into the db.
	namekvs := map[string][]byte{"name": []byte(req.ViewName)}
	if err := handler.addKVs(viewsPath+req.ViewID, namekvs); err != nil {
		return err
	}
	aCLs, err := handler.aCLsFromTopPath(req.ACLIDs)
	if err != nil {
		return err
	}
	//insert aCLs into viewid table
	if err := handler.insertACLs(req.ViewID, aCLs); err != nil {
		return err
	}
	if err := handler.rewriteNamedFile(); err != nil {
		return err
	}
	//update bind
	if err := handler.rndcReconfig(); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) UpdateView(req pb.UpdateViewReq) error {
	if err := handler.updatePriority(int(req.Priority), req.ViewID); err != nil {
		return err
	}
	//add new ips for aCL
	ipsMap := map[string][]byte{}
	for _, ip := range req.NewIPList {
		ipsMap[ip] = []byte("")
	}
	if err := handler.addKVs(viewsPath+req.ViewID+aCLsPath+req.ACLID+iPsEndPath, ipsMap); err != nil {
		return err
	}
	//delete ips for aCL
	if err := handler.deleteKVs(viewsPath+req.ViewID+aCLsPath+req.ACLID+iPsEndPath, req.DeleteIPList); err != nil {
		return err
	}
	if err := handler.rewriteNamedFile(); err != nil {
		return err
	}
	if err := handler.rewriteACLsFile(); err != nil {
		return err
	}
	//update bind
	if err := handler.rndcReconfig(); err != nil {
		return err
	}

	return nil
}

func (handler *BindHandler) DeleteView(req pb.DeleteViewReq) error {
	handler.deletePriority(req.ViewID)
	//delete table
	if err := handler.db.DeleteTable(kv.TableName(viewsPath + req.ViewID)); err != nil {
		return err
	}
	if err := handler.rewriteNamedFile(); err != nil {
		return nil
	}
	if err := handler.rewriteZonesFile(); err != nil {
		return nil
	}
	if err := handler.rewriteACLsFile(); err != nil {
		return nil
	}
	if err := handler.rewriteNzfsFile(); err != nil {
		return nil
	}
	if err := handler.rndcReconfig(); err != nil {
		return nil
	}
	return nil
}

func (handler *BindHandler) CreateZone(req pb.CreateZoneReq) error {
	//put the zone into db
	names := map[string][]byte{}
	names["name"] = []byte(req.ZoneName)
	names["zonefile"] = []byte(req.ZoneFileName)
	if err := handler.addKVs(viewsPath+req.ViewID+zonesPath+req.ZoneID, names); err != nil {
		return err
	}
	//update file
	if err := handler.rewriteZonesFile(); err != nil {
		return err
	}
	var err error
	out := map[string][]byte{}
	if out, err = handler.tableKVs(viewsPath + req.ViewID); err != nil {
		return err
	}
	viewName := out["name"]
	if err := handler.rndcAddZone(req.ZoneName, req.ZoneFileName, string(viewName)); err != nil {
		return err
	}

	return nil
}

func (handler *BindHandler) DeleteZone(req pb.DeleteZoneReq) error {
	var names map[string][]byte
	var err error
	if names, err = handler.tableKVs(viewsPath + req.ViewID + zonesPath + req.ZoneID); err != nil {
		return err
	}
	zoneName := names["name"]
	zoneFile := names["zonefile"]
	var out map[string][]byte
	if out, err = handler.tableKVs(viewsPath + req.ViewID); err != nil {
		return err
	}
	viewName := out["name"]
	if err := handler.db.DeleteTable(kv.TableName(viewsPath + req.ViewID + zonesPath + req.ZoneID)); err != nil {
		return nil
	}
	if err := handler.rndcDelZone(string(zoneName), string(zoneFile), string(viewName)); err != nil {
		return err
	}

	if err := os.Remove(handler.dnsConfPath + string(zoneFile)); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) CreateRR(req pb.CreateRRReq) error {
	rrsMap := map[string][]byte{}
	rrsMap[req.RRID] = []byte(req.RRData)
	if err := handler.addKVs(viewsPath+req.ViewID+zonesPath+req.ZoneID+rRsEndPath, rrsMap); err != nil {
		return err
	}
	var names map[string][]byte
	var err error
	if names, err = handler.tableKVs(viewsPath + req.ViewID + zonesPath + req.ZoneID); err != nil {
		return err
	}
	if err := rrupdate.UpdateRR(rrKey, rrSecret, req.RRData, string(names["name"]), true); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) UpdateRR(req pb.UpdateRRReq) error {
	var rrsMap map[string][]byte
	var err error
	if rrsMap, err = handler.tableKVs(viewsPath + req.ViewID + zonesPath + req.ZoneID + rRsEndPath); err != nil {
		return err
	}
	rrData := rrsMap[req.RRID]
	var names map[string][]byte
	if names, err = handler.tableKVs(viewsPath + req.ViewID + zonesPath + req.ZoneID); err != nil {
		return err
	}
	if err := rrupdate.UpdateRR(rrKey, rrSecret, string(rrData), string(names["name"]), false); err != nil { //string(rrData[:])
		return err
	}
	if err := rrupdate.UpdateRR(rrKey, rrSecret, req.NewRRData, string(names["name"]), true); err != nil {
		return err
	}
	rrsMap[req.RRID] = []byte(req.NewRRData)
	if err := handler.updateKVs(viewsPath+req.ViewID+zonesPath+req.ZoneID+rRsEndPath, rrsMap); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) DeleteRR(req pb.DeleteRRReq) error {
	var rrsMap map[string][]byte
	var err error
	if rrsMap, err = handler.tableKVs(viewsPath + req.ViewID + zonesPath + req.ZoneID + rRsEndPath); err != nil {
		return err
	}
	rrData := rrsMap[req.RRID]
	var names map[string][]byte
	if names, err = handler.tableKVs(viewsPath + req.ViewID + zonesPath + req.ZoneID); err != nil {
		return err
	}
	if err := rrupdate.UpdateRR(rrKey, rrSecret, string(rrData), string(names["name"]), false); err != nil { //string(rrData[:])
		return err
	}
	if err := handler.deleteKVs(viewsPath+req.ViewID+zonesPath+req.ZoneID+rRsEndPath, []string{req.RRID}); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) namedConfData() (namedData, error) {
	var err error
	data := namedData{ConfigPath: handler.dnsConfPath}
	var kvs map[string][]byte
	kvs, err = handler.tableKVs(viewsEndPath)
	if err != nil {
		return data, err
	}
	if len(kvs) == 0 {
		return data, nil
	}
	for _, viewid := range kvs {
		nameKvs, err := handler.tableKVs(viewsPath + string(viewid))
		if err != nil {
			return data, err
		}
		viewName := nameKvs["name"]
		tables, err := handler.tables(viewsPath + string(viewid) + aCLsEndPath)
		if err != nil {
			return data, err
		}
		var aCLs []ACL
		for _, aCLid := range tables {
			aCLNames, err := handler.tableKVs(viewsPath + string(viewid) + aCLsPath + aCLid)
			if err != nil {
				return data, err
			}
			aCLName := aCLNames["name"]
			ipsMap, err := handler.tableKVs(viewsPath + string(viewid) + aCLsPath + aCLid + iPsEndPath)
			if err != nil {
				return data, err
			}
			var ips []string
			for _, ip := range ipsMap {
				ips = append(ips, string(ip))
			}
			aCL := ACL{Name: string(aCLName)}
			aCLs = append(aCLs, aCL)
		}
		view := View{Name: string(viewName), ACLs: aCLs}
		data.Views = append(data.Views, view)
	}
	return data, nil
}

func (handler *BindHandler) aCLsData() ([]ACL, error) {
	var err error
	var aCLsData []ACL
	var viewTables []string
	viewTables, err = handler.tables(viewsEndPath)
	if err != nil {
		return nil, err
	}
	for _, viewid := range viewTables {
		var aCLs []ACL
		aCLTables, err := handler.tables(viewsPath + viewid + aCLsEndPath)
		if err != nil {
			return nil, err
		}
		for _, aCLid := range aCLTables {
			aCLNames, err := handler.tableKVs(viewsPath + viewid + aCLsPath + aCLid)
			if err != nil {
				return nil, err
			}
			aCLName := aCLNames["name"]
			ipsMap, err := handler.tableKVs(viewsPath + viewid + aCLsPath + aCLid + iPsEndPath)
			if err != nil {
				return nil, err
			}
			var ips []string
			for ip, _ := range ipsMap {
				ips = append(ips, ip)
			}
			aCL := ACL{Name: string(aCLName), IPs: ips}
			aCLs = append(aCLs, aCL)
		}
		aCLsData = append(aCLsData, aCLs[0:]...)
	}
	var aCLTables []string
	aCLTables, err = handler.tables(aCLsEndPath)
	if err != nil {
		return nil, err
	}
	for _, aCLId := range aCLTables {
		names, err := handler.tableKVs(aCLsPath + aCLId)
		if err != nil {
			return nil, err
		}
		aCLName := names["name"]
		ipsMap, err := handler.tableKVs(aCLsPath + aCLId + iPsEndPath)
		if err != nil {
			return nil, err
		}
		var ips []string
		for ip, _ := range ipsMap {
			ips = append(ips, ip)
		}
		oneACL := ACL{Name: string(aCLName), IPs: ips}
		aCLsData = append(aCLsData, oneACL)
	}
	return aCLsData, nil
}

func (handler *BindHandler) nzfsData() ([]nzfData, error) {
	var data []nzfData
	viewIDs, err := handler.tables(viewsEndPath)
	if err != nil {
		return nil, err
	}
	for _, viewid := range viewIDs {
		nameKvs, err := handler.tableKVs(viewsPath + viewid)
		if err != nil {
			return nil, err
		}
		viewName := nameKvs["name"]
		zoneTables, err := handler.tables(viewsPath + viewid + zonesEndPath)
		if err != nil {
			return nil, err
		}
		var zones []Zone
		for _, zoneId := range zoneTables {
			Names, err := handler.tableKVs(viewsPath + viewid + zonesPath + zoneId)
			if err != nil {
				return nil, err
			}
			zoneName := Names["name"]
			zoneFile := Names["zonefile"]
			zone := Zone{Name: string(zoneName), ZoneFile: string(zoneFile)}
			zones = append(zones, zone)
		}
		oneNzfData := nzfData{ViewName: string(viewName), Zones: zones}
		data = append(data, oneNzfData)
	}
	return data, nil
}

func (handler *BindHandler) zonesData() ([]zoneData, error) {
	var zonesData []zoneData
	viewIDs, err := handler.tables(viewsEndPath)
	if err != nil {
		return nil, err
	}
	for _, viewid := range viewIDs {
		nameKvs, err := handler.tableKVs(viewsPath + viewid)
		if err != nil {
			return nil, err
		}
		viewName := nameKvs["name"]
		zoneTables, err := handler.tables(viewsPath + viewid + zonesEndPath)
		if err != nil {
			return nil, err
		}
		for _, zoneID := range zoneTables {
			var rrs []RR
			names, err := handler.tableKVs(viewsPath + viewid + zonesPath + zoneID)
			if err != nil {
				return nil, err
			}
			zoneName := names["name"]
			zoneFile := names["zonefile"]
			datas, err := handler.tableKVs(viewsPath + viewid + zonesPath + zoneID + rRsEndPath)
			if err != nil {
				return nil, err
			}
			for _, data := range datas {
				rr := RR{Data: string(data)}
				rrs = append(rrs, rr)
			}
			one := zoneData{ViewName: string(viewName), Name: string(zoneName), ZoneFile: string(zoneFile), RRs: rrs}
			zonesData = append(zonesData, one)
		}
	}
	return zonesData, nil
}

func (handler *BindHandler) tableKVs(table string) (map[string][]byte, error) {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(table))
	if err != nil {
		return nil, err
	}
	var ts kv.Transaction
	if ts, err = tb.Begin(); err != nil {
		return nil, err
	}
	defer ts.Rollback()
	kvs, err := ts.List()
	if err != nil {
		return nil, err
	}
	return kvs, nil
}

func (handler *BindHandler) tables(table string) ([]string, error) {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(table))
	if err != nil {
		return nil, err
	}
	var ts kv.Transaction
	if ts, err = tb.Begin(); err != nil {
		return nil, err
	}
	defer ts.Rollback()
	tables, err := ts.Tables()
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (handler *BindHandler) addKVs(tableName string, values map[string][]byte) error {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(tableName))
	if err != nil {
		return err
	}
	var ts kv.Transaction
	ts, err = tb.Begin()
	if err != nil {
		return err
	}
	defer ts.Rollback()
	for k, value := range values {
		if err := ts.Add(k, value); err != nil {
			return err
		}
	}
	if err := ts.Commit(); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) updateKVs(tableName string, values map[string][]byte) error {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(tableName))
	if err != nil {
		return err
	}
	var ts kv.Transaction
	ts, err = tb.Begin()
	if err != nil {
		return err
	}
	defer ts.Rollback()
	for k, value := range values {
		if err := ts.Update(k, value); err != nil {
			return err
		}
	}
	if err := ts.Commit(); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) deleteKVs(tableName string, keys []string) error {
	tb, err := handler.db.CreateOrGetTable(kv.TableName(tableName))
	if err != nil {
		return err
	}
	var ts kv.Transaction
	ts, err = tb.Begin()
	if err != nil {
		return err
	}
	defer ts.Rollback()
	for _, key := range keys {
		if err := ts.Delete(key); err != nil {
			return err
		}
	}
	if err := ts.Commit(); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) rewriteNamedFile() error {
	var namedConfData namedData
	var err error
	if namedConfData, err = handler.namedConfData(); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	if err = handler.tpl.ExecuteTemplate(buffer, namedTpl, namedConfData); err != nil {
		return err
	}
	if err := ioutil.WriteFile(handler.dnsConfPath+mainConfName, buffer.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) rewriteZonesFile() error {
	zonesData, err := handler.zonesData()
	if err != nil {
		return err
	}
	for _, zoneData := range zonesData {
		buf := new(bytes.Buffer)
		if err = handler.tpl.ExecuteTemplate(buf, zoneTpl, zoneData); err != nil {
			return err
		}
		if err := ioutil.WriteFile(handler.dnsConfPath+zoneData.ZoneFile, buf.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (handler *BindHandler) rewriteACLsFile() error {
	aCLs, err := handler.aCLsData()
	if err != nil {
		return err
	}
	for _, aCL := range aCLs {
		buf := new(bytes.Buffer)
		if err = handler.tpl.ExecuteTemplate(buf, aCLTpl, aCL); err != nil {
			return err
		}
		if err := ioutil.WriteFile(handler.dnsConfPath+aCL.Name+".conf", buf.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (handler *BindHandler) rewriteNzfsFile() error {
	nzfsData, err := handler.nzfsData()
	if err != nil {
		return err
	}
	for _, nzfData := range nzfsData {
		buf := new(bytes.Buffer)
		if err = handler.tpl.ExecuteTemplate(buf, nzfTpl, nzfData); err != nil {
			return err
		}
		if err := ioutil.WriteFile(handler.dnsConfPath+nzfData.ViewName+".nzf", buf.Bytes(), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (handler *BindHandler) addPriority(pri int, viewid string) error {
	var kvs map[string][]byte
	var err error
	if kvs, err = handler.tableKVs(viewsEndPath); err != nil {
		return err
	}
	if pri > len(kvs)+1 {
		pri = len(kvs) + 1
	} else if pri < 1 {
		pri = 1
	}
	i := len(kvs)
	for i >= pri {
		kvs[strconv.Itoa(i+1)] = kvs[strconv.Itoa(i)]
		i--
	}
	kvs[strconv.Itoa(pri)] = []byte(viewid)
	if len(kvs) > 1 {
		addKVs := map[string][]byte{strconv.Itoa(len(kvs)): kvs[strconv.Itoa(len(kvs))]}
		if err := handler.addKVs(viewsEndPath, addKVs); err != nil {
			return err
		}
		delete(kvs, strconv.Itoa(len(kvs)))
		if err := handler.updateKVs(viewsEndPath, kvs); err != nil {
			return err
		}
	} else {
		if err := handler.addKVs(viewsEndPath, kvs); err != nil {
			return err
		}
	}

	return nil
}

func (handler *BindHandler) updatePriority(pri int, viewid string) error {
	var kvs map[string][]byte
	var err error
	if kvs, err = handler.tableKVs(viewsEndPath); err != nil {
		return err
	}
	if pri > len(kvs) {
		pri = len(kvs)
	} else if pri < 1 {
		pri = 1
	}
	var oriIndex string
	var v []byte
	for oriIndex, v = range kvs {
		if string(v) == viewid {
			break
		}
	}
	key := strconv.Itoa(pri)
	if oriIndex != key {
		tmp := kvs[key]
		kvs[key] = kvs[oriIndex]
		kvs[oriIndex] = tmp
		if err := handler.updateKVs(viewsEndPath, kvs); err != nil {
			return err
		}
	}
	return nil
}

func (handler *BindHandler) deletePriority(viewID string) error {
	//query priority
	kvs, err := handler.tableKVs(viewsEndPath)
	if err != nil {
		return err
	}
	//delete priority
	var k string
	var v []byte
	for k, v = range kvs {
		if string(v) == viewID {
			break
		}
	}
	if err = handler.deleteKVs(viewsEndPath, []string{strconv.Itoa(len(kvs))}); err != nil {
		return err
	}
	delete(kvs, k)
	i := 1
	for _, v := range kvs {
		kvs[strconv.Itoa(i)] = v
		i++
	}
	if len(kvs) == 0 {
		return nil
	}
	if err = handler.updateKVs(viewsEndPath, kvs); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) aCLsFromTopPath(aCLids []string) ([]ACL, error) {
	var aCLs []ACL
	for _, aCLid := range aCLids {
		names, err := handler.tableKVs(aCLsPath + aCLid)
		if err != nil {
			return nil, err
		}
		name := names["name"]
		var ipsmap map[string][]byte
		ipsmap, err = handler.tableKVs(aCLsPath + aCLid + iPsEndPath)
		if err != nil {
			return nil, err
		}
		var ips []string
		for ip, _ := range ipsmap {
			ips = append(ips, ip)
		}
		aCL := ACL{Name: string(name), IPs: ips, ID: aCLid}
		aCLs = append(aCLs, aCL)
	}
	return aCLs, nil
}
func (handler *BindHandler) insertACLs(viewid string, aCLs []ACL) error {
	for _, aCL := range aCLs {
		nameMap := map[string][]byte{"name": []byte(aCL.Name)}
		if err := handler.addKVs(viewsPath+viewid+aCLsPath+aCL.ID, nameMap); err != nil {
			return err
		}
		ipsMap := map[string][]byte{}
		for _, ip := range aCL.IPs {
			ipsMap[ip] = []byte("")
		}
		if err := handler.addKVs(viewsPath+viewid+aCLsPath+aCL.ID+iPsEndPath, ipsMap); err != nil {
			return err
		}
	}
	return nil
}

func (handler *BindHandler) rndcReconfig() error {
	//update bind
	var para1 string = "-c" + handler.dnsConfPath + "rndc.conf"
	var para2 string = "-s" + "localhost"
	var para3 string = "-p" + rndcPort
	var para4 string = "reconfig"
	if _, err := shell.Shell("rndc", para1, para2, para3, para4); err != nil {
		return err
	}
	return nil
}
func (handler *BindHandler) rndcAddZone(name string, zoneFile string, viewName string) error {
	//update bind
	var para1 string = "-c" + handler.dnsConfPath + "rndc.conf"
	var para2 string = "-s" + "localhost"
	var para3 string = "-p" + rndcPort
	var para4 string = "addzone " + name + " in " + viewName + " { type master; file \"" + zoneFile + "\";};"
	if _, err := shell.Shell("rndc", para1, para2, para3, para4); err != nil {
		return err
	}
	return nil
}

func (handler *BindHandler) rndcDelZone(name string, zoneFile string, viewName string) error {
	//update bind
	var para1 string = "-c" + handler.dnsConfPath + "rndc.conf"
	var para2 string = "-s" + "localhost"
	var para3 string = "-p" + rndcPort
	var para4 string = "delzone " + name + " in " + viewName
	if _, err := shell.Shell("rndc", para1, para2, para3, para4); err != nil {
		return err
	}
	return nil
}
