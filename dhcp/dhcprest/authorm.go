package dhcprest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"sync"
)

type AuthOrm struct {
	db   *gorm.DB
	user []*User
	lock sync.Mutex
}

type User struct {
	db        *gorm.DB
	username  string
	password  string
	lastLogin string
}

func (handler *PGDB) CheckLogin(db *gorm.DB, name string, passwd string) error {
	var user = dhcporm.OrmUser{
		Username: name,
		Password: passwd,
	}

	//todo check user and passwd
	query := db.Create(&user)

	if query.Error != nil {
		return fmt.Errorf("login  error, subnet name: " + name)
	}

	return nil
}
