package restfulapi

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"io/ioutil"
	"math"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	res "github.com/linkingthing/ddi/ipam"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing.com",
		Version: "example/v1",
	}
	FormatError = goresterr.ErrorCode{"Unauthorized", 400}
)

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
	if err = DBCon.CreateSubtree(p); err != nil {
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
	if err := DBCon.DeleteSubtree(p.ID); err != nil {
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
	if data, err = DBCon.GetSubtree(id); err != nil {
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
	if err := DBCon.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	p.ID = "0"
	if err = DBCon.CreateSubtree(p); err != nil {
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
	if err := DBCon.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	p.ID = "0"
	if err = DBCon.CreateSubtree(p); err != nil {
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
	if err := DBCon.DeleteSubtree(p.ID); err != nil {
		fmt.Println(err)
		c.JSON(500, "error")
		return
	}
	var ret *res.SplitSubnetResult
	if ret, err = DBCon.SplitSubnet(p); err != nil {
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
	if subtree, err = DBCon.GetSubtree(ctx.Resource.GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return subtree,nil
}

func (h *subtreeHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	subtree := ctx.Resource.(*res.Subtree)
	var err error
	if err = DBCon.CreateSubtree(subtree); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return subtree, nil
}

func (h *subtreeHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	subtree := ctx.Resource.(*res.Subtree)
	if err := DBCon.DeleteSubtree(subtree.GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *subtreeHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	subtree := ctx.Resource.(*res.Subtree)
	var err error
	if err = DBCon.CreateSubtree(subtree); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return subtree, nil
}
*/
