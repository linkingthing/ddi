package main

import (
	"fmt"
	"github.com/ben-han-cn/gorest"
	"github.com/ben-han-cn/gorest/adaptor"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/ben-han-cn/gorest/resource/schema"
	"github.com/gin-gonic/gin"
	api "github.com/linkingthing/ddi/dns/restfulapi"
	"net/http"
	"time"
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
	schemas.Import(&version, api.RR{}, api.NewRRHandler(state))
	router := gin.Default()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] client:%s \"%s %s\" %s %d %s %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
		)
	}))
	adaptor.RegisterHandler(router, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
	router.StaticFS("/public", http.Dir("/opt/website"))
	router.Run("0.0.0.0:8081")
}
