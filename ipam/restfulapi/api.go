package restfulapi

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	goresterr "github.com/ben-han-cn/gorest/error"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
	res "github.com/linkingthing/ddi/ipam"
	"io/ioutil"
	"math"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing.com",
		Version: "example/v1",
	}
	dividedAddressKind = resource.DefaultKindName(res.DividedAddress{})
	scanAddressKind    = resource.DefaultKindName(res.ScanAddress{})
	db                 *gorm.DB
	FormatError        = goresterr.ErrorCode{"Unauthorized", 400}
)

/*type res.DividedAddress struct {
	resource.ResourceBase `json:",inline"`
	Reserved              []string `json:"-"`
	Dynamic               []string `json:"-"`
	Stable                []string `json:"-"`
	Manual                []string `json:"-"`
	Lease                 []string `json:"-"`
}*/

type dividedAddressHandler struct {
	dividedAddresses *DividedAddressState
}

func NewDividedAddressHandler(s *DividedAddressState) *dividedAddressHandler {
	return &dividedAddressHandler{
		dividedAddresses: s,
	}
}

func (h *dividedAddressHandler) Get(ctx *resource.Context) resource.Resource {
	var err error
	var dividedAddress *res.DividedAddress
	if dividedAddress, err = dhcprest.PGDBConn.GetDividedAddress(ctx.Resource.GetID()); err != nil {
		return nil
	}
	return dividedAddress
}

type DividedAddressState struct {
	DividedAddress []*res.DividedAddress
}

func NewDividedAddressState() *DividedAddressState {
	return &DividedAddressState{}
}

/////////////////
/*type ScanAddress struct {
	resource.ResourceBase `json:",inline"`
	Collision             []string `json:"acls"`
	Dead                  []string `json:"acls"`
}*/

type scanAddressHandler struct {
	scanAddresses *ScanAddressState
}

func NewScanAddressHandler(s *ScanAddressState) *scanAddressHandler {
	return &scanAddressHandler{
		scanAddresses: s,
	}
}

func (h *scanAddressHandler) Get(ctx *resource.Context) resource.Resource {
	var err error
	var scanAddress *res.ScanAddress
	if scanAddress, err = dhcprest.PGDBConn.GetScanAddress(ctx.Resource.GetID()); err != nil {
		return nil
	}
	return scanAddress
}

type ScanAddressState struct {
	ScanAddress []*res.ScanAddress
}

func NewScanAddressState() *ScanAddressState {
	return &ScanAddressState{}
}

type ipv6Prefix struct {
	Prefix string `json:"prefix"`
}

func CheckPrefix(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &ipv6Prefix{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		return
	}
	s := strings.Split(p.Prefix, "/")
	var L int
	if L, err = strconv.Atoi(s[len(s)-1]); err != nil {
		return
	}
	fmt.Println(L)
	var ipv6Addr net.IP
	ipv6Addr = net.ParseIP(s[0])
	if ipv6Addr == nil {
		return
	}
	//M := 8-int(L%8)
	var offset float64
	offset = float64(8 - int(L%8))
	ipv6Addr[L/8] = ipv6Addr[L/8] & (byte(math.Pow(2, 8)) - byte(math.Pow(2, offset)))
	another := &ipv6Prefix{}
	another.Prefix = ipv6Addr.String() + "/" + s[1]
	jsonContext, err := json.Marshal(another)
	if err != nil {
		return
	}
	fmt.Fprintln(c.Writer, string(jsonContext))
}

func CreateSubtree(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &res.AlloPrefix{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		return
	}
	for i, n := range p.Nodes {
		var ipv6Addr net.IP
		ipv6Addr = net.ParseIP(p.ParentIPv6)
		if ipv6Addr == nil {
			return
		}
		ipv6Addr[(p.PrefixLength+p.BitNum)/8] = ipv6Addr[(p.PrefixLength+p.BitNum)/8] + n.NodeCode*byte(math.Pow(2, float64(8-p.PrefixLength%8-p.BitNum+1)))
		p.Nodes[i].Subnet = ipv6Addr.String() + "/" + strconv.Itoa(int(p.PrefixLength+p.BitNum))
	}
	if err = dhcprest.PGDBConn.CreateSubtree(p); err != nil {
		fmt.Println(err)
		return
	}
	data, err := json.Marshal(p)
	if err != nil {
		return
	}
	fmt.Fprintln(c.Writer, string(data))
}

type idJson struct {
	ID string `json:"id"`
}

func DeleteSubtree(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &idJson{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		return
	}
	if err := dhcprest.PGDBConn.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		return
	}
}

func GetSubtree(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &idJson{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		return
	}
	var data *res.NodesTree
	if data, err = dhcprest.PGDBConn.GetSubtree(p.ID); err != nil {
		fmt.Println(err)
		return
	}
	jsonContext, err := json.Marshal(&data)
	if err != nil {
		return
	}
	fmt.Fprintln(c.Writer, string(jsonContext))
}
