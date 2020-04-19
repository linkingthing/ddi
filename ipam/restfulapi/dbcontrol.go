package restfulapi

import (
	"fmt"
	"github.com/jinzhu/gorm"
	//"github.com/linkingthing/ddi/dhcp/grpc"
	dhcptb "github.com/linkingthing/ddi/dhcp/dhcporm"
	"github.com/linkingthing/ddi/dhcp/dhcprest"
	dhcpgrpc "github.com/linkingthing/ddi/dhcp/grpc"
	"github.com/linkingthing/ddi/ipam"
	tb "github.com/linkingthing/ddi/ipam/dbtables"
	//"github.com/linkingthing/ddi/utils/arp"
	"github.com/paulstuart/ping"
	"log"
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
	if err := tx.AutoMigrate(&tb.DividedAddress{}).Error; err != nil {
		panic(err)
	}
	tx.Commit()
	return one
}
func (controller *DBController) Close() {
	controller.db.Close()
}

func (controller *DBController) GetDividedAddresses(subNetID string, ip string, hostName string, mac string) ([]*ipam.DividedAddress, error) {
	log.Println("into dhcptb GetDividedAddress, subNetID ", subNetID)
	//get the reservation address
	reservData := dhcprest.PGDBConn.OrmReservationList(subNetID)
	allData := make(map[string]tb.DividedAddress, 255)
	for _, a := range reservData {
		if a.ReservType == "hw-address" || a.ReservType == "client-id" {
			//get the stable address
			tmp := tb.DividedAddress{IP: a.IpAddress, AddressType: "stable"}
			allData[a.IpAddress] = tmp
		} else {
			tmp := tb.DividedAddress{IP: a.IpAddress, AddressType: "reserved"}
			allData[a.IpAddress] = tmp
		}
	}
	//get the pools under the subnet
	pools := dhcprest.PGDBConn.OrmPoolList(subNetID)
	var dynamicAddress []string
	for _, pool := range pools {
		beginNums := strings.Split(pool.BeginAddress, ".")
		endNums := strings.Split(pool.EndAddress, ".")
		var err error
		var begin int
		var end int
		if begin, err = strconv.Atoi(string(beginNums[3])); err != nil {
			break
		}
		if end, err = strconv.Atoi(string(endNums[3])); err != nil {
			break
		}
		var beginPart string
		beginPart = beginNums[0] + "." + beginNums[1] + "." + beginNums[2] + "."
		for i := begin; i <= end; i++ {
			dynamicAddress = append(dynamicAddress, beginPart+strconv.Itoa(i))
		}
	}
	found := false
	for _, ip := range dynamicAddress {
		found = false
		for _, a := range reservData {
			if ip == a.IpAddress {
				found = true
				break
			}
		}
		if !found {
			tmp := tb.DividedAddress{IP: ip, AddressType: "dynamic"}
			allData[ip] = tmp
		}
	}
	//get manual address
	var manuals []dhcptb.ManualAddress
	if err := controller.db.Where("subnetv4_id = ?", subNetID).Find(&manuals).Error; err != nil {
		return nil, err
	}
	for _, v := range manuals {
		tmp := tb.DividedAddress{IP: v.IpAddress, AddressType: "manual"}
		allData[v.IpAddress] = tmp
	}
	//get the release address for the subnet
	leases := dhcpgrpc.GetLeases(subNetID)
	for _, l := range leases {
		var macAddr string
		for i := 0; i < len(l.HwAddress); i++ {
			tmp := fmt.Sprintf("%d", l.HwAddress[i])
			macAddr += tmp
		}

		tmp := tb.DividedAddress{IP: l.IpAddress, MacAddress: macAddr, AddressType: "lease", LeaseStartTime: l.Expire - int64(l.ValidLifetime), LeaseEndTime: l.Expire}
		allData[l.IpAddress] = tmp
	}
	if len(pools) > 0 {
		beginNums := strings.Split(pools[0].BeginAddress, ".")
		prefix := beginNums[0] + "." + beginNums[1] + "." + beginNums[2] + "."
		for i := 1; i < 256; i++ {
			if allData[prefix+strconv.Itoa(i)].AddressType == "" {
				tmp := tb.DividedAddress{IP: prefix + strconv.Itoa(i), AddressType: "unused"}
				allData[prefix+strconv.Itoa(i)] = tmp
			}
		}
	}
	var err error
	var data []*ipam.DividedAddress
	var filterData []*ipam.DividedAddress
	var input []tb.DividedAddress
	for _, v := range allData {
		input = append(input, v)
	}
	if data, err = controller.UpdateDividedAddresses(input); err != nil {
		return nil, err
	}
	if ip != "" {
		for _, v := range data {
			if v.IP == ip {
				filterData = append(filterData, v)
			}
		}
		data = filterData
		filterData = filterData[0:0]
	}

	if hostName != "" {
		for _, v := range data {
			if v.HostName == hostName {
				if v.IP == ip {
					filterData = append(filterData, v)
				}
			}
		}
		data = filterData
		filterData = filterData[0:0]
	}
	if mac != "" {
		for _, v := range data {
			if v.MacAddress == mac {
				if v.IP == ip {
					filterData = append(filterData, v)
				}
			}
		}
		data = filterData
	}

	return data, nil
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

func (controller *DBController) UpdateDividedAddresses(dividedAddresses []tb.DividedAddress) ([]*ipam.DividedAddress, error) {
	//create new data in the database
	tx := controller.db.Begin()
	defer tx.Rollback()
	for i, one := range dividedAddresses {
		var tmp tb.DividedAddress
		if err := tx.Where("ip = ?", one.IP).Find(&tmp).Error; err != nil {
			if err := tx.Create(&dividedAddresses[i]).Error; err != nil {
				return nil, err
			}
		} else {
			if one.MacAddress != "" {
				if err := tx.Model(&tmp).UpdateColumn("mac_address", one.MacAddress).Error; err != nil {
					return nil, err
				}
			}
			if one.LeaseStartTime != 0 {
				if err := tx.Model(&tmp).UpdateColumn("lease_start_time", one.LeaseStartTime).Error; err != nil {
					return nil, err
				}
			}
			if one.LeaseEndTime != 0 {
				if err := tx.Model(&tmp).UpdateColumn("lease_end_time", one.LeaseEndTime).Error; err != nil {
					return nil, err
				}
			}
			dividedAddresses[i].ID = tmp.ID
			dividedAddresses[i].MacVender = tmp.MacVender
			dividedAddresses[i].AddressType = tmp.AddressType
			dividedAddresses[i].OperSystem = tmp.OperSystem
			dividedAddresses[i].NetBIOSName = tmp.NetBIOSName
			dividedAddresses[i].HostName = tmp.HostName
			dividedAddresses[i].InterfaceID = tmp.InterfaceID
			dividedAddresses[i].FingerPrint = tmp.FingerPrint
			dividedAddresses[i].DeviceTypeFlag = tmp.DeviceTypeFlag
			dividedAddresses[i].DeviceType = tmp.DeviceType
			dividedAddresses[i].BusinessFlag = tmp.BusinessFlag
			dividedAddresses[i].Business = tmp.Business
			dividedAddresses[i].ChargePersonFlag = tmp.ChargePersonFlag
			dividedAddresses[i].ChargePerson = tmp.ChargePerson
			dividedAddresses[i].TelFlag = tmp.TelFlag
			dividedAddresses[i].Tel = tmp.Tel
			dividedAddresses[i].DepartmentFlag = tmp.DepartmentFlag
			dividedAddresses[i].Department = tmp.Department
			dividedAddresses[i].PositionFlag = tmp.PositionFlag
			dividedAddresses[i].Position = tmp.Position

		}
	}
	tx.Commit()
	var data []*ipam.DividedAddress
	for _, v := range dividedAddresses {
		var tmp ipam.DividedAddress
		tmp.SetID(strconv.Itoa(int(v.ID)))
		tmp.IP = v.IP
		tmp.MacAddress = v.MacAddress
		tmp.MacVender = v.MacVender
		tmp.AddressType = v.AddressType
		tmp.OperSystem = v.OperSystem
		tmp.NetBIOSName = v.NetBIOSName
		tmp.HostName = v.HostName
		tmp.InterfaceID = v.InterfaceID
		tmp.FingerPrint = v.FingerPrint
		tmp.LeaseStartTime = v.LeaseStartTime
		tmp.LeaseEndTime = v.LeaseEndTime
		tmp.DeviceTypeFlag = v.DeviceTypeFlag
		tmp.DeviceType = v.DeviceType
		tmp.BusinessFlag = v.BusinessFlag
		tmp.Business = v.Business
		tmp.ChargePersonFlag = v.ChargePersonFlag
		tmp.ChargePerson = v.ChargePerson
		tmp.TelFlag = v.TelFlag
		tmp.Tel = v.Tel
		tmp.DepartmentFlag = v.DepartmentFlag
		tmp.Department = v.Department
		tmp.PositionFlag = v.PositionFlag
		tmp.Position = v.Position
		data = append(data, &tmp)
	}
	return data, nil
}

func (controller *DBController) UpdateDividedAddress(one *ipam.DividedAddress) error {
	var tmp tb.DividedAddress
	var err error
	var num int
	if num, err = strconv.Atoi(one.ID); err != nil {
		return err
	}
	tmp.ID = uint(num)
	tmp.IP = one.IP
	tmp.MacAddress = one.MacAddress
	tmp.MacVender = one.MacVender
	tmp.AddressType = one.AddressType
	tmp.OperSystem = one.OperSystem
	tmp.NetBIOSName = one.NetBIOSName
	tmp.HostName = one.HostName
	tmp.InterfaceID = one.InterfaceID
	tmp.FingerPrint = one.FingerPrint
	tmp.LeaseStartTime = one.LeaseStartTime
	tmp.LeaseEndTime = one.LeaseEndTime
	tmp.DeviceTypeFlag = one.DeviceTypeFlag
	tmp.DeviceType = one.DeviceType
	tmp.BusinessFlag = one.BusinessFlag
	tmp.Business = one.Business
	tmp.ChargePersonFlag = one.ChargePersonFlag
	tmp.ChargePerson = one.ChargePerson
	tmp.TelFlag = one.TelFlag
	tmp.Tel = one.Tel
	tmp.DepartmentFlag = one.DepartmentFlag
	tmp.Department = one.Department
	tmp.PositionFlag = one.PositionFlag
	tmp.Position = one.Position
	tx := controller.db.Begin()
	defer tx.Rollback()
	if err := tx.Save(&tmp).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}
func (controller *DBController) UpdateIPAttrAppend(attrAppend *ipam.IPAttrAppend) error {
	tx := controller.db.Begin()
	defer tx.Rollback()
	var one tb.DividedAddress
	if err := tx.First(&one, attrAppend.ID).Error; err != nil {
		return err
	}
	one.DeviceTypeFlag = attrAppend.DeviceTypeFlag
	one.BusinessFlag = attrAppend.BusinessFlag
	one.ChargePersonFlag = attrAppend.ChargePersonFlag
	one.TelFlag = attrAppend.TelFlag
	one.DepartmentFlag = attrAppend.DepartmentFlag
	one.PositionFlag = attrAppend.PositionFlag
	if err := tx.Save(&one).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}
func (controller *DBController) GetIPAttrAppend(id string) (*ipam.IPAttrAppend, error) {
	var tmp tb.DividedAddress
	if err := controller.db.First(&tmp, id).Error; err != nil {
		return nil, err
	}
	var one ipam.IPAttrAppend
	one.SetID(id)
	one.DeviceTypeFlag = tmp.DeviceTypeFlag
	one.BusinessFlag = tmp.BusinessFlag
	one.ChargePersonFlag = tmp.ChargePersonFlag
	one.TelFlag = tmp.TelFlag
	one.DepartmentFlag = tmp.DepartmentFlag
	one.PositionFlag = tmp.PositionFlag
	return &one, nil
}
