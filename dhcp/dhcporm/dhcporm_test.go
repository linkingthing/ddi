package dhcporm

import (
	"log"
	"testing"

	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
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
	db, err := gorm.Open("postgres", dhcprest.CRDBAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dhcpv4 := dhcprest.NewDhcpv4(dhcprest.NewPGDB().DB)

	s4 := dhcprest.Subnetv4{
		Name:          "subnetname2",
		ValidLifetime: "3333",
	}
	err = dhcpv4.CreateSubnetv4(&s4)
	ut.Assert(t, err == nil, "create subnetv4 ok")

	var subnets *dhcprest.Subnetv4
	subnets = dhcpv4.GetSubnetv4ById(s4.ID)

	if len(subnets.ID) > 0 {
		subnets.Name = "subnetname3"
		dhcpv4.UpdateSubnetv4(subnets)
	}

	err = dhcpv4.DeleteSubnetv4(subnets)
	ut.Assert(t, err == nil, "delete subnetv4 ok", subnets.ID)
}
