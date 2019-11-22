package dhcporm

import (
	"log"
	"testing"

	"fmt"

	ut "github.com/ben-han-cn/cement/unittest"
	"github.com/jinzhu/gorm"
)

func TestListSubnetv4(t *testing.T) {

	const addr = "postgresql://maxroach@localhost:26257/postgres?ssl=true&sslmode=require&sslrootcert=/root/download/cockroach-v19.2.0/certs/ca.crt&sslkey=/root/download/cockroach-v19.2.0/certs/client.maxroach.key&sslcert=/root/download/cockroach-v19.2.0/certs/client.maxroach.crt"
	db, err := gorm.Open("postgres", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	list := Subnetv4List(db, "4", "1.1.1.1")
	if len(list) == 3 {
		fmt.Println("len(list) = 3")
	}

	ut.Assert(t, err == nil, "list subnetv4 ok")
}
