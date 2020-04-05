package dhcprest

import (
	"fmt"
	"github.com/jinzhu/gorm"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
	"log"
	"time"
)

func NewAuth(db *gorm.DB) *AuthOrm {
	return &AuthOrm{db: db}
}

type authHandler struct {
	auth *AuthOrm
}

func NewAuthHandler(a *AuthOrm) *authHandler {
	return &authHandler{
		auth: a,
	}
}

func (s *AuthOrm) LoginOrm(ao *AuthRest) error {
	fmt.Println("into LoginOrm")

	s.lock.Lock()
	defer s.lock.Unlock()

	err := PGDBConn.CheckLogin(s.db, ao.Username, ao.Password)
	if err != nil {
		return err
	}

	return nil
}

func (h *authHandler) Login(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Login")

	ar := ctx.Resource.(*AuthRest)
	ar.SetID(ar.ID)
	ar.SetCreationTimestamp(time.Now())
	if err := h.auth.LoginOrm(ar); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	} else {
		return ar, nil
	}

}
