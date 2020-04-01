package dhcprest

import (
	"fmt"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"github.com/nexus166/trimv4/ipv4"
	pl "github.com/nexus166/trimv4/process_lists"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing",
		Version: "dhcp/v1",
	}

	subnetv4Kind    = resource.DefaultKindName(RestSubnetv4{})
	ReservationKind = resource.DefaultKindName(RestReservation{})
	PoolKind        = resource.DefaultKindName(RestPool{})
	OptionKind      = resource.DefaultKindName(RestOption{})

	db *gorm.DB
)

//type Dhcpv4Serv struct {
//	resource.ResourceBase `json:",inline"`
//	ConfigJson            string `json:"configJson" rest:"required=true,minLen=1,maxLen=1000000"`
//}

type RestOption struct {
	resource.ResourceBase `json:"embedded,inline"`
	AlwaysSend            bool   `gorm:"column:always-send"`
	Code                  uint64 `gorm:"column:code"`
	CsvFormat             bool   `json:"csv-format"`
	Data                  string `json:"data"`
	Name                  string `json:"name"`
	Space                 string `json:"space"`
}

type RestReservation struct {
	resource.ResourceBase `json:"embedded,inline"`
	BootFileName          string `json:"boot-file-name"`
	//ClientClasses []interface{} `json:"client-classes"`
	//ClientId string `json:"client-id"` //reservations can be multi-types, need to split  todo
	Duid           string       `json:"duid"`
	Hostname       string       `json:"hostname"`
	IpAddress      string       `json:"ip-address"`
	NextServer     string       `json:"next-server"`
	OptionData     []RestOption `json:"option-data"`
	ServerHostname string       `json:"server-hostname"`
}

type RestPool struct {
	resource.ResourceBase `json:"embedded,inline"`
	Subnetv4Id            string       `json:"subnetv4Id"`
	OptionData            []RestOption `json:"optionData"`
	BeginAddress          string       `json:"beginAddress,omitempty" rest:"required=true,minLen=1,maxLen=12"`
	EndAddress            string       `json:"endAddress,omitempty" rest:"required=true,minLen=1,maxLen=12"`
	MaxValidLifetime      int          `json:"maxValidLifetime,omitempty"`
	ValidLifetime         int          `json:"validLifetime,omitempty"`
}

//type Subnetv4 struct {
//	resource.ResourceBase `json:"embedded,inline"`
//	Name                  string `json:"name,omitempty" rest:"required=true,minLen=1,maxLen=255"`
//	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
//	SubnetId              string `json:"subnet_id"`
//	ValidLifetime         string `json:"validLifetime"`
//	Reservations          []*RestReservation
//	Pools                 []*RestPool
//}
type RestSubnetv4 struct {
	resource.ResourceBase `json:"embedded,inline"`
	Name                  string `json:"name,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	SubnetId              string `json:"subnet_id"`
	ValidLifetime         string `json:"validLifetime"`
	Reservations          []*RestReservation
	Pools                 []*RestPool
	SubnetTotal           string `json:"total"`
	SubnetUsage           string `json:"usage"`
}

func (s4 RestSubnetv4) CreateAction(name string) *resource.Action {
	switch name {
	case "mergesplit":
		return &resource.Action{
			Name:  "mergesplit",
			Input: &MergeSplitData{},
		}
	default:
		return nil
	}
}

type MergeSplitData struct {
	resource.ResourceBase `json:",inline"`
	Oper                  string `json:"oper" rest:"required=true,minLen=1,maxLen=20"`
	Mask                  string `json:"mask" rest:"required=true,minLen=1,maxLen=20"`
	IPs                   string `json:"ips" rest:"required=true"`
}

type Subnetv4State struct {
	Subnetv4s []*RestSubnetv4
}
type Subnetv4Handler struct {
	subnetv4s *Subnetv4State
}

func NewSubnetv4State() *Subnetv4State {
	return &Subnetv4State{}
}

type PoolsState struct {
	Pools []*RestPool
}

func NewPoolsState() *PoolsState {
	return &PoolsState{}
}

type Dhcpv4 struct {
	db        *gorm.DB
	subnetv4s []*RestSubnetv4
	lock      sync.Mutex
}

type Subnetv4s struct {
	Subnetv4s []*RestSubnetv4
	db        *gorm.DB
}

func NewSubnetv4s(db *gorm.DB) *Subnetv4s {
	return &Subnetv4s{db: db}
}

type ReservationHandler struct {
	subnetv4s *Subnetv4s
	db        *gorm.DB
	lock      sync.Mutex
}

func NewReservationHandler(s *Subnetv4s) *ReservationHandler {
	return &ReservationHandler{
		subnetv4s: s,
		db:        s.db,
	}
}

type PoolHandler struct {
	subnetv4s *Subnetv4s
	db        *gorm.DB
	lock      sync.Mutex
}

func NewPoolHandler(s *Subnetv4s) *PoolHandler {
	return &PoolHandler{
		subnetv4s: s,
		db:        s.db,
	}
}

//tools func
func ConvertReservationsFromOrmToRest(rs []dhcporm.Reservation) []*RestReservation {

	var restRs []*RestReservation
	for _, v := range rs {
		restR := RestReservation{
			//Duid:         v.Duid,
			BootFileName: v.BootFileName,
			Hostname:     v.Hostname,
		}
		restR.ID = strconv.Itoa(int(v.ID))
		restRs = append(restRs, &restR)
	}

	return restRs
}

//tools func
func ConvertPoolsFromOrmToRest(ps []*dhcporm.Pool) []*RestPool {
	log.Println("into ConvertPoolsFromOrmToRest")

	var restPs []*RestPool
	for _, v := range ps {
		restP := RestPool{
			BeginAddress: v.BeginAddress,
			EndAddress:   v.EndAddress,
		}
		restP.ID = strconv.Itoa(int(v.ID))
		restPs = append(restPs, &restP)
	}

	return restPs
}

func (s *Dhcpv4) ConvertSubnetv4FromOrmToRest(v *dhcporm.OrmSubnetv4) *RestSubnetv4 {

	v4 := &RestSubnetv4{}
	v4.SetID(strconv.Itoa(int(v.ID)))
	v4.Subnet = v.Subnet
	v4.Name = v.Name
	v4.SubnetId = strconv.Itoa(int(v.ID))
	v4.ValidLifetime = v.ValidLifetime
	v4.Reservations = ConvertReservationsFromOrmToRest(v.Reservations)
	v4.CreationTimestamp = resource.ISOTime(v.CreatedAt)
	return v4
}
func (r *ReservationHandler) convertSubnetv4ReservationFromOrmToRest(v *dhcporm.Reservation) *RestReservation {
	rsv := &RestReservation{}

	if v == nil {
		return rsv
	}

	rsv.SetID(strconv.Itoa(int(v.ID)))
	rsv.BootFileName = v.BootFileName
	//rsv.Duid = v.Duid

	return rsv
}
func (r *PoolHandler) convertSubnetv4PoolFromOrmToRest(v *dhcporm.Pool) *RestPool {
	pool := &RestPool{}

	if v == nil {
		return pool
	}

	pool.SetID(strconv.Itoa(int(v.ID)))
	pool.BeginAddress = v.BeginAddress

	return pool
}
func (n RestReservation) GetParents() []resource.ResourceKind {
	log.Println("dhcprest, into RestReservation GetParents")
	return []resource.ResourceKind{RestSubnetv4{}}
}
func (n RestPool) GetParents() []resource.ResourceKind {
	log.Println("dhcprest, into RestPool GetParents")
	return []resource.ResourceKind{RestSubnetv4{}}
}

//func (s *Dhcpv4) convertSubnetv4ReservationFromOrmToRest(v *dhcporm.Reservation) *RestReservation {
//
//	rr := &RestReservation{}
//	rr.SetID(strconv.Itoa(int(v.ID)))
//	rr.BootFileName = v.BootFileName
//	rr.Duid = v.Duid
//
//	return rr
//}
func ConvertStringToUint(s string) uint {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Errorf("convert string to uint error, s: %s", s)
		return 0
	}

	return uint(i)
}
func ConvertStringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Errorf("convert string to int error, s: %s", s)
		return 0
	}

	return i
}

func (r *ReservationHandler) GetReservations(subnetId string) []*RestReservation {
	list := PGDBConn.OrmReservationList(subnetId)
	rsv := ConvertReservationsFromOrmToRest(list)

	return rsv
}
func (r *ReservationHandler) GetSubnetv4Reservation(subnetId string, rsv_id string) *RestReservation {
	orm := PGDBConn.OrmGetReservation(subnetId, rsv_id)
	rsv := r.convertSubnetv4ReservationFromOrmToRest(orm)

	return rsv
}

func (r *PoolHandler) GetPools(subnetId string) []*RestPool {
	list := PGDBConn.OrmPoolList(subnetId)
	pool := ConvertPoolsFromOrmToRest(list)

	return pool
}
func (r *PoolHandler) GetSubnetv4Pool(subnetId string, pool_id string) *RestPool {
	orm := PGDBConn.OrmGetPool(subnetId, pool_id)
	pool := r.convertSubnetv4PoolFromOrmToRest(orm)

	return pool
}

func Ip2long(ipstr string) (ip uint32) {
	r := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})`
	reg, err := regexp.Compile(r)
	if err != nil {
		return 0
	}
	ips := reg.FindStringSubmatch(ipstr)
	if ips == nil {
		return 0
	}

	ip1, _ := strconv.Atoi(ips[1])
	ip2, _ := strconv.Atoi(ips[2])
	ip3, _ := strconv.Atoi(ips[3])
	ip4, _ := strconv.Atoi(ips[4])

	if ip1 > 255 || ip2 > 255 || ip3 > 255 || ip4 > 255 {
		return 0
	}

	ip += uint32(ip1 * 0x1000000)
	ip += uint32(ip2 * 0x10000)
	ip += uint32(ip3 * 0x100)
	ip += uint32(ip4)

	return ip
}
func Long2ip(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip>>24, ip<<8>>24, ip<<16>>24, ip<<24>>24)
}
func getCidrIpRange(cidr string) (string, string) {
	ip := strings.Split(cidr, "/")[0]
	ipSegs := strings.Split(ip, ".")
	//fmt.Println("ipsegs: ", ipSegs)

	maskLen, _ := strconv.Atoi(strings.Split(cidr, "/")[1])

	seg1MinIp, seg1MaxIp := getIpSeg1Range(ipSegs, maskLen)
	seg2MinIp, seg2MaxIp := getIpSeg2Range(ipSegs, maskLen)
	seg3MinIp, seg3MaxIp := getIpSeg3Range(ipSegs, maskLen)
	seg4MinIp, seg4MaxIp := getIpSeg4Range(ipSegs, maskLen)

	//ipPrefix := ipSegs[0] + "." + ipSegs[1] + "."
	ipPrefixMin := strconv.Itoa(seg1MinIp) + "." + strconv.Itoa(seg2MinIp) + "."
	ipPrefixMax := strconv.Itoa(seg1MaxIp) + "." + strconv.Itoa(seg2MaxIp) + "."

	return ipPrefixMin + strconv.Itoa(seg3MinIp) + "." + strconv.Itoa(seg4MinIp),
		ipPrefixMax + strconv.Itoa(seg3MaxIp) + "." + strconv.Itoa(seg4MaxIp)
}

//计算得到CIDR地址范围内可拥有的主机数量
func getCidrHostNum(maskLen int) uint {
	cidrIpNum := uint(0)
	var i uint = uint(32 - maskLen - 1)
	for ; i >= 1; i-- {
		cidrIpNum += 1 << i
	}
	return cidrIpNum
}

//获取Cidr的掩码
func getCidrIpMask(maskLen int) string {
	fmt.Println("into getCidrIpMask, maskLen: ", maskLen)

	// ^uint32(0)二进制为32个比特1，通过向左位移，得到CIDR掩码的二进制
	cidrMask := ^uint32(0) << uint(32-maskLen)
	//fmt.Println(fmt.Sprintf("%b \n", cidrMask))
	//计算CIDR掩码的四个片段，将想要得到的片段移动到内存最低8位后，将其强转为8位整型，从而得到
	cidrMaskSeg1 := uint8(cidrMask >> 24)
	cidrMaskSeg2 := uint8(cidrMask >> 16)
	cidrMaskSeg3 := uint8(cidrMask >> 8)
	cidrMaskSeg4 := uint8(cidrMask & uint32(255))

	mask := fmt.Sprint(cidrMaskSeg1) + "." + fmt.Sprint(cidrMaskSeg2) + "." + fmt.Sprint(cidrMaskSeg3) + "." + fmt.Sprint(cidrMaskSeg4)
	//fmt.Println("in getCidrIpMask, mask: ", mask)

	return mask
}

//得到第1段IP的区间（第一片段.第二片段.第三片段.第四片段）
func getIpSeg1Range(ipSegs []string, maskLen int) (int, int) {

	if maskLen > 8 {
		segIp, _ := strconv.Atoi(ipSegs[0])
		return segIp, segIp
	}
	ipSeg, _ := strconv.Atoi(ipSegs[0])
	return getIpSegRange(uint8(ipSeg), uint8(8-maskLen))
}

//得到第2段IP的区间（第一片段.第二片段.第三片段.第四片段）
func getIpSeg2Range(ipSegs []string, maskLen int) (int, int) {
	if maskLen > 16 {
		segIp, _ := strconv.Atoi(ipSegs[1])
		return segIp, segIp
	}
	ipSeg, _ := strconv.Atoi(ipSegs[1])
	return getIpSegRange(uint8(ipSeg), uint8(16-maskLen))
}

//得到第三段IP的区间（第一片段.第二片段.第三片段.第四片段）
func getIpSeg3Range(ipSegs []string, maskLen int) (int, int) {
	if maskLen > 24 {
		segIp, _ := strconv.Atoi(ipSegs[2])
		return segIp, segIp
	}
	ipSeg, _ := strconv.Atoi(ipSegs[2])
	return getIpSegRange(uint8(ipSeg), uint8(24-maskLen))
}

//得到第四段IP的区间（第一片段.第二片段.第三片段.第四片段）
func getIpSeg4Range(ipSegs []string, maskLen int) (int, int) {
	ipSeg, _ := strconv.Atoi(ipSegs[3])
	segMinIp, segMaxIp := getIpSegRange(uint8(ipSeg), uint8(32-maskLen))
	return segMinIp + 1, segMaxIp
}

//根据用户输入的基础IP地址和CIDR掩码计算一个IP片段的区间
func getIpSegRange(userSegIp, offset uint8) (int, int) {
	var ipSegMax uint8 = 255
	netSegIp := ipSegMax << offset
	segMinIp := netSegIp & userSegIp
	segMaxIp := userSegIp&(255<<offset) | ^(255 << offset)
	return int(segMinIp), int(segMaxIp)
}

func getSegs(cidr string, newMask int) []string {
	log.Println("into getSegs, cidr: ", cidr)

	minIp, _ := getCidrIpRange(cidr)

	curMask, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
	//curTotal := math.Exp2(float64(32 - curMask))
	//fmt.Println("maskLen: ", curMask, ", curTotal: ", curTotal)
	//newMask := 25
	total := int(math.Exp2(float64(32 - newMask)))
	//fmt.Println("new total: ", total)

	number := int(math.Exp2(float64(newMask - curMask)))
	//fmt.Println("number: ", number)

	var retStr []string
	for i := 0; i < number; i = i + 1 {
		currIP := int(Ip2long(minIp) + uint32(i*total) - 1)
		fmt.Println("currIP: ", Long2ip(uint32(currIP)))

		currIPStr := Long2ip(uint32(currIP))

		retStr = append(retStr, currIPStr+"/"+strconv.Itoa(newMask))
	}
	return retStr
}

func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

//get
func GetMergedSubnetv4Name(str string) (string, error) {
	log.Println("GetMergedSubnetv4Name str:", str)

	b := []byte(str)
	newCidr, e := ipv4.MergeIPNets(pl.ParseIPv4AndCIDR(string(b)))
	if e == nil {
		fmt.Println("merged: ", newCidr)
		//for ip := range merged {
		//fmt.Println(merged[ip])
		//}
	} else {
		fmt.Println("error: ", e)
	}
	if len(newCidr) != 1 {
		errStr := "错误, 无法合并子网"
		log.Println(errStr)
		return "", fmt.Errorf(errStr)
	}
	//log.Println("=== merged: ", newCidr)

	newIP := newCidr[0].IP.String()
	newMaskSize, _ := newCidr[0].Mask.Size()
	retStr := newIP + "/" + strconv.Itoa(newMaskSize)

	//log.Println("retStr: ", retStr)
	return retStr, nil
}
