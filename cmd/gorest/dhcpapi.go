package main

import (
	"github.com/ben-han-cn/gorest"
	"github.com/ben-han-cn/gorest/adaptor"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/ben-han-cn/gorest/resource/schema"
	"github.com/gin-gonic/gin"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
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

	dhcpv4 := dhcprest.NewDhcpv4(dhcprest.NewPGDB().DB)
	schemas.Import(&version, dhcprest.Subnetv4{}, dhcprest.NewSubnetv4Handler(dhcpv4))
	subnetv4s := dhcprest.NewSubnetv4s(dhcprest.NewPGDB().DB)
	schemas.Import(&version, dhcprest.RestReservation{}, dhcprest.NewReservationHandler(subnetv4s))

	dhcpv6 := dhcprest.NewDhcpv6(dhcprest.NewPGDB().DB)
	schemas.Import(&version, dhcprest.Subnetv6{}, dhcprest.NewSubnetv6Handler(dhcpv6))

	router := gin.Default()
	adaptor.RegisterHandler(router, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
	router.Run("0.0.0.0:1234")

}
