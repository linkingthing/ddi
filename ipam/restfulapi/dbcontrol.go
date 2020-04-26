package restfulapi

import (
	"fmt"
	"github.com/jinzhu/gorm"
	dhcptb "github.com/linkingthing/ddi/dhcp/dhcporm"
	"github.com/linkingthing/ddi/ipam"
	tb "github.com/linkingthing/ddi/ipam/dbtables"
	"github.com/paulstuart/ping"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

var DBCon *DBController

type DBController struct {
	db *gorm.DB
}

func NewDBController(db *gorm.DB) *DBController {
	one := &DBController{}
	one.db = db
	tx := one.db.Begin()
	defer tx.Rollback()
	if err := tx.AutoMigrate(&tb.Ipv6PlanedAddrTree{}).Error; err != nil {
		panic(err)
	}
	if err := tx.AutoMigrate(&tb.BitsUseFor{}).Error; err != nil {
		panic(err)
	}
	tx.Commit()
	return one
}
func (controller *DBController) Close() {
	controller.db.Close()
}

func (controller *DBController) DetectAliveAddress() error {
	//get all the resevation address where reserv_type equal "hw-address" or "client-id"
	var reservs []dhcptb.OrmReservation
	if err := controller.db.Find(&reservs).Error; err != nil {
		return err
	}
	type stable struct {
		IP         string
		Subnetv4ID uint
	}
	var stables []stable
	for _, r := range reservs {
		if r.ReservType == "hw-address" || r.ReservType == "client-id" {
			tmp := stable{IP: r.IpAddress, Subnetv4ID: r.Subnetv4ID}
			stables = append(stables, tmp)
		}
	}
	//var alives []alive
	var alives []tb.AliveAddress
	for _, s := range stables {
		if ping.Ping(s.IP, 2) {
			tmp := tb.AliveAddress{ScanTime: time.Now().Unix(), LastAliveTime: time.Now().Unix(), IPAddress: s.IP, Subnetv4ID: s.Subnetv4ID}
			alives = append(alives, tmp)
		} else {
			tmp := tb.AliveAddress{ScanTime: time.Now().Unix(), LastAliveTime: 0, IPAddress: s.IP, Subnetv4ID: s.Subnetv4ID}
			alives = append(alives, tmp)
		}
	}
	tx := controller.db.Begin()
	defer tx.Rollback()
	for _, a := range alives {
		if a.LastAliveTime == 0 {
			tmp := tb.AliveAddress{IPAddress: a.IPAddress}
			if err := tx.First(&tmp).Error; err != nil {
				if err := tx.Save(&a).Error; err != nil {
					return err
				}
			} else {
				a.LastAliveTime = tmp.LastAliveTime
				if err := tx.Save(&a).Error; err != nil {
					return err
				}
			}
		} else {
			if err := tx.Save(&a).Error; err != nil {
				return err
			}
		}
	}
	tx.Commit()
	return nil
}

func (controller *DBController) CreateSubtree(data *ipam.Subtree) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := controller.CreateSubtreeRecursive(data, 0, tx, 0, 0); err != nil {
		return err
	}

	tx.Commit()
	return nil
}
func (controller *DBController) CreateSubtreeRecursive(data *ipam.Subtree, parentid uint, tx *gorm.DB, depth int, maxCode int) error {
	//general subnet
	var max byte
	for _, v := range data.Nodes {
		if max < v.EndNodeCode {
			max = v.EndNodeCode
		}
	}
	if data.SubtreeBitNum == 0 || (data.SubtreeBitNum > 0 && int(max)+1 > int(math.Pow(2, float64(data.SubtreeBitNum)))) {
		var newMaxCode int
		if int(max)+1 > len(data.Nodes)*2 {
			newMaxCode = int(max) + 1
		} else {
			newMaxCode = len(data.Nodes) * 2
		}
		var f float64
		f = math.Log2(float64(newMaxCode))
		if int(f*10)%10 > 0 {
			data.SubtreeBitNum = byte(f) + 1
		} else {
			data.SubtreeBitNum = byte(f)
		}
	}
	controller.CaculateSubnet(data)
	fmt.Println("CreateSubtreeRecursive:", data)
	//add data to table Ipv6PlanedAddrTree
	one := tb.Ipv6PlanedAddrTree{}
	one.Depth = depth
	one.Name = data.Name
	one.ParentID = parentid
	one.BeginSubnet = data.BeginSubnet
	one.EndSubnet = data.EndSubnet
	one.BeginNodeCode = int(data.BeginNodeCode)
	one.EndNodeCode = int(data.EndNodeCode)
	one.MaxCode = maxCode
	if len(data.Nodes) == 0 {
		one.IsLeaf = true
	} else {
		one.IsLeaf = false
	}
	/*if data.ID != "" {
		var num int
		var err error
		if num, err = strconv.Atoi(data.ID); err != nil {
			return err
		}
		one.ID = uint(num)
		if err := tx.Save(&one).Error; err != nil {
			return err
		}
	} else {*/
	if err := tx.Create(&one).Error; err != nil {
		return err
	}
	//}
	data.ID = strconv.Itoa(int(one.ID))
	data.Depth = depth
	//add data to table BitsUseFor
	if data.Nodes != nil {
		bitsUsedFor := tb.BitsUseFor{}
		bitsUsedFor.Parentid = one.ID
		bitsUsedFor.UsedFor = data.SubtreeUseDFor
		/*if data.ID != "" {
			if err := tx.Save(&bitsUsedFor).Error; err != nil {
				return err
			}
		} else {*/
		if err := tx.Create(&bitsUsedFor).Error; err != nil {
			return err
		}
		//}
	}
	for i, _ := range data.Nodes {
		controller.CreateSubtreeRecursive(&data.Nodes[i], one.ID, tx, depth+1, int(math.Pow(2, float64(data.SubtreeBitNum))))
	}
	return nil
}

func (controller *DBController) PrefixIncrementN(beginIpv6 string, prefixLength int, n int) (string, error) {
	var ipv6Addr net.IP
	ipv6Addr = net.ParseIP(beginIpv6)
	if ipv6Addr == nil {
		return "", fmt.Errorf("ipv6 adderss %v parse err!", beginIpv6)
	}
	var frontDest int64
	//var frontSour int64
	for i := 0; i < 8; i++ {
		frontDest += int64(ipv6Addr[i]) * int64(math.Pow(2, float64((7-i)*8)))
	}
	frontDest += int64(n) * int64(math.Pow(2, float64(64-prefixLength)))
	for i := 0; i < 8; i++ {
		fmt.Println(byte(frontDest >> ((7 - i) * 8) & 0x000000FF))
		ipv6Addr[i] = byte(frontDest >> ((7 - i) * 8) & 0x000000FF)
	}
	return ipv6Addr.String(), nil
}

func (controller *DBController) CaculateSubnet(p *ipam.Subtree) error {
	//caculate the same prefix ipv6 address between BeginSubnet and EndSubnet's ipv6 adderss.
	var sameIpv6 string
	var prefixLength int
	if p.BeginNodeCode != p.EndNodeCode {
		begin := strings.Split(p.BeginSubnet, "/")
		if len(begin) != 2 {
			return fmt.Errorf("subnet id:", p.ID, "subnet format error!")
		}
		var err error
		if prefixLength, err = strconv.Atoi(begin[1]); err != nil {
			return err
		}
		prefixLength -= int(math.Log2(float64(p.EndNodeCode - p.BeginNodeCode + 1)))
		var beginIpv6 net.IP
		beginIpv6 = net.ParseIP(begin[0])
		if beginIpv6 == nil {
			return fmt.Errorf("ip parse error: %v", begin[0])
		}
		end := strings.Split(p.EndSubnet, "/")
		if len(end) != 2 {
			return fmt.Errorf("subnet id:", p.ID, "subnet format error!")
		}
		var endIpv6 net.IP
		endIpv6 = net.ParseIP(end[0])
		if endIpv6 == nil {
			return fmt.Errorf("ip parse error: %v", end[0])
		}

		var samePrefix net.IP
		samePrefix = net.ParseIP("::")
		if samePrefix == nil {
			return fmt.Errorf("ip parse error: %v", samePrefix)
		}
		for i := 0; i < 16; i++ {
			samePrefix[i] = beginIpv6[i] & endIpv6[i]
		}
		sameIpv6 = samePrefix.String()
	} else {
		s := strings.Split(p.BeginSubnet, "/")
		if len(s) != 2 {
			return fmt.Errorf("subnet id:", p.ID, "subnet format error!")
		}
		var err error
		if prefixLength, err = strconv.Atoi(s[1]); err != nil {
			return err
		}
		sameIpv6 = s[0]
	}
	for i, n := range p.Nodes {
		var beginIpv6 string
		var endIpv6 string
		var err error
		if beginIpv6, err = controller.PrefixIncrementN(sameIpv6, prefixLength+int(p.SubtreeBitNum), int(n.BeginNodeCode)); err != nil {
			return err
		}
		if endIpv6, err = controller.PrefixIncrementN(sameIpv6, prefixLength+int(p.SubtreeBitNum), int(n.EndNodeCode)); err != nil {
			return err
		}
		p.Nodes[i].BeginSubnet = beginIpv6 + "/" + strconv.Itoa(int(prefixLength+int(p.SubtreeBitNum)))
		p.Nodes[i].EndSubnet = endIpv6 + "/" + strconv.Itoa(int(prefixLength+int(p.SubtreeBitNum)))
	}
	return nil
}

func (controller *DBController) DeleteSubtree(id string) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	one := tb.Ipv6PlanedAddrTree{}
	if err := tx.First(&one, id).Error; err != nil {
		return err
	}
	if !one.IsLeaf {
		var childs []tb.Ipv6PlanedAddrTree
		if err := tx.Where("parent_id = ?", id).Find(&childs).Error; err != nil {
			return err
		}
		for _, c := range childs {
			if err := controller.DeleteOne(strconv.Itoa(int(c.ID)), tx); err != nil {
				return err
			}
		}
	}
	if err := tx.Unscoped().Where("parentid = ?", one.ID).Delete(&tb.BitsUseFor{}).Error; err != nil {
		return err
	}
	//update the parent's IsLeaf to be true if it's exists.
	if one.ParentID != 0 {
		parent := tb.Ipv6PlanedAddrTree{}
		parent.ID = one.ParentID
		if err := tx.Model(&parent).UpdateColumn("is_leaf", "true").Error; err != nil {
			return err
		}
	}
	if err := tx.Unscoped().Delete(&one).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}
func (controller *DBController) DeleteOne(id string, tx *gorm.DB) error {
	one := tb.Ipv6PlanedAddrTree{}
	if err := tx.First(&one, id).Error; err != nil {
		return err
	}
	if !one.IsLeaf {
		var childs []tb.Ipv6PlanedAddrTree
		if err := tx.Where("parent_id = ?", id).Find(&childs).Error; err != nil {
			return err
		}
		for _, c := range childs {
			if err := controller.DeleteOne(strconv.Itoa(int(c.ID)), tx); err != nil {
				return err
			}
		}
	}
	if one.ParentID != 0 {
		if err := tx.Unscoped().Where("parentid = ?", one.ParentID).Delete(&tb.BitsUseFor{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Unscoped().Delete(&one).Error; err != nil {
		return err
	}
	return nil
}

func (controller *DBController) GetSubtree(id string) (*ipam.Subtree, error) {
	data := ipam.Subtree{}
	one := tb.Ipv6PlanedAddrTree{}
	var many []tb.Ipv6PlanedAddrTree
	//var one tb.Ipv6PlanedAddrTree
	if id == "" {
		if err := controller.db.Where("parent_id = ?", 0).Find(&many).Error; err != nil {
			return nil, nil
		}
		fmt.Println(many)
		if len(many) >= 1 {
			id = strconv.Itoa(int(many[0].ID))
			one = many[0]
		} else {
			return nil, nil
		}
	} else {
		if err := controller.db.First(&one, id).Error; err != nil {
			if err := controller.db.Where("parent_id = ?", 0).Find(&one).Error; err != nil {
				return nil, err
			}
		}
	}
	data.ID = strconv.Itoa(int(one.ID))
	data.Name = one.Name
	data.BeginSubnet = one.BeginSubnet
	data.EndSubnet = one.EndSubnet
	data.BeginNodeCode = byte(one.BeginNodeCode)
	data.EndNodeCode = byte(one.EndNodeCode)
	data.SubtreeBitNum = 0
	data.Depth = one.Depth
	var usedFors []tb.BitsUseFor
	if err := controller.db.Where("parentid = ?", one.ID).Find(&usedFors).Error; err != nil {
		return nil, err
	}
	if len(usedFors) >= 1 {
		data.SubtreeUseDFor = usedFors[0].UsedFor
	}
	var bitNum int
	var err error
	if !one.IsLeaf {
		if bitNum, err = controller.GetNextTree(&data.Nodes, one.ID); err != nil {
			return nil, err
		}
	}
	data.SubtreeBitNum = byte(bitNum)
	return &data, nil
}

func (controller *DBController) GetNextTree(p *[]ipam.Subtree, parentid uint) (int, error) {
	var many []tb.Ipv6PlanedAddrTree
	if err := controller.db.Where("parent_id = ?", parentid).Find(&many).Error; err != nil {
		return 0, err
	}
	for _, one := range many {
		data := ipam.Subtree{}
		data.ID = strconv.Itoa(int(one.ID))
		data.Name = one.Name
		data.BeginSubnet = one.BeginSubnet
		data.EndSubnet = one.EndSubnet
		data.BeginNodeCode = byte(one.BeginNodeCode)
		data.EndNodeCode = byte(one.EndNodeCode)
		data.SubtreeBitNum = 0
		data.Depth = one.Depth
		var usedFors []tb.BitsUseFor
		if err := controller.db.Where("parentid = ?", one.ID).Find(&usedFors).Error; err != nil {
			return 0, err
		}
		if len(usedFors) == 1 {
			data.SubtreeUseDFor = usedFors[0].UsedFor
		}
		var bitNum int
		var err error
		if !one.IsLeaf {
			if bitNum, err = controller.GetNextTree(&data.Nodes, one.ID); err != nil {
				return 0, err
			}
		}
		data.SubtreeBitNum = byte(bitNum)
		*p = append(*p, data)
	}
	if len(many) > 0 {
		return int(math.Log2(float64(many[0].MaxCode))), nil
	} else {
		return 0, nil
	}
}

func (controller *DBController) SplitSubnet(p *ipam.SplitSubnet) (*ipam.SplitSubnetResult, error) {
	data := ipam.SplitSubnetResult{}
	return &data, nil
}
