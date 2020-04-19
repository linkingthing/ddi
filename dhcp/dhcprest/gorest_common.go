package dhcprest

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	"github.com/zdnscloud/gorest/resource"
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
	OptionNameKind  = resource.DefaultKindName(RestOptionName{})

	//DividedAddressKind = resource.DefaultKindName(res.DividedAddress{})
	//OptionNameConfigKind = resource.DefaultKindName(RestOptionNameConfig{})

	db *gorm.DB
)

//type Dhcpv4Serv struct {
//	resource.ResourceBase `json:",inline"`
//	ConfigJson            string `json:"configJson" rest:"required=true,minLen=1,maxLen=1000000"`
//}

// added for option list v4 or v6
type RestOptionName struct {
	resource.ResourceBase `json:"embedded,inline"`
	OptionVer             string `json:"optionVer"` // v4 or v6
	OptionId              int    `json:"optionId"`
	OptionName            string `json:"optionName"`
	OptionType            string `json:"optionType"`
}

type OptionNameData struct {
	resource.ResourceBase `json:",inline"`
	Oper                  string `json:"oper" rest:"required=true,minLen=1,maxLen=20"`
}

func (r RestOptionName) CreateAction(name string) *resource.Action {
	log.Println("into RestOptionName, create action")
	switch name {
	case "list":
		return &resource.Action{
			Name:  "list",
			Input: &OptionNameData{},
		}
	default:
		return nil
	}
}

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
	ClientId       string       `json:"clientId"` //reservations can be multi-types, need to split  todo
	CircuitId      string       `json:"circuitId"`
	Duid           string       `json:"duid"`
	Hostname       string       `json:"hostname"`
	IpAddress      string       `json:"ipAddress"`
	HwAddress      string       `json:"hwAddress"`
	NextServer     string       `json:"nextServer"`
	OptionData     []RestOption `json:"optionData"`
	ServerHostname string       `json:"serverHostname"`
	ResvType       string       `json:"resvType"` // resv or stable
}

type RestPool struct {
	resource.ResourceBase `json:"embedded,inline"`
	Subnetv4Id            string       `json:"subnetv4Id"`
	OptionData            []RestOption `json:"optionData"`
	BeginAddress          string       `json:"beginAddress,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	EndAddress            string       `json:"endAddress,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	MaxValidLifetime      string       `json:"maxValidLifetime,omitempty"`
	ValidLifetime         string       `json:"validLifetime,omitempty"`
	Total                 uint32       `json:"total"`
	Usage                 float32      `json:"usage"`
	AddressType           string       `json:"addressType"`
	PoolName              string       `json:"poolName"`
	Gateway               string       `json:"gateway"`
	DnsServer             string       `json:"dnsServer"`
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

type RestSubnetv46 struct {
	resource.ResourceBase `json:"embedded,inline"`
	Type                  string `json:"type"` // v4 or v6

	Name             string `json:"name,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	Subnet           string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	SubnetId         string `json:"subnet_id"`
	ValidLifetime    string `json:"validLifetime"`
	MaxValidLifetime string `json:"maxValidLifetime"`
	Reservations     []*RestReservation
	Pools            []*RestPool
	SubnetTotal      string `json:"total"`
	SubnetUsage      string `json:"usage"`
	Gateway          string `json:"gateway"`
	DnsServer        string `json:"dnsServer"`
	//added for new zone handler
	DhcpEnable int    `json:"dhcpEnable"`
	DnsEnable  int    `json:"dnsEnable"`
	ZoneName   string `json:"zoneName"`
	ViewId     string `json:"viewId"`
	Notes      string `json:"notes"`
}

type RestSubnetv4 struct {
	resource.ResourceBase `json:"embedded,inline"`
	Name                  string `json:"name,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	Subnet                string `json:"subnet,omitempty" rest:"required=true,minLen=1,maxLen=255"`
	SubnetId              string `json:"subnet_id"`
	ValidLifetime         string `json:"validLifetime"`
	MaxValidLifetime      string `json:"maxValidLifetime"`
	Reservations          []*RestReservation
	Pools                 []*RestPool
	SubnetTotal           string `json:"total"`
	SubnetUsage           string `json:"usage"`
	Gateway               string `json:"gateway"`
	DnsServer             string `json:"dnsServer"`
	//added for new zone handler
	DhcpEnable int    `json:"dhcpEnable"`
	DnsEnable  int    `json:"dnsEnable"`
	ZoneName   string `json:"zoneName"`
	ViewId     string `json:"viewId"`
	Notes      string `json:"notes"`
}

func (s4 RestSubnetv4) GetActions() []resource.Action {
	log.Println("into RestSubnetv4, GetActions")
	var actions []resource.Action
	action := resource.Action{
		Name:   "mergesplit",
		Input:  &MergeSplitData{},
		Output: &MergeSplitData{},
	}
	actions = append(actions, action)

	//log.Println("in cluster GetActions, actions: ", actions)
	return actions
}

func (s4 RestSubnetv4) CreateAction(name string) *resource.Action {
	log.Println("into RestSubnetv4, create action")
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
type Subnetv46s struct {
	Subnetv46s []*RestSubnetv46
	db         *gorm.DB
}

func NewSubnetv4s(db *gorm.DB) *Subnetv4s {
	return &Subnetv4s{db: db}
}
func NewSubnetv46s(db *gorm.DB) *Subnetv46s {
	return &Subnetv46s{db: db}
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

type OptionNameStatistics struct {
	OptionVer string `json:"optionVer"`
	Total     int    `json:"total"`
}
type OptionNameStatisticsRet struct {
	V4Num int `json:"v4Num"`
	V6Num int `json:"v6Num"`
}
type OptionNameConfigRet struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Num   int    `json:"num"`
	Notes string `json:"notes"`
}
type OptionNameConfigsRet struct {
	resource.ResourceBase `json:"embedded,inline"`
	V4                    RestOptionNameConfig `json:"v4"`
	V6                    RestOptionNameConfig `json:"v4"`
}
type RestOptionNameConfig struct {
	//resource.ResourceBase `json:"embedded,inline"`
	OptionNotes string `json:"optionNotes"` // v4 or v6
	OptionNum   int    `json:"optionNum"`
	OptionName  string `json:"optionName"`
	OptionType  string `json:"optionType"`
}

// added for option list v4 or v6
//type optionNameHandler struct {
//	optionNames *OptionNamesState
//}
//
//type OptionNamesState struct {
//	OptionNames []*RestOptionName
//}
//
//func NewOptionNamesState() *OptionNamesState {
//	log.Println("into NewOptionNamesState")
//	return &OptionNamesState{}
//}
//
func NewOptionNameHandler(s *Subnetv4s) *optionNameHandler {
	return &optionNameHandler{
		subnetv4s: s,
		db:        s.db,
	}
}

type optionNameHandler struct {
	subnetv4s *Subnetv4s
	db        *gorm.DB
	lock      sync.Mutex
}

type OptionHandler struct {
	subnetv4s *Subnetv4s
	db        *gorm.DB
	lock      sync.Mutex
}

func NewOptionHandler(s *Subnetv4s) *OptionHandler {
	return &OptionHandler{
		subnetv4s: s,
		db:        s.db,
	}
}

//tools func
func ConvertReservationsFromOrmToRest(rs []dhcporm.OrmReservation) []*RestReservation {

	var restRs []*RestReservation
	for _, v := range rs {
		restR := RestReservation{
			Duid:         v.Duid,
			BootFileName: v.BootFileName,
			Hostname:     v.Hostname,
			IpAddress:    v.IpAddress,
		}
		restR.ID = strconv.Itoa(int(v.ID))
		restRs = append(restRs, &restR)
	}

	return restRs
}
func ConvertReservationFromOrmToRest(v *dhcporm.OrmReservation) *RestReservation {

	restR := RestReservation{
		Duid:         v.Duid,
		BootFileName: v.BootFileName,
		Hostname:     v.Hostname,
		IpAddress:    v.IpAddress,
	}
	restR.ID = strconv.Itoa(int(v.ID))

	return &restR
}

func ConvertOptionNamesFromOrmToRest(ps []*dhcporm.OrmOptionName) []*RestOptionName {
	log.Println("into ConvertOptionsFromOrmToRest")

	var restOPs []*RestOptionName
	for _, v := range ps {
		restOP := RestOptionName{
			OptionId:   v.OptionId,
			OptionVer:  v.OptionVer,
			OptionType: v.OptionType,
			OptionName: v.OptionName,
		}
		restOP.ID = strconv.Itoa(int(v.ID))
		restOP.CreationTimestamp = resource.ISOTime(v.CreatedAt)
		restOPs = append(restOPs, &restOP)
	}

	return restOPs
}

func ConvertOptionsFromOrmToRest(ps []*dhcporm.Option) []*RestOption {
	log.Println("into ConvertOptionsFromOrmToRest")

	var restOPs []*RestOption
	for _, v := range ps {
		restOP := RestOption{
			Code: v.Code,
		}
		restOP.ID = strconv.Itoa(int(v.ID))
		restOPs = append(restOPs, &restOP)
	}

	return restOPs
}

//tools func
func (r *PoolHandler) ConvertPoolsFromOrmToRest(ps []*dhcporm.Pool) []*RestPool {
	log.Println("into ConvertPoolsFromOrmToRest")

	var restPs []*RestPool
	for _, v := range ps {
		//restP := RestPool{
		//	BeginAddress: v.BeginAddress,
		//	EndAddress:   v.EndAddress,
		//}
		restP := r.convertSubnetv4PoolFromOrmToRest(v)

		//todo get gateway dnsServer from ormSubnetv4

		restPs = append(restPs, restP)

	}

	return restPs
}

func backtoIP4(ipInt int64) string {

	// need to do two bit shifting and “0xff” masking
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}
func ipv42Long(ip string) uint32 {
	var long uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &long)
	return long
}
func (s *Dhcpv4) ConvertSubnetv4FromOrmToRest(v *dhcporm.OrmSubnetv4) *RestSubnetv4 {

	//log.Println("---into ConvertSubnetv4FromOrmToRest")
	v4 := &RestSubnetv4{}
	v4.SetID(strconv.Itoa(int(v.ID)))
	v4.Subnet = v.Subnet
	v4.Name = v.Name
	v4.SubnetId = strconv.Itoa(int(v.ID))
	v4.ValidLifetime = v.ValidLifetime
	v4.Reservations = ConvertReservationsFromOrmToRest(v.Reservations)
	v4.CreationTimestamp = resource.ISOTime(v.CreatedAt)

	v4.Gateway = v.Gateway
	v4.DnsServer = v.DnsServer
	v4.DhcpEnable = v.DhcpEnable
	v4.DnsEnable = v.DnsEnable
	v4.ViewId = v.ViewId
	v4.Notes = v.Notes

	if len(v4.ZoneName) == 0 {
		v4.ZoneName = v4.Name
	}

	return v4
}
func (r *ReservationHandler) convertSubnetv4ReservationFromOrmToRest(v *dhcporm.OrmReservation) *RestReservation {
	rsv := &RestReservation{}

	if v == nil {
		return rsv
	}

	rsv.SetID(strconv.Itoa(int(v.ID)))
	rsv.BootFileName = v.BootFileName
	rsv.Duid = v.Duid
	rsv.IpAddress = v.IpAddress

	return rsv
}
func (r *PoolHandler) convertSubnetv4PoolFromOrmToRest(v *dhcporm.Pool) *RestPool {
	if v == nil {
		log.Println("into convertSubnetv4PoolFromOrmToRest, v is null, error")
		return nil
	}
	log.Println("into convertSubnetv4PoolFromOrmToRest, v.beginAddress: ", v.BeginAddress)
	pool := &RestPool{}

	if v == nil {
		return pool
	}

	pool.SetID(strconv.Itoa(int(v.ID)))
	pool.BeginAddress = v.BeginAddress
	pool.EndAddress = v.EndAddress
	pool.Total = ipv42Long(v.EndAddress) - ipv42Long(v.BeginAddress) + 1
	pool.ID = strconv.Itoa(int(v.ID))

	// todo get usage of a pool, (put it to somewhere)
	pool.Usage = 0
	pool.AddressType = "resv"
	pool.CreationTimestamp = resource.ISOTime(v.CreatedAt)
	pool.PoolName = v.BeginAddress + "-" + v.EndAddress

	//get ormSubnetv4 from subnetv4Id
	pgdb := NewPGDB(r.db)
	subnetv4Id := strconv.Itoa(int(v.Subnetv4ID))

	s4 := pgdb.GetSubnetv4ById(subnetv4Id)
	pool.Gateway = s4.Gateway
	pool.DnsServer = s4.DnsServer
	pool.Subnetv4Id = subnetv4Id
	pool.MaxValidLifetime = strconv.Itoa(v.MaxValidLifetime)
	pool.ValidLifetime = strconv.Itoa(v.ValidLifetime)

	log.Println("into convertSubnetv4PoolFromOrmToRest, v.MaxValidLifetime: ", v.MaxValidLifetime)
	log.Println("into convertSubnetv4PoolFromOrmToRest, v.ValidLifetime: ", v.ValidLifetime)
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

func (r *optionNameHandler) GetOptionNames() []*RestOptionName {
	list := PGDBConn.OrmOptionNameList()
	option := ConvertOptionNamesFromOrmToRest(list)

	return option
}

//func (r *OptionHandler) GetOptions(subnetId string) []*RestOption {
//	list := PGDBConn.OrmOptionList(subnetId)
//	option := ConvertOptionsFromOrmToRest(list)
//
//	return option
//}

func (r *PoolHandler) GetPools(subnetId string) []*RestPool {
	list := PGDBConn.OrmPoolList(subnetId)
	pool := r.ConvertPoolsFromOrmToRest(list)

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
	newCidr, e := MergeIPNets(ParseIPv4AndCIDR(string(b)))
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

// extract from trimv4 module
func ParseIPv4AndCIDR(data string) []*net.IPNet {
	fmt.Println("data: ", data)
	var reIPv4 = regexp.MustCompile(`(((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])+)(\/(3[0-2]|[1-2][0-9]|[0-9]))?`)
	//var reIPv4 = regexp.MustCompile(`(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/(3[0-2]|[1-2][0-9]|[0-9]))`)
	scanner := bufio.NewScanner(strings.NewReader(data))

	addrs := make([]*net.IPNet, 0)
	for scanner.Scan() {
		x := reIPv4.FindString(scanner.Text())
		fmt.Println("in ParseIPv4AndCIDR, x: ", x)
		if !strings.Contains(x, "/") {
			if !strings.Contains(x, ":") {
				x = x + "/32"
			} else {
				x = x + "/128"
			}
		}
		if addr, cidr, e := net.ParseCIDR(x); e == nil {
			//if !ipv4.IsRFC4193(addr) && !ipv4.IsLoopback(addr) && !ipv4.IsBogonIP(addr) {
			if !IsLoopback(addr) && !IsBogonIP(addr) {
				addrs = append(addrs, cidr)
			}
		}
	}
	return addrs
}

type CIDRBlockIPv4 struct {
	first uint32
	last  uint32
}

type CIDRBlockIPv4s []*CIDRBlockIPv4

type IPNets []*net.IPNet

func BCast4(addr uint32, prefix uint) uint32 {
	return addr | ^Netmask(prefix)
}

func SetBit(addr uint32, bit uint, val uint) uint32 {
	if bit < 0 {
		panic("negative bit index")
	}

	if val == 0 {
		return addr & ^(1 << (32 - bit))
	} else if val == 1 {
		return addr | (1 << (32 - bit))
	} else {
		panic("set bit is not 0 or 1")
	}
}

func UI32ToIPv4(addr uint32) net.IP {
	ip := make([]byte, net.IPv4len)
	binary.BigEndian.PutUint32(ip, addr)
	return ip
}

func IPv4ToUI32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

func IPV4Merge(blocks CIDRBlockIPv4s) ([]*net.IPNet, error) {
	sort.Sort(blocks)

	for i := len(blocks) - 1; i > 0; i-- {
		if blocks[i].first <= blocks[i-1].last+1 {
			blocks[i-1].last = blocks[i].last
			if blocks[i].first < blocks[i-1].first {
				blocks[i-1].first = blocks[i].first
			}
			blocks[i] = nil
		}
	}

	var merged []*net.IPNet
	for _, block := range blocks {
		if block == nil {
			continue
		}

		if err := IPv4RangeSplit(0, 0, block.first, block.last, &merged); err != nil {
			return nil, err
		}
	}

	return merged, nil
}

func NewBlockIPv4(ip net.IP, mask net.IPMask) *CIDRBlockIPv4 {
	var block CIDRBlockIPv4
	block.first = IPv4ToUI32(ip)
	prefix, _ := mask.Size()
	block.last = BCast4(block.first, uint(prefix))

	return &block
}

func Netmask(prefix uint) uint32 {
	if prefix == 0 {
		return 0
	}
	return ^uint32((1 << (32 - prefix)) - 1)
}

func (c CIDRBlockIPv4s) Len() int {
	return len(c)
}

func (c CIDRBlockIPv4s) Less(i, j int) bool {
	lhs := c[i]
	rhs := c[j]

	if lhs.last < rhs.last {
		return true
	} else if lhs.last > rhs.last {
		return false
	}

	if lhs.first < rhs.first {
		return true
	} else if lhs.first > rhs.first {
		return false
	}

	return false
}

func (c CIDRBlockIPv4s) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func MergeIPNets(nets []*net.IPNet) ([]*net.IPNet, error) {
	fmt.Println("ip nets: ", nets)
	if nets == nil {
		fmt.Println("no IPs detected in file")
		return nil, nil
	}
	if len(nets) == 0 {
		return make([]*net.IPNet, 0), nil
	}

	var block4s CIDRBlockIPv4s
	for _, net := range nets {
		ip4 := net.IP.To4()
		if ip4 != nil {
			block4s = append(block4s, NewBlockIPv4(ip4, net.Mask))
		} else {
			return nil, errors.New("Not implemented")
		}
	}

	merged, err := IPV4Merge(block4s)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return merged, nil
}

func IPv4RangeSplit(addr uint32, prefix uint, lo, hi uint32, cidrs *[]*net.IPNet) error {
	if prefix > 32 {
		return fmt.Errorf("Invalid mask size: %d", prefix)
	}

	bc := BCast4(addr, prefix)
	if (lo < addr) || (hi > bc) {
		return fmt.Errorf("%d, %d out of range for network %d/%d, broadcast %d", lo, hi, addr, prefix, bc)
	}

	if (lo == addr) && (hi == bc) {
		cidr := net.IPNet{IP: UI32ToIPv4(addr), Mask: net.CIDRMask(int(prefix), 8*net.IPv4len)}
		*cidrs = append(*cidrs, &cidr)
		return nil
	}

	prefix++
	lowerHalf := addr
	upperHalf := SetBit(addr, prefix, 1)
	if hi < upperHalf {
		return IPv4RangeSplit(lowerHalf, prefix, lo, hi, cidrs)
	} else if lo >= upperHalf {
		return IPv4RangeSplit(upperHalf, prefix, lo, hi, cidrs)
	} else {
		err := IPv4RangeSplit(lowerHalf, prefix, lo, BCast4(lowerHalf, prefix), cidrs)
		if err != nil {
			return err
		}
		return IPv4RangeSplit(upperHalf, prefix, upperHalf, hi, cidrs)
	}
}
