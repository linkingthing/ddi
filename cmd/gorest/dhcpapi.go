package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/cmd/websocket/server"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
	"github.com/linkingthing/ddi/ipam"
	ipamapi "github.com/linkingthing/ddi/ipam/restfulapi"
	"github.com/linkingthing/ddi/utils"
	"github.com/linkingthing/ddi/utils/config"
	"github.com/zdnscloud/gorest"
	"github.com/zdnscloud/gorest/adaptor"
	"github.com/zdnscloud/gorest/resource"
	"github.com/zdnscloud/gorest/resource/schema"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing",
		Version: "dhcp/v1",
	}
)

func main() {

	//var db *gorm.DB
	db, err := gorm.Open("postgres", utils.DBAddr)
	if err != nil {
		panic(err)
	}
	utils.SetHostIPs(config.YAML_CONFIG_FILE) //set global vars from yaml conf

	dhcprest.PGDBConn = dhcprest.NewPGDB(db)
	defer dhcprest.PGDBConn.Close()
	schemas := schema.NewSchemaManager()

	//auth := dhcprest.NewAuth(dhcprest.NewPGDB().DB)
	//schemas.Import(&version, dhcprest.AuthRest{}, dhcprest.NewAuthHandler(auth))

	// start of dhcp model
	//go dhcpv4agent.Dhcpv4Client()
	dhcpv4 := dhcprest.NewDhcpv4(db)
	err = schemas.Import(&version, dhcprest.RestSubnetv4{}, dhcprest.NewSubnetv4Handler(dhcpv4))
	if err != nil {
		log.Println("schemas import err: ", err)
	}
	subnetv4s := dhcprest.NewSubnetv4s(db)
	schemas.Import(&version, dhcprest.RestReservation{}, dhcprest.NewReservationHandler(subnetv4s))
	schemas.Import(&version, dhcprest.RestPool{}, dhcprest.NewPoolHandler(subnetv4s))
	schemas.Import(&version, dhcprest.RestOptionName{}, dhcprest.NewOptionNameHandler(subnetv4s))

	devidedAddressState := ipamapi.NewDividedAddressState()
	schemas.Import(&version, ipam.DividedAddress{}, ipamapi.NewDividedAddressHandler(devidedAddressState))

	//state := dhcprest.NewOptionNamesState()
	//err = schemas.Import(&version, dhcprest.RestOptionName{}, dhcprest.NewOptionNameHandler(state))
	//if err != nil {
	//	log.Println("schemas import err: ", err)
	//}
	// end of dhcp model

	//dhcpv6 := dhcprest.NewDhcpv6(dhcprest.NewPGDB().DB)
	//schemas.Import(&version, dhcprest.Subnetv6{}, dhcprest.NewSubnetv6Handler(dhcpv6))

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

	// web socket server, consume kafka topic prom and check ping/pong msg
	port := utils.WebSocket_Port

	go server.SocketServer(port)

	log.Println("Starting dhcp gorest controller")
	router.GET("/apis/linkingthing/dashboard/v1/dashdns", nodeGetDashDns)
	router.GET("/apis/linkingthing/dashboard/v1/dhcpassign", nodeGetDhcpAssign)

	router.Run("0.0.0.0:1235")
	//router.GET("/apis/linkingthing/dashboard/v1/dashdns", nodeGetDashDns)

}

func nodeGetDashDns(c *gin.Context) {
	server.GetDashDns(c.Writer, c.Request)
}
func nodeGetDhcpAssign(c *gin.Context) {
	server.DashDhcpAssign(c.Writer, c.Request)
}
