package restfulapi

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"io/ioutil"
	"log"
	"math"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
	res "github.com/linkingthing/ddi/ipam"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing.com",
		Version: "example/v1",
	}
	dividedAddressKind = resource.DefaultKindName(res.DividedAddress{})
	scanAddressKind    = resource.DefaultKindName(res.ScanAddress{})
	//subtreeKind        = resource.DefaultKindName(res.Subtree{})
	db          *gorm.DB
	FormatError = goresterr.ErrorCode{"Unauthorized", 400}
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

func (h *dividedAddressHandler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	var err error
	var dividedAddress *res.DividedAddress
	if dividedAddress, err = dhcprest.PGDBConn.GetDividedAddress(ctx.Resource.GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return dividedAddress, nil
}

func (h *dividedAddressHandler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dividedAddressHandler Action, ctx.Resource: ", ctx.Resource)

	r := ctx.Resource
	dividedAddressData, _ := r.GetAction().Input.(*res.DividedAddressData)

	log.Println("in Action, name: ", r.GetAction().Name)
	log.Println("in Action, oper: ", dividedAddressData.Oper)
	log.Println("in Action, data: ", dividedAddressData.Data)
	switch r.GetAction().Name {
	case "change":
		if dividedAddressData.Oper == "tostable" {
			log.Println("in Action, oper=tostable ")
			//todo add one stable
			changeData := dividedAddressData.Data
			log.Println("in Action,changeData.IpAddress:", changeData.IpAddress)
			if len(changeData.CircuitId) > 0 || len(changeData.HwAddress) > 0 {
				//get subnetv4Id and build one restReservation object
				subnetv4Id := changeData.Subnetv4Id
				var restRsv dhcprest.RestReservation
				restRsv.IpAddress = changeData.IpAddress
				restRsv.CircuitId = changeData.CircuitId
				restRsv.HwAddress = changeData.HwAddress
				if ormRsv, err := dhcprest.PGDBConn.OrmCreateReservation(subnetv4Id, &restRsv); err != nil {
					log.Println("OrmCreateReservation error, restRsv.IpAddress:", restRsv.IpAddress)
					log.Println("newly created ormRsv.ID:", ormRsv.ID)
					restRsv := dhcprest.ConvertReservationFromOrmToRest(&ormRsv)
					log.Println("tostable, ret restRsv.IpAddress:", restRsv.IpAddress)
					return &restRsv, nil
				}

			} else {
				log.Println("change to stable IP error, need CirtcuitId or HwAddress")
				return nil, nil
			}

		}
		if dividedAddressData.Oper == "toresv" {
			log.Println("in Action, oper=toresv ")
			//todo add one resv
			changeData := dividedAddressData.Data
			log.Println("in Action,changeData.IpAddress:", changeData.IpAddress)
			if len(changeData.CircuitId) == 0 && len(changeData.HwAddress) == 0 {
				//get subnetv4Id and build one restReservation object
				subnetv4Id := changeData.Subnetv4Id
				var restRsv dhcprest.RestReservation
				restRsv.Duid = changeData.Duid
				restRsv.Hostname = changeData.Hostname
				restRsv.ClientId = changeData.ClientId
				restRsv.IpAddress = changeData.IpAddress

				if ormRsv, err := dhcprest.PGDBConn.OrmCreateReservation(subnetv4Id, &restRsv); err != nil {
					log.Println("OrmCreateReservation error, restRsv.IpAddress:", restRsv.IpAddress)
					log.Println("newly created ormRsv.ID:", ormRsv.ID)
					restRsv := dhcprest.ConvertReservationFromOrmToRest(&ormRsv)
					log.Println("toresv, ret restRsv.IpAddress:", restRsv.IpAddress)
					return &restRsv, nil
				}

			} else {
				log.Println("change to reserv IP error, CirtcuitId or HwAddress should not exist")
				return nil, nil
			}
		}
	}
	return nil, nil
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

func (h *scanAddressHandler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	var err error
	var scanAddress *res.ScanAddress
	if scanAddress, err = dhcprest.PGDBConn.GetScanAddress(ctx.Resource.GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return scanAddress, nil
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
	p := &res.Subtree{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		fmt.Println(err)
		return
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
	id := c.Query("id")
	/*body, _ := ioutil.ReadAll(c.Request.Body)
	p := &idJson{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		return
	}*/
	var err error
	var data *res.Subtree
	if data, err = dhcprest.PGDBConn.GetSubtree(id); err != nil {
		fmt.Println(err)
		return
	}
	jsonContext, err := json.Marshal(&data)
	if err != nil {
		return
	}
	fmt.Fprintln(c.Writer, string(jsonContext))
}
func UpdateSubtree(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &res.Subtree{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := dhcprest.PGDBConn.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		return
	}
	p.ID = "0"
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

//new resource subtree
/*type subtreeHandler struct {
	subtrees *SubtreeState
}

func NewSubtreeHandler(s *SubtreeState) *subtreeHandler {
	return &subtreeHandler{
		subtrees: s,
	}
}

type SubtreeState struct {
	Subtree []*res.Subtree
}

func NewSubtreeState() *SubtreeState {
	return &SubtreeState{}
}

func (h *subtreeHandler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	var err error
	var subtree *res.Subtree
	if subtree, err = dhcprest.PGDBConn.GetSubtree(ctx.Resource.GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return subtree,nil
}

func (h *subtreeHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	subtree := ctx.Resource.(*res.Subtree)
	var err error
	if err = dhcprest.PGDBConn.CreateSubtree(subtree); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return subtree, nil
}

func (h *subtreeHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	subtree := ctx.Resource.(*res.Subtree)
	if err := dhcprest.PGDBConn.DeleteSubtree(subtree.GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *subtreeHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	subtree := ctx.Resource.(*res.Subtree)
	var err error
	if err = dhcprest.PGDBConn.CreateSubtree(subtree); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return subtree, nil
}
*/
