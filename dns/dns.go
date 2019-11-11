package dns

import (
	"bytes"
	"fmt"
	"github.com/ben-han-cn/cement/shell"
	"github.com/boltdb/bolt"
	"github.com/linkingthing/ddi/pb"
	"io/ioutil"
	"os"
	"text/template"
)

type BindHandler struct {
	ConfigPath   string
	MainConfName string
	ConfContent  string
	ViewList     []View
	FreeACLList  map[string]ACL
}

func (t *BindHandler) StartDNS(req pb.DNSStartReq) error {
	tmpl, err := template.ParseFiles(t.ConfigPath + "/templates/named.tpl")
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	tmpl.Execute(buffer, t)
	t.ConfContent = string(buffer.Bytes())
	if len(t.ConfContent) > 0 {
		if err := ioutil.WriteFile(t.ConfigPath+"/"+t.MainConfName, []byte(t.ConfContent), 0644); err != nil {
			return err
		}
		db, err := bolt.Open("my.db", 0600, nil)
		if err != nil {
			return err
		}
		defer db.Close()
		err = db.Update(func(tx *bolt.Tx) error {

			b := tx.Bucket([]byte("binddb"))
			if b == nil {
				b, err = tx.CreateBucket([]byte("binddb"))
				if err != nil {
					return err
				}
			}
			err = b.Put([]byte(t.MainConfName), []byte(t.ConfContent))
			if err != nil {
				return err
			}
			fmt.Println("Put the key value success into binddb!")

			return nil
		})
	}

	var param string = "-c" + t.ConfigPath + "/" + t.MainConfName
	if _, err := shell.Shell("named", param); err != nil {
		return err
	}
	return nil

}

func (t *BindHandler) StopDNS() error {
	var err error
	if _, err = shell.Shell("rndc", "halt"); err != nil {
		return err
	}
	return nil
}

func (t *BindHandler) CreateACL(req pb.CreateACLReq) error {
	acl := ACL{ID: req.ACLID, Name: req.ACLName, IpList: req.IpList}
	t.FreeACLList[req.ACLID] = acl
	t.updateConfigFile()

	return nil
}

func (t *BindHandler) DeleteACL(req pb.DeleteACLReq) error {
	name := t.ConfigPath + "/" + t.FreeACLList[req.ACLID].Name + ".conf"
	if err := os.Remove(name); err != nil {
		return err
	}
	delete(t.FreeACLList, req.ACLID)
	return nil
}

func (t *BindHandler) CreateView(req pb.CreateViewReq) error {
	list := make(map[string]ACL, 10)
	for _, v := range req.ACLIDs {
		oneACL := t.FreeACLList[v]
		list[v] = oneACL
		delete(t.FreeACLList, v)
	}
	oneView := View{ID: req.ViewID, ViewName: req.ViewName, ACLList: list, ZoneList: make(map[string]Zone, 5)}
	if int(req.Priority) > len(t.ViewList) {
		t.ViewList = append(t.ViewList, oneView)
	} else if req.Priority >= 1 {
		t.ViewList = append(t.ViewList[:req.Priority-1], append([]View{oneView}, t.ViewList[req.Priority-1:]...)...)
	}

	tmpl, err := template.ParseFiles(t.ConfigPath + "/templates/named.tpl")
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	tmpl.Execute(buffer, t)
	t.ConfContent = buffer.String()
	if err := ioutil.WriteFile(t.ConfigPath+"/"+t.MainConfName, buffer.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func (t *BindHandler) UpdateView(req pb.UpdateViewReq) error {
	i := 0
	var view View
	for i, view = range t.ViewList {
		if view.ID == req.ViewID {
			break
		}
	}
	for _, ipDel := range req.DeleteIPList {
		for k, ipOld := range view.ACLList[req.ACLID].IpList {
			if ipDel == ipOld {
				acl := view.ACLList[req.ACLID]
				acl.IpList = append(acl.IpList[:k], acl.IpList[k+1:]...)
				t.ViewList[i].ACLList[req.ACLID] = acl
			}
		}
	}
	for _, ip := range req.NewIPList {
		a := t.ViewList[i].ACLList[req.ACLID]
		a.IpList = append(a.IpList, ip)
		t.ViewList[i].ACLList[req.ACLID] = a
	}

	if int(req.Priority-1) != i {
		view = t.ViewList[i]
		t.ViewList = append(t.ViewList[:i], t.ViewList[i+1:]...)
		t.ViewList = append(t.ViewList[:req.Priority-1], append([]View{view}, t.ViewList[req.Priority-1:]...)...)
	}

	if err := t.updateConfigFile(); err != nil {
		return err
	}
	return nil
}

func (t *BindHandler) DeleteView(req pb.DeleteViewReq) error {
	for k, v := range t.ViewList {
		if v.ID == req.ViewID {
			t.ViewList = append(t.ViewList[:k], t.ViewList[k+1:]...)
		}
	}
	t.updateConfigFile()
	return nil
}

func (t *BindHandler) CreateZone(req pb.CreateZoneReq) error {
	zone := Zone{ID: req.ZoneID, ZoneName: req.ZoneName, ZoneFileName: req.ZoneFileName, RRList: make(map[string]RR, 10)}
	for k, view := range t.ViewList {
		if view.ID == req.ViewID {
			t.ViewList[k].ZoneList[req.ZoneID] = zone
			break
		}
	}
	t.updateConfigFile()
	return nil
}

func (t *BindHandler) DeleteZone(req pb.DeleteZoneReq) error {
	for _, view := range t.ViewList {
		if view.ID == req.ViewID {
			zone := view.ZoneList[req.ZoneID]
			name := t.ConfigPath + "/" + zone.ZoneFileName + ".zone"
			if err := os.Remove(name); err != nil {
				return err
			}
			delete(view.ZoneList, req.ZoneID)
			break
		}
	}
	t.updateConfigFile()
	return nil
}

func (t *BindHandler) CreateRR(req pb.CreateRRReq) error {
	for _, view := range t.ViewList {
		if view.ID == req.ViewID {
			zone, ok := view.ZoneList[req.ZoneID]
			if ok == true {
				rr := RR{ID: req.RrID, Data: req.Rrdata}
				zone.RRList[req.RrID] = rr
			}

			break
		}
	}
	t.updateConfigFile()
	return nil
}

func (t *BindHandler) UpdateRR(req pb.UpdateRRReq) error {
	r := pb.CreateRRReq{ViewID: req.ViewID, ZoneID: req.ZoneID, RrID: req.RrID, Rrdata: req.NewrrData}
	t.CreateRR(r)
	return nil
}

func (t *BindHandler) DeleteRR(req pb.DeleteRRReq) error {
	for _, view := range t.ViewList {
		if view.ID == req.ViewID {
			zone, ok := view.ZoneList[req.ZoneID]
			if ok == true {
				delete(zone.RRList, req.RrID)
			}

			break
		}
	}
	t.updateConfigFile()
	return nil
}

func (t *BindHandler) updateConfigFile() error {
	tmpl, err := template.ParseFiles(t.ConfigPath + "/templates/named.tpl")
	if err != nil {
		return err
	}
	tmpl, err = tmpl.ParseFiles(t.ConfigPath + "/templates/acl.tpl")
	if err != nil {
		return err
	}
	tmpl, err = tmpl.ParseFiles(t.ConfigPath + "/templates/zone.tpl")
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	tmpl.ExecuteTemplate(buffer, "named.tpl", t)
	t.ConfContent = string(buffer.Bytes())
	if err := ioutil.WriteFile(t.ConfigPath+"/"+t.MainConfName, []byte(t.ConfContent), 0644); err != nil {
		return err
	}

	for _, view := range t.ViewList {
		for _, acl := range view.ACLList {
			buffer := new(bytes.Buffer)
			tmpl.ExecuteTemplate(buffer, "acl.tpl", acl)
			if err := ioutil.WriteFile(t.ConfigPath+"/"+acl.Name+".conf", buffer.Bytes(), 0644); err != nil {
				return err
			}

		}
	}
	for _, acl := range t.FreeACLList {
		buffer := new(bytes.Buffer)
		tmpl.ExecuteTemplate(buffer, "acl.tpl", acl)
		if err := ioutil.WriteFile(t.ConfigPath+"/"+acl.Name+".conf", buffer.Bytes(), 0644); err != nil {
			return err
		}
	}
	for _, view := range t.ViewList {
		for _, zone := range view.ZoneList {
			buffer := new(bytes.Buffer)
			tmpl.ExecuteTemplate(buffer, "zone.tpl", zone)
			if err := ioutil.WriteFile(t.ConfigPath+"/"+zone.ZoneFileName+".zone", buffer.Bytes(), 0644); err != nil {
				return err
			}

		}
	}
	return nil
}
