package dns

import (
	"github.com/linkingthing/ddi/pb"
)

type ACL struct {
	ID     string
	Name   string
	IpList []string
}

func (t *ACL) CreateACL(request pb.CreateACLReq) {
	return
}

func (t *ACL) DeleteACL(request pb.DeleteACLReq) {
	return
}

/*
import (
		"fmt"
			"github.com/boltdb/bolt"
				"log"
			)*/

/*var DNSPath string

func init() {
	DNSPath = "/root/bindtest/"
}
func CreateACL(aCLName string, aCLID string, ipList []string) error {
	fileName := DNSPath + aCLName
	fileName += ".conf"
	var fileContent string
	fileContent = "acl \"" + aCLName + "\"{\n"
	for _, one := range ipList {
		fileContent += one + ";\n"
	}
	fileContent += "};\n"
	WriteWithIOutil(fileName, fileContent)
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

				log.Fatal(err)
			}

		}
		err = b.Put([]byte(aCLID+"ACLName"), []byte(aCLName))
		if err != nil {
			return err
		}

		err = b.Put([]byte(aCLID+"ACLContent"), []byte(fileContent))
		if err != nil {
			return err
		}

		fmt.Println("Put the key value success into binddb!")

		return nil
	})

	return nil
}*/
