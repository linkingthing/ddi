package main

import (
	"fmt"
	"github.com/appleboy/gin-jwt"
	"github.com/ben-han-cn/gorest"
	"github.com/ben-han-cn/gorest/adaptor"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/ben-han-cn/gorest/resource/schema"
	"github.com/gin-gonic/gin"
	"github.com/lifei6671/gocaptcha"
	api "github.com/linkingthing/ddi/dns/restfulapi"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing.com",
		Version: "example/v1",
	}
	checkValueLock sync.Mutex
	checkValues    []data
)

const (
	dx              = 150
	dy              = 50
	delay           = 120000
	checkValueToken = "CheckValueToken"
	checkValue      = "CheckValue"
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
	api.DBCon = api.NewDBController()
	defer api.DBCon.Close()
	schemas := schema.NewSchemaManager()
	aCLsState := api.NewACLsState()
	forwardState := api.NewForwardState()
	dnsState := api.NewDefaultDNS64State()
	blackHoleState := api.NewIPBlackHoleState()
	conState := api.NewRecursiveConcurrentState()
	schemas.Import(&version, api.ACL{}, api.NewACLHandler(aCLsState))
	schemas.Import(&version, api.Forward{}, api.NewForwardHandler(forwardState))
	schemas.Import(&version, api.DefaultDNS64{}, api.NewDefaultDNS64Handler(dnsState))
	schemas.Import(&version, api.IPBlackHole{}, api.NewIPBlackHoleHandler(blackHoleState))
	schemas.Import(&version, api.RecursiveConcurrent{}, api.NewRecursiveConcurrentHandler(conState))
	state := api.NewViewsState()
	schemas.Import(&version, api.View{}, api.NewViewHandler(state))
	schemas.Import(&version, api.Zone{}, api.NewZoneHandler(state))
	schemas.Import(&version, api.RR{}, api.NewRRHandler(state))
	schemas.Import(&version, api.Redirection{}, api.NewRedirectionHandler(state))
	schemas.Import(&version, api.DNS64{}, api.NewDNS64Handler(state))
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
	err := gocaptcha.ReadFonts("fonts", ".ttf")
	router.GET("/apis/linkingthing.com/example/v1", Index)
	router.GET("/apis/linkingthing.com/example/v1/getcheckimage.jpeg", Get)
	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
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
			password := loginVals.Password

			if (userID == "admin" && password == "admin") || (userID == "test" && password == "test") {
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
	router.POST("/apis/linkingthing.com/example/v1/logout", authMiddleware.LogoutHandler)

	router.GET("/apis/linkingthing.com/example/v1/checkvalue", CheckValue)
	auth := router.Group("/")
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		adaptor.RegisterHandler(auth, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
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
