package dns

import (
	"fmt"
	"github.com/ben-han-cn/cement/shell"
	"github.com/boltdb/bolt"
	"github.com/linkingthing/ddi/pb"
	"io/ioutil"
	"os"
)

type BindHandler struct {
	ConfigPath   string
	MainConfName string
	ConfContent  string
	ViewList     map[int]View
	FreeACLList  map[string]ACL
}

func (t *BindHandler) StartDNS(req pb.DNSStartReq) error {
	if len(req.Config) > 0 {
		t.ConfContent = req.Config
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
			err = b.Put([]byte(t.MainConfName), []byte(req.Config))
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
	var fileContent string = "acl \"" + req.ACLName + "\" {\n\t"
	for _, ip := range req.IpList {
		fileContent += ip + ";\n\t"
	}
	fileContent += "};\n"
	if err := ioutil.WriteFile(t.ConfigPath+"/"+req.ACLName+".conf", []byte(fileContent), 0644); err != nil {
		return err
	}
	t.ConfContent += "include \"" + t.ConfigPath + "/" + req.ACLName + ".conf\";\n"
	if err := ioutil.WriteFile(t.ConfigPath+"/"+t.MainConfName, []byte(t.ConfContent), 0644); err != nil {
		return err
	}

	return nil
}

func (t *BindHandler) DeleteACL(req pb.DeleteACLReq) error {
	os.RemoveALL(t.ConfigPath + t.FreeACLList[req.ACLID].Name + ".conf")
	delete(t.FreeACLList, req.ACLID)
	return nil
}

func (t *BindHandler) CreateView(req pb.CreateViewReq) error {
	return nil
}

func (t *BindHandler) UpdateView(req pb.UpdateViewReq) error {
	return nil
}

func (t *BindHandler) DeleteView(req pb.DeleteViewReq) error {
	return nil
}

func (t *BindHandler) CreateZone(req pb.CreateZoneReq) error {
	return nil
}

func (t *BindHandler) DeleteZone(req pb.DeleteZoneReq) error {
	return nil
}

func (t *BindHandler) CreateRR(req pb.CreateRRReq) error {
	return nil
}

func (t *BindHandler) UpdateRR(req pb.UpdateRRReq) error {
	return nil
}

func (t *BindHandler) DeleteRR(req pb.DeleteRRReq) error {
	return nil
}
