package main

import (
	"github.com/ben-han-cn/gorest"
	"github.com/ben-han-cn/gorest/adaptor"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/ben-han-cn/gorest/resource/schema"
	"github.com/gin-gonic/gin"
	api "github.com/linkingthing/ddi/dns/restfulapi"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing.com",
		Version: "example/v1",
	}
)

func main() {
	api.DBCon = api.NewDBController()
	defer api.DBCon.Close()
	schemas := schema.NewSchemaManager()
	aCLsState := api.NewACLsState()
	schemas.Import(&version, api.ACL{}, api.NewACLHandler(aCLsState))
	state := api.NewViewsState()
	schemas.Import(&version, api.View{}, api.NewViewHandler(state))
	schemas.Import(&version, api.Zone{}, api.NewZoneHandler(state))
	router := gin.Default()
	adaptor.RegisterHandler(router, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
	router.Run("0.0.0.0:1234")
}
