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
	FormatError        = goresterr.ErrorCode{"Unauthorized", 400}
)

type dividedAddressHandler struct {
	dividedAddresses *DividedAddressState
}

func NewDividedAddressHandler(s *DividedAddressState) *dividedAddressHandler {
	return &dividedAddressHandler{
		dividedAddresses: s,
	}
}

func (h *dividedAddressHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	dividedAddress := ctx.Resource.(*res.DividedAddress)
	if err := DBCon.UpdateDividedAddress(dividedAddress); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return dividedAddress, nil
}

func (h *dividedAddressHandler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	var err error
	var dividedAddresses []*res.DividedAddress
	filter := ctx.GetFilters()
	var subnetid string
	var ip string
	var hostName string
	var macAddr string
	if len(filter) == 0 {
		return nil, nil
	}
	for _, v := range filter {
		if v.Name == "subnetid" {
			subnetid = v.Values[0]
		}
		if v.Name == "ip" {
			ip = v.Values[0]
		}
		if v.Name == "hostname" {
			hostName = v.Values[0]
		}
		if v.Name == "mac" {
			macAddr = v.Values[0]
		}
	}
	if dividedAddresses, err = DBCon.GetDividedAddresses(subnetid, ip, hostName, macAddr); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return dividedAddresses, nil
}

func (h *dividedAddressHandler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dividedAddressHandler Action, ctx.Resource: ", ctx.Resource)

	r := ctx.Resource
	dividedAddressData, _ := r.GetAction().Input.(*res.DividedAddressData)

	var restRet dhcprest.BaseJsonOptionName
	restRet.Status = "200"
	restRet.Message = "操作成功"

	switch r.GetAction().Name {
	case "change":
		if dividedAddressData.Oper == "tostable" {

			//todo add one stable
			changeData := dividedAddressData.Data

			if len(changeData.CircuitId) > 0 || len(changeData.HwAddress) > 0 {
				//get subnetv4Id and build one restReservation object
				subnetv4Id := changeData.Subnetv4Id
				var restRsv dhcprest.RestReservation
				restRsv.IpAddress = changeData.IpAddress
				restRsv.CircuitId = changeData.CircuitId
				restRsv.HwAddress = changeData.HwAddress
				restRsv.ResvType = "stable"
				if ormRsv, err := dhcprest.PGDBConn.OrmCreateReservation(subnetv4Id, &restRsv); err != nil {

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
				restRsv.ResvType = "resv"

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

	return restRet, nil
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

/*type scanAddressHandler struct {
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
*/
////////////////////
type ipAttrAppendHandler struct {
	ipAttrAppends *IPAttrAppendState
}

func NewIPAttrAppendHandler(s *IPAttrAppendState) *ipAttrAppendHandler {
	return &ipAttrAppendHandler{
		ipAttrAppends: s,
	}
}

func (h *ipAttrAppendHandler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	var err error
	var ipAttrAppend *res.IPAttrAppend
	if ipAttrAppend, err = DBCon.GetIPAttrAppend(ctx.Resource.GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return ipAttrAppend, nil
}

func (h *ipAttrAppendHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	ipAttrAppend := ctx.Resource.(*res.IPAttrAppend)
	if err := DBCon.UpdateIPAttrAppend(ipAttrAppend); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return ipAttrAppend, nil
}

type IPAttrAppendState struct {
	IPAttrAppend []*res.IPAttrAppend
}

func NewIPAttrAppendState() *IPAttrAppendState {
	return &IPAttrAppendState{}
}

//////////////////

type ipv6Prefix struct {
	Prefix string `json:"prefix"`
}

type ipv6AddressTrans struct {
	Prefix string `json:"prefix"`
	Binary string `json:"binary"`
}

func convertToBin(num int) string {
	s := ""

	if num == 0 {
		return "00000000"
	}

	// num /= 2 每次循环的时候 都将num除以2  再把结果赋值给 num
	for ; num > 0; num /= 2 {
		lsb := num % 2
		// strconv.Itoa() 将数字强制性转化为字符串
		s = strconv.Itoa(lsb) + s
	}
	if len(s) < 8 {
		var tmp string
		for i := 0; i < 8-len(s); i++ {
			tmp += "0"
		}
		s = tmp + s
	}
	return s
}

func CheckPrefix(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &ipv6Prefix{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		c.JSON(500, "error")
		return
	}
	s := strings.Split(p.Prefix, "/")
	var L int
	if L, err = strconv.Atoi(s[len(s)-1]); err != nil {
		c.JSON(500, "error")
		return
	}
	fmt.Println(L)
	var ipv6Addr net.IP
	ipv6Addr = net.ParseIP(s[0])
	if ipv6Addr == nil {
		c.JSON(500, "error")
		return
	}
	//M := 8-int(L%8)
	var offset float64
	offset = float64(8 - int(L%8))
	ipv6Addr[L/8] = ipv6Addr[L/8] & (byte(math.Pow(2, 8)) - byte(math.Pow(2, offset)))
	another := &ipv6Prefix{}
	another.Prefix = ipv6Addr.String() + "/" + s[1]
	tmp := ipv6AddressTrans{Prefix: another.Prefix}
	for i := 0; i < 16; i++ {
		tmp.Binary = tmp.Binary + convertToBin(int(ipv6Addr[i]))
	}
	jsonContext, err := json.Marshal(tmp)
	if err != nil {
		c.JSON(500, "error")
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
		c.JSON(500, "error")
		return
	}
	if err = dhcprest.PGDBConn.CreateSubtree(p); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	data, err := json.Marshal(p)
	if err != nil {
		c.JSON(500, "error")
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
		c.JSON(500, "error")
		return
	}
	if err := dhcprest.PGDBConn.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
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
		c.JSON(500, err)
		return
	}
	jsonContext, err := json.Marshal(&data)
	if err != nil {
		c.JSON(500, err)
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
		c.JSON(500, "error")
		return
	}
	if err := dhcprest.PGDBConn.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	p.ID = "0"
	if err = dhcprest.PGDBConn.CreateSubtree(p); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	data, err := json.Marshal(p)
	if err != nil {
		c.JSON(500, "error")
		return
	}
	fmt.Fprintln(c.Writer, string(data))

}

type fileData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetSubtreeMember(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &res.Subtree{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	if err := dhcprest.PGDBConn.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	p.ID = "0"
	if err = dhcprest.PGDBConn.CreateSubtree(p); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	data, err := json.Marshal(p)
	if err != nil {
		c.JSON(500, "error")
		return
	}
	fmt.Fprintln(c.Writer, string(data))

}

func SplitSubnet(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	p := &res.SplitSubnet{}
	err := json.Unmarshal([]byte(body), p)
	if err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	if err := dhcprest.PGDBConn.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	var ret *res.SplitSubnetResult
	if ret, err = dhcprest.PGDBConn.SplitSubnet(p); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	data, err := json.Marshal(ret)
	if err != nil {
		c.JSON(500, "error")
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
