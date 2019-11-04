package dns

import (
	"fmt"
	"github.com/ben-han-cn/cement/shell"
	"github.com/boltdb/bolt"
	"github.com/linkingthing/ddi/pb"
)

type BindHandler struct {
	ConfigPath   string
	MainConfName string
	ViewList     map[string]View
	FreeACLList  map[string]ACL
}

func (t *BindHandler) StartDNS(request pb.DNSStartReq) error {
	if len(request.Config) > 0 {
		WriteWithIOutil(t.ConfigPath+"/"+t.MainConfName, request.Config)
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
			err = b.Put([]byte(t.MainConfName), []byte(request.Config))
			if err != nil {
				return err
			}
			fmt.Println("Put the key value success into binddb!")

			return nil
		})
	}
	/*var command string = t.ConfigPath + "/named.pid"
	var ret bool
	if ret = shell.IsDirExists(command); ret == false {
		fmt.Println("hasn't started!start now!")
		var start_cmd string = "named"

		if ret, err := shell.Shell(start_cmd); err != nil {
			fmt.Printf("start fail! return message:%s\n", ret)
			fmt.Println(err)
		} else {
			fmt.Printf("start success! return message:%s\n", ret)
		}

	} else if ret {
		fmt.Println("had started!")
	} else {
		fmt.Println("Nothing done!")
	}*/
	var start_cmd string = "named -c " + t.ConfigPath + "/" + t.MainConfName
	if _, err := shell.Shell(start_cmd); err != nil {
		return err
	}
	return nil

}

func (t *BindHandler) StopDNS() {
	return
}

func (t *BindHandler) CreateACL(pb.CreateACLReq) {
	return
}

func (t *BindHandler) DeleteACL(pb.DeleteACLReq) {
	return
}

func (t *BindHandler) CreateView(request pb.CreateViewReq) {
	return
}

func (t *BindHandler) UpdateView(request pb.UpdateViewReq) {
	return
}

func (t *BindHandler) DeleteView(request pb.DeleteViewReq) {
	return
}

func (t *BindHandler) CreateZone(request pb.CreateZoneReq) {
	return
}

func (t *BindHandler) DeleteZone(request pb.DeleteZoneReq) {
	return
}

func (t *BindHandler) CreateRR(request pb.CreateRRReq) {
	return
}

func (t *BindHandler) UpdateRR(request pb.UpdateRRReq) {
	return
}

func (t *BindHandler) DeleteRR(request pb.DeleteRRReq) {
	return
}
