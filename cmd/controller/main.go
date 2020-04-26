package main

import (
	"fmt"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/lifei6671/gocaptcha"
	physicalMetrics "github.com/linkingthing/ddi/cmd/metrics"
	"github.com/linkingthing/ddi/cmd/node"
	metric "github.com/linkingthing/ddi/cmd/websocket/server"
	dnsapi "github.com/linkingthing/ddi/dns/restfulapi"
	"github.com/linkingthing/ddi/ipam"
	ipamapi "github.com/linkingthing/ddi/ipam/restfulapi"
	"github.com/linkingthing/ddi/utils/config"
	"github.com/zdnscloud/gorest"
	"github.com/zdnscloud/gorest/adaptor"
	"github.com/zdnscloud/gorest/resource"
	"github.com/zdnscloud/gorest/resource/schema"

	//"github.com/linkingthing/ddi/pb"
	"github.com/linkingthing/ddi/utils"
	kfkcli "github.com/linkingthing/ddi/utils/kafkaclient"

	//"google.golang.org/grpc"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing.com",
		Version: "example/v1",
	}
	checkValueLock   sync.Mutex
	checkValues      []data
	aCLKind          = resource.DefaultKindName(dnsapi.ACL{})
	viewKind         = resource.DefaultKindName(dnsapi.View{})
	zoneKind         = resource.DefaultKindName(dnsapi.Zone{})
	rRKind           = resource.DefaultKindName(dnsapi.RR{})
	forwardKind      = resource.DefaultKindName(dnsapi.Forward{})
	redirectionKind  = resource.DefaultKindName(dnsapi.Redirection{})
	defaultDNS64Kind = resource.DefaultKindName(dnsapi.DefaultDNS64{})
	dNS64Kind        = resource.DefaultKindName(dnsapi.DNS64{})
	ipBlackHoleKind  = resource.DefaultKindName(dnsapi.IPBlackHole{})
	recursiveConKind = resource.DefaultKindName(dnsapi.RecursiveConcurrent{})
	sortListKind     = resource.DefaultKindName(dnsapi.SortList{})
	ipAddressKind    = resource.DefaultKindName(dhcprest.IPAddress{})
	//scanAddressKind    = resource.DefaultKindName(ipam.ScanAddress{})
	subtreeKind = resource.DefaultKindName(ipam.Subtree{})
)

const (
	dx              = 150
	dy              = 50
	delay           = 120000
	checkValueToken = "CheckValueToken"
	checkValue      = "CheckValue"
	checkDuration   = 24 * time.Second
)

type data struct {
	InsertTime int64
	Value      string
}

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "id"

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(identityKey)
	c.JSON(200, gin.H{
		"userID":   claims[identityKey],
		"userName": user.(*User).UserName,
		"text":     "Hello World.",
	})
}

// User demo
type User struct {
	UserName  string
	FirstName string
	LastName  string
}

func main() {
	var err error
	var db *gorm.DB
	db, err = gorm.Open("postgres", utils.DBAddr)
	if err != nil {
		panic(err)
	}

	utils.SetHostIPs(config.YAML_CONFIG_FILE) //set global vars from yaml conf
	kfkcli.KafkaClient = kfkcli.NewKafkaCliHandler("dns", "dhcpv4", "dhcpv6", utils.KafkaServerProm)
	go getKafkaMsg()
	go node.RegisterNode("/etc/vanguard/vanguard.conf", "controller")
	go physicalMetrics.NodeExporter()
	dnsapi.DBCon = dnsapi.NewDBController(db)
	ipamapi.DBCon = ipamapi.NewDBController(db)
	defer dnsapi.DBCon.Close()
	schemas := schema.NewSchemaManager()
	aCLsState := dnsapi.NewACLsState()
	forwardState := dnsapi.NewForwardState()
	dnsState := dnsapi.NewDefaultDNS64State()
	blackHoleState := dnsapi.NewIPBlackHoleState()
	conState := dnsapi.NewRecursiveConcurrentState()
	sortListsState := dnsapi.NewSortListsState()
	schemas.MustImport(&version, dnsapi.ACL{}, dnsapi.NewACLHandler(aCLsState))
	schemas.MustImport(&version, dnsapi.Forward{}, dnsapi.NewForwardHandler(forwardState))
	schemas.MustImport(&version, dnsapi.DefaultDNS64{}, dnsapi.NewDefaultDNS64Handler(dnsState))
	schemas.MustImport(&version, dnsapi.IPBlackHole{}, dnsapi.NewIPBlackHoleHandler(blackHoleState))
	schemas.MustImport(&version, dnsapi.RecursiveConcurrent{}, dnsapi.NewRecursiveConcurrentHandler(conState))
	schemas.MustImport(&version, dnsapi.SortList{}, dnsapi.NewSortListHandler(sortListsState))
	state := dnsapi.NewViewsState()
	schemas.MustImport(&version, dnsapi.View{}, dnsapi.NewViewHandler(state))
	schemas.MustImport(&version, dnsapi.Zone{}, dnsapi.NewZoneHandler(state))
	schemas.MustImport(&version, dnsapi.RR{}, dnsapi.NewRRHandler(state))
	schemas.MustImport(&version, dnsapi.Redirection{}, dnsapi.NewRedirectionHandler(state))
	schemas.MustImport(&version, dnsapi.DNS64{}, dnsapi.NewDNS64Handler(state))
	//ipam interfaces

	// web socket server, consume kafka topic prom and check ping/pong msg
	port := utils.WebSocket_Port
	go metric.SocketServer(port)
	log.Println("Starting dhcp gorest controller")

	// start of dhcp model
	dhcprest.PGDBConn = dhcprest.NewPGDB(db)
	defer dhcprest.PGDBConn.Close()

	dhcpv4 := dhcprest.NewDhcpv4(db)
	schemas.MustImport(&version, dhcprest.RestSubnetv4{}, dhcprest.NewSubnetv4Handler(dhcpv4))
	schemas.MustImport(&version, dhcprest.RestSubnetv46{}, dhcprest.NewSubnetv46Handler(dhcpv4))
	//schemas.MustImport(&version, .RestSubnetv46{}, ipamapi.NewSubnetv46Handler(dhcpv4))
	subnetv4s := dhcprest.NewSubnetv4s(db)
	schemas.MustImport(&version, dhcprest.RestReservation{}, dhcprest.NewReservationHandler(subnetv4s))
	schemas.MustImport(&version, dhcprest.RestPool{}, dhcprest.NewPoolHandler(subnetv4s))
	schemas.MustImport(&version, dhcprest.RestOptionName{}, dhcprest.NewOptionNameHandler(subnetv4s))

	dhcpv6 := dhcprest.NewDhcpv6(db)
	schemas.MustImport(&version, dhcprest.RestSubnetv6{}, dhcprest.NewSubnetv6Handler(dhcpv6))
	subnetv6s := dhcprest.NewSubnetv6s(db)
	schemas.MustImport(&version, dhcprest.RestPoolv6{}, dhcprest.NewPoolv6Handler(subnetv6s))

	schemas.MustImport(&version, dhcprest.IPAddress{}, dhcprest.NewIPAddressHandler(subnetv4s))
	schemas.MustImport(&version, dhcprest.IPAttrAppend{}, dhcprest.NewIPAttrAppendHandler(subnetv4s))
	// end of dhcp model

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
	err = gocaptcha.ReadFonts("fonts", ".ttf")
	router.GET("/apis/linkingthing.com/example/v1", Index)
	router.GET("/apis/linkingthing.com/example/v1/getcheckimage.jpeg", Get)
	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour * 24,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				UserName: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := loginVals.Username
			pwd, err := dnsapi.DBCon.GetUserPWD(userID)
			if err != nil {
				return nil, err
			}
			if loginVals.Password == *pwd {
				return &User{
					UserName:  userID,
					LastName:  "Bo-Yi",
					FirstName: "Wu",
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*User); ok && v.UserName == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	router.POST("/apis/linkingthing.com/example/v1/login", authMiddleware.LoginHandler)
	router.GET("/apis/linkingthing.com/example/v1/checkvalue", CheckValue)
	auth := router.Group("/")
	//auth.Use(authMiddleware.MiddlewareFunc())
	{
		adaptor.RegisterHandler(auth, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
		auth.POST("/apis/linkingthing.com/example/v1/changepwd", ChangePWD)
		auth.POST("/apis/linkingthing.com/example/v1/logout", authMiddleware.LogoutHandler)
		auth.GET("/apis/linkingthing.com/example/v1/nodes", nodeQuery)
		auth.GET("/apis/linkingthing.com/example/v1/hists", nodeQueryRange)
		auth.GET("/apis/linkingthing.com/example/v1/servers", nodeServers)
		auth.GET("/apis/linkingthing.com/example/v1/dashdns", nodeDashDns)
		auth.GET("/apis/linkingthing.com/example/v1/dashdhcpassign", nodeDashDhcpAssign)
		auth.GET("/apis/linkingthing.com/example/v1/retcode", retCodeHandler)
		auth.GET("/apis/linkingthing.com/example/v1/memhit", memHitHandler)
		auth.POST("/apis/linkingthing.com/example/v1/checkipv6prefix", ipamapi.CheckPrefix)
		auth.POST("/apis/linkingthing.com/example/v1/createsubtree", ipamapi.CreateSubtree)
		auth.POST("/apis/linkingthing.com/example/v1/deletesubtree", ipamapi.DeleteSubtree)
		auth.GET("/apis/linkingthing.com/example/v1/getsubtree", ipamapi.GetSubtree)
		auth.POST("/apis/linkingthing.com/example/v1/updatesubtree", ipamapi.UpdateSubtree)
		auth.POST("/apis/linkingthing.com/example/v1/splitsubnet", ipamapi.SplitSubnet)
	}
	router.StaticFS("/public", http.Dir("/opt/website"))
	go CheckValueDestroy()
	router.Run("0.0.0.0:8081")
}

func Index(c *gin.Context) {
	t, err := template.ParseFiles("tpl/index.html")
	if err != nil {
		fmt.Println(err)
	}
	_ = t.Execute(c.Writer, nil)
}

func Get(c *gin.Context) {
	captchaImage := gocaptcha.NewCaptchaImage(dx, dy, gocaptcha.RandLightColor())

	text := gocaptcha.RandText(4)
	fmt.Println(text)
	checkValueLock.Lock()
	t := time.Now().UnixNano() / 1e6
	one := data{InsertTime: t, Value: text}
	fmt.Println(one)
	checkValues = append(checkValues, one)
	checkValueLock.Unlock()
	err := captchaImage.DrawNoise(gocaptcha.CaptchaComplexLower).
		DrawTextNoise(gocaptcha.CaptchaComplexLower).
		DrawText(text).
		DrawBorder(gocaptcha.ColorToRGB(0x17A7A7A)).
		DrawSineLine().Error

	if err != nil {
		fmt.Println(err)
	}

	c.Header("test", "test")
	//c.Writer.Header().Set(checkValueToken, strconv.FormatInt(t, 10))
	c.Header(checkValueToken, strconv.FormatInt(t, 10))
	_ = captchaImage.SaveImage(c.Writer, gocaptcha.ImageFormatJpeg)
}

func CheckValueDestroy() {
	for {
		tmp := data{}
		if len(checkValues) > 0 {
			checkValueLock.Lock()
			tmp.InsertTime = checkValues[0].InsertTime
			tmp.Value = checkValues[0].Value
			checkValueLock.Unlock()
			currTime := time.Now().UnixNano() / 1e6
			if tmp.InsertTime < currTime-2*delay {
				checkValueLock.Lock()
				fmt.Println(checkValues[0])
				checkValues = checkValues[1:]
				checkValueLock.Unlock()
			} else {
				time.Sleep(time.Duration((tmp.InsertTime+2*delay-currTime)/1000) * time.Second)
			}
		} else {
			time.Sleep(time.Duration(2*delay/1000) * time.Second)
			continue
		}
	}
}

func CheckValue(c *gin.Context) {
	insertTime, ok := c.GetQuery(checkValueToken)
	if !ok {
		c.String(200, "err, token required or format is not correct!!")
		return
	}
	value, ok := c.GetQuery(checkValue)
	if !ok {
		c.String(200, "err, checknumber required or format is not correct!")
		return
	}
	currTime := time.Now().UnixNano() / 1e6
	var num int64
	var err error
	num, err = strconv.ParseInt(insertTime, 10, 64)
	if err != nil {
		c.String(200, "token error!")
	}
	if num < currTime-delay {
		c.String(200, "the check value is expire!", currTime-delay-num)
		return
	}
	checkValueLock.Lock()
	notExist := true
	for i, v := range checkValues {
		if v.InsertTime == num {
			if strings.ToLower(v.Value) == strings.ToLower(value) {
				c.String(200, "check value success!")
				notExist = false
			}
			checkValues = append(checkValues[:i], checkValues[i+1:]...)
			break
		}
	}
	if notExist {
		c.String(200, "check value fail!")
	}
	checkValueLock.Unlock()
}

func ChangePWD(c *gin.Context) {
	var loginVals login
	if err := c.ShouldBind(&loginVals); err != nil {
		c.String(200, "username or password value format is not correct!")
		return
	}
	if err := dnsapi.DBCon.UpdatePWD(loginVals.Username, loginVals.Password); err != nil {
		c.String(200, "change password value fail!")
		return
	}
	c.String(200, "change password success!")
	return
}
func nodeQuery(c *gin.Context) {
	metric.Query(c.Writer, c.Request)
}
func nodeQueryRange(c *gin.Context) {
	metric.Query_range(c.Writer, c.Request)
}
func nodeServers(c *gin.Context) {
	metric.List_server(c.Writer, c.Request)
}
func nodeDashDns(c *gin.Context) {
	metric.GetDashDns(c.Writer, c.Request)
}
func nodeDashDhcpAssign(c *gin.Context) {
	metric.GetDashDns(c.Writer, c.Request)
}

func getKafkaMsg() {
	log.Println("into getKafkaMsg")
	for {

		utils.ConsumerProm()
		time.Sleep(checkDuration)
		//time.Sleep(20 * time.Second)
	}
}

func retCodeHandler(c *gin.Context) {
	client := &http.Client{}
	startTime := c.Query("start")
	var numStart int64
	var err error
	if numStart, err = strconv.ParseInt(startTime, 10, 64); err != nil {
		return
	}
	endTime := c.Query("end")
	var step int
	if endTime == "" {
		numStart++
		endTime = strconv.FormatInt(numStart, 10)
		step = 10
	} else {
		var numEnd int64
		if numEnd, err = strconv.ParseInt(endTime, 10, 64); err != nil {
			return
		}
		if numEnd > numStart {
			step = int((numEnd - numStart) / 30)
			fmt.Println("step is:", step)
		}
		if step > 10000 {
			step = 10000
		}
	}
	host := c.Query("node")
	url := "http://" + utils.PromServer + ":" + utils.PromPort + "/api/v1/query_range?" +
		"query=dns_counter%7Bdata_type%3D~%22SERVFAIL%7CNXDOMAIN%7CNOERROR%7CREFUSED%22%2Cinstance%3D%22" +
		host + "%3A8001%22%2Cjob%3D%22dns_exporter%22%7D%20&start=" + startTime + "&end=" + endTime +
		"&step=" + strconv.Itoa(step)
	fmt.Println(url)
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	response, _ := client.Do(reqest)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	fmt.Fprintln(c.Writer, string(body))
}

func memHitHandler(c *gin.Context) {
	client := &http.Client{}
	startTime := c.Query("start")
	var numStart int64
	var err error
	if numStart, err = strconv.ParseInt(startTime, 10, 64); err != nil {
		return
	}
	endTime := c.Query("end")
	var step int
	if endTime == "" {
		numStart++
		endTime = strconv.FormatInt(numStart, 10)
		step = 10
	} else {
		var numEnd int64
		if numEnd, err = strconv.ParseInt(endTime, 10, 64); err != nil {
			return
		}
		if numEnd > numStart {
			step = int((numEnd - numStart) / 30)
			fmt.Println("step is:", step)
		}
		if step > 10000 {
			step = 10000
		}
	}
	host := c.Query("node")
	url := "http://" + utils.PromServer + ":" + utils.PromPort + "/api/v1/query_range?query=dns_counter%7Bdata_type%3D~%22memhit%7Cquerys%22%2Cinstance%3D%22" + host + "%3A8001%22%7D&start=" + startTime + "&end=" + endTime + "&step=" + strconv.Itoa(step)
	fmt.Println(url)
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	response, _ := client.Do(reqest)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	fmt.Fprintln(c.Writer, string(body))
}
