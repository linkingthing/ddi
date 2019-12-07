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

	//dhcpv6 := newDhcpv6()
	//schemas.Import(&version, Subnetv6{}, newSubnetv4Handler(dhcpv6))

	router := gin.Default()
	adaptor.RegisterHandler(router, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
	router.Run("0.0.0.0:1234")

}
