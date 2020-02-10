package main

import (
	"fmt"
	"github.com/ben-han-cn/gorest"
	"github.com/ben-han-cn/gorest/adaptor"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/ben-han-cn/gorest/resource/schema"
	"github.com/gin-gonic/gin"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
	"time"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing",
		Version: "dhcp/v1",
	}
)

func main() {

	dhcprest.PGDBConn = dhcprest.NewPGDB()
	defer dhcprest.PGDBConn.Close()
	schemas := schema.NewSchemaManager()

	//auth := dhcprest.NewAuth(dhcprest.NewPGDB().DB)
	//schemas.Import(&version, dhcprest.AuthRest{}, dhcprest.NewAuthHandler(auth))

	dhcpv4 := dhcprest.NewDhcpv4(dhcprest.NewPGDB().DB)
	schemas.Import(&version, dhcprest.Subnetv4{}, dhcprest.NewSubnetv4Handler(dhcpv4))
	subnetv4s := dhcprest.NewSubnetv4s(dhcprest.NewPGDB().DB)
	schemas.Import(&version, dhcprest.RestReservation{}, dhcprest.NewReservationHandler(subnetv4s))

	dhcpv6 := dhcprest.NewDhcpv6(dhcprest.NewPGDB().DB)
	schemas.Import(&version, dhcprest.Subnetv6{}, dhcprest.NewSubnetv6Handler(dhcpv6))

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
	router.Run("0.0.0.0:1234")

}
