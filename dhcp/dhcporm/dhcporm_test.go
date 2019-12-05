package dhcporm

import (
	"log"
	"testing"

	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/jinzhu/gorm"
)

//func TestListSubnetv4(t *testing.T) {
//
//	//const addr = "postgresql://maxroach@localhost:26257/postgres?ssl=true&sslmode=require&sslrootcert=/root/download/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/download/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/download/cockroach-v19.2.0/certs/client.maxroach.crt"
//	db, err := gorm.Open("postgres", CRDBAddr)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	list := Subnetv4List(db)
//	if len(list) == 4 {
//		fmt.Println("len(list) = 3")
//	}
//
//	ut.Assert(t, err == nil, "list subnetv4 ok")
//}
func TestCreateSubnetv4(t *testing.T) {

	//const addr = "postgresql://maxroach@localhost:26257/postgres?ssl=true&sslmode=require&sslrootcert=/root/download/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/download/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/download/cockroach-v19.2.0/certs/client.maxroach.crt"
	db, err := gorm.Open("postgres", CRDBAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = CreateSubnetv4(db, "subnetname2", "3333")
	ut.Assert(t, err == nil, "create subnetv4 ok")

	var subnets []Subnetv4
	subnets = GetSubnetv4(db, "subnetname2")

	if len(subnets) > 0 {
		UpdateSubnetv4(db, "subnetname2", "55555")
	}

	err = DeleteSubnetv4(db, "subnetname2")
	ut.Assert(t, err == nil, "delete subnetv4 ok")
}
