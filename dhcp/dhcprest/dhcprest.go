package dhcprest

import (
	"fmt"
	"time"

	"log"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/cmd/websocket/server"
	"github.com/linkingthing/ddi/dhcp/dhcporm"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
)

func (s *Dhcpv4) ConvertV4sToV46s(v *RestSubnetv4) *RestSubnetv46 {

	var v46 RestSubnetv46
	v46.Type = "v4"
	v46.Name = v.Name
	v46.Subnet = v.Subnet
	v46.SubnetId = v.SubnetId
	v46.ValidLifetime = v.ValidLifetime
	v46.Reservations = v.Reservations
	v46.Pools = v.Pools
	v46.SubnetTotal = v.SubnetTotal
	v46.SubnetUsage = v.SubnetUsage
	v46.Gateway = v.Gateway
	v46.DnsServer = v.DnsServer
	v46.DhcpEnable = v.DhcpEnable
	v46.DnsEnable = v.DnsEnable
	v46.ZoneName = v.ZoneName
	v46.ViewId = v.ViewId
	v46.Notes = v.Notes
	v46.CreationTimestamp = v.CreationTimestamp

	return &v46
}

func (s *Dhcpv4) ConvertV6sToV46s(v *RestSubnetv6) *RestSubnetv46 {

	var v46 RestSubnetv46
	v46.Type = "v6"
	v46.Name = v.Name
	v46.Subnet = v.Subnet
	v46.SubnetId = v.SubnetId
	v46.ValidLifetime = v.ValidLifetime
	//v46.Reservations = v.Reservations
	//v46.Pools = v.Pools
	v46.SubnetTotal = v.SubnetTotal
	v46.SubnetUsage = v.SubnetUsage
	//v46.Gateway = v.Gateway
	v46.DnsServer = v.DnsServer
	v46.DhcpEnable = v.DhcpEnable
	v46.DnsEnable = v.DnsEnable
	v46.ZoneName = v.ZoneName
	v46.ViewId = v.ViewId
	v46.Notes = v.Notes

	v46.CreationTimestamp = v.CreationTimestamp
	return &v46
}

type SubnetSearch struct {
	DhcpVer string // v4 or v6
	Subnet  string
}

func (s *Dhcpv4) GetSubnetv46s(search *SubnetSearch) []*RestSubnetv46 {
	log.Println("into GetSubnetv46s()")
	s.lock.Lock()
	defer s.lock.Unlock()

	var all []*RestSubnetv46
	if search.DhcpVer != "" && search.DhcpVer == "v4" {
		log.Println("in GetSubnetv46s, v4 search.Subnet: ", search.Subnet)
		v4s := s.GetSubnetv4s(search)
		for _, v4 := range v4s {
			all = append(all, s.ConvertV4sToV46s(v4))
		}
		return all
	}
	if search.DhcpVer != "" && search.DhcpVer == "v6" {
		log.Println("in GetSubnetv46s, v6 search.Subnet: ", search.Subnet)
		dhcpv6 := NewDhcpv6(db)
		v6s := dhcpv6.GetSubnetv6s(search)
		for _, v6 := range v6s {
			//log.Println("v6: ", v6)
			all = append(all, s.ConvertV6sToV46s(v6))
		}
		return all
	}

	v4s := s.GetSubnetv4s(search)
	for _, v4 := range v4s {
		log.Println("v4: ", v4)
		all = append(all, s.ConvertV4sToV46s(v4))
	}
	dhcpv6 := NewDhcpv6(db)
	v6s := dhcpv6.GetSubnetv6s(search)
	for _, v6 := range v6s {
		log.Println("v6: ", v6)
		all = append(all, s.ConvertV6sToV46s(v6))
	}
	log.Println("in GetSubnetv46s, search is nil, all: ", all)
	return all

}
func NewDhcpv4(db *gorm.DB) *Dhcpv4 {
	return &Dhcpv4{db: db}
}

//func NewSubnetv4(db *gorm.DB) *Dhcpv4 {
//	return &Subnetv4State{db: db}
//}

func (s *Dhcpv4) CreateSubnetv4(subnetv4 *RestSubnetv4) error {
	log.Println("into CreateSubnetv4, subnetv4: ", subnetv4)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv4BySubnet(subnetv4.Subnet); c != nil {
		errStr := "subnet " + subnetv4.Subnet + " already exist"
		return fmt.Errorf(errStr)
	}

	subnetv4.DhcpEnable = 1
	log.Println("in dhcp/dhcprest CreateSubnetv4, subnetv4: ", subnetv4)
	s4, err := PGDBConn.CreateSubnetv4(subnetv4)
	if err != nil {
		return err
	}
	if s4.Subnet == "" {
		return fmt.Errorf("添加子网失败")
	}

	// set newly inserted id
	subnetv4.ID = strconv.Itoa(int(s4.ID))
	subnetv4.SubnetId = strconv.Itoa(int(s4.ID))
	subnetv4.SetCreationTimestamp(s4.CreatedAt)
	log.Println("newly inserted id: ", s4.ID)

	return nil
}

func (s *Dhcpv4) UpdateSubnetv4(subnetv4 *RestSubnetv4) error {
	log.Println("into dhcp/dhcprest/UpdateSubnetv4")
	//log.Println("in UpdateSubnetv4(), subnetv4 ID: ", subnetv4.ID)
	//log.Println("in UpdateSubnetv4(), subnetv4 name: ", subnetv4.Name)
	//log.Println("in UpdateSubnetv4(), subnetv4 subnet: ", subnetv4.Subnet)

	s.lock.Lock()
	defer s.lock.Unlock()

	if c := s.getSubnetv4ById(subnetv4.ID); c == nil {
		return fmt.Errorf("subnet %s not exist", subnetv4.ID)
	}

	err := PGDBConn.OrmUpdateSubnetv4(subnetv4)
	if err != nil {
		return err
	}

	subnetv4.CreationTimestamp = resource.ISOTime(subnetv4.GetCreationTimestamp())
	log.Println("subnetv4.CreationTimestamp ", subnetv4.CreationTimestamp)

	return nil
}

func (s *Dhcpv4) DeleteSubnetv4(subnetv4 *RestSubnetv4) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Println("dhcp/dhcprest DeleteSubnetv4() subnetv4 id: ", subnetv4.ID)
	if c := s.getSubnetv4ById(subnetv4.ID); c == nil {
		return fmt.Errorf("subnet %s not exist", subnetv4.Subnet)
	}

	err := PGDBConn.DeleteSubnetv4(subnetv4.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Dhcpv4) SplitSubnetv4(s4 *RestSubnetv4, newMask int) ([]*RestSubnetv4, error) {
	log.Println("into SplitSubnetv4, s4: ", s4)
	var s4s []*RestSubnetv4
	var ormS4s []*dhcporm.OrmSubnetv4
	var err error

	ormS4 := PGDBConn.GetSubnetv4ById(s4.GetID())
	log.Println("ormS4.subnet: ", ormS4.Subnet)

	out := strings.Split(ormS4.Subnet, "/")
	//log.Println("out: ", out)
	curMask := 0
	if len(out) > 0 {

		curMask = ConvertStringToInt(out[1])
		log.Println("cur mask: ", curMask)
	} else {
		return s4s, fmt.Errorf("current subnet illegal, subnet name: ", ormS4.Subnet)
	}

	// newmask
	if curMask > newMask {
		return s4s, fmt.Errorf("error, new mask is smaller")
	}

	//create new subnetv4s, and delete current one
	log.Println("in dhcp/dhcprest SplitSubnetv4, cur subnet: ", ormS4.Subnet)
	ormS4s, err = PGDBConn.OrmSplitSubnetv4(ormS4, newMask)
	if err != nil {
		return s4s, err
	}
	for _, v := range ormS4s {
		r := s.ConvertSubnetv4FromOrmToRest(v)
		s4s = append(s4s, r)
	}

	return s4s, nil
}
func (s *Dhcpv4) MergeSubnetv4(ids string) (*RestSubnetv4, error) {
	log.Println("into MergeSubnetv4, cidrs: ", ids)

	var newS4 *RestSubnetv4
	var ormS4 *dhcporm.OrmSubnetv4
	var subnetArr []string
	var cidrs string

	idArr := strings.Split(ids, ",")
	for _, id := range idArr {
		subnet := s.getSubnetv4ById(id).Subnet
		subnetArr = append(subnetArr, subnet)
	}
	log.Println("subnetArr: ", subnetArr)
	cidrs = strings.TrimSpace(strings.Join(subnetArr, "\n"))
	log.Println("cidrs:", cidrs, "__")

	newSubnetName, err := GetMergedSubnetv4Name(cidrs)
	if err != nil {
		return newS4, err
	}
	log.Println("new subnet: ", newSubnetName)

	//return newS4, nil

	//todo
	// 1 delete every subnet in cidrs
	// 2 create new subnet with subnet: newSubnet

	//create new subnetv4s, and delete current one
	log.Println("in dhcp/dhcprest MergeSubnetv4, begin to merge subnetv4s")

	ormS4, err = PGDBConn.OrmMergeSubnetv4(idArr, newSubnetName)
	if err != nil {
		return newS4, err
	}
	newS4 = s.ConvertSubnetv4FromOrmToRest(ormS4)
	log.Println("after ormMergeSubnetv4, newS4: ", newS4)

	return newS4, nil
}

func (s *Dhcpv4) GetSubnetv4ById(id string) *RestSubnetv4 {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.getSubnetv4ById(id)
}

func (s *Dhcpv4) getSubnetv4ById(id string) *RestSubnetv4 {

	v := PGDBConn.GetSubnetv4ById(id)
	if v.ID == 0 {
		return nil
	}

	v4 := s.ConvertSubnetv4FromOrmToRest(v)
	return v4
}

func (s *Dhcpv4) getSubnetv4BySubnet(subnet string) *RestSubnetv4 {
	log.Println("In dhcprest getSubnetv4BySubnet, subnet: ", subnet)

	v := PGDBConn.getOrmSubnetv4BySubnet(subnet)
	if v.ID == 0 {
		return nil
	}
	v4 := s.ConvertSubnetv4FromOrmToRest(&v)

	return v4
}

func (s *Dhcpv4) GetSubnetv4s(search *SubnetSearch) []*RestSubnetv4 {
	log.Println("into GetSubnetv4s()")

	//todo get subnet name, usage, totalIP
	usage := server.GetSubnetUsage()
	//log.Println("in Subnetv4List, usage: ", usage)
	var getUsages = map[string]server.DhcpAssignStat{}
	for _, v := range usage.Data {
		//log.Println("in usage.data, k: ", k, ", v.addr: ", v.Addr)
		//log.Println("in usage.data, k: ", k, ", v.Usage: ", v.Usage)
		getUsages[v.Addr] = v
	}
	//log.Println("getUsages: ", getUsages)

	list := PGDBConn.Subnetv4List(search)
	var v4 []*RestSubnetv4
	for _, v := range list {

		var subnet *RestSubnetv4
		subnet = s.ConvertSubnetv4FromOrmToRest(&v)

		subnet.SubnetTotal = "0"
		subnet.SubnetUsage = "0.0"

		//subnet.Name = ""
		if _, ok := getUsages[v.Subnet]; ok {
			//存在

			subnet.SubnetTotal = strconv.Itoa(getUsages[v.Subnet].Total)
			subnet.SubnetUsage = fmt.Sprintf("%.2f", getUsages[v.Subnet].Usage)
		}

		v4 = append(v4, subnet)
		//subnetv4Front.DbS4 = *subnet
	}

	log.Println("GetSubnetv4s, v4: ", v4)
	return v4
}

type subnetv4Handler struct {
	subnetv4s *Dhcpv4
}

func NewSubnetv4Handler(s *Dhcpv4) *subnetv4Handler {
	return &subnetv4Handler{
		subnetv4s: s,
	}
}

type subnetv46Handler struct {
	subnetv46s *Dhcpv4
}

func NewSubnetv46Handler(s *Dhcpv4) *subnetv46Handler {
	return &subnetv46Handler{
		subnetv46s: s,
	}
}

type subnetv4PoolHandler struct {
	subnetv4s *Dhcpv4
}

func (h *subnetv4Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Create")

	subnetv4 := ctx.Resource.(*RestSubnetv4)
	//subnetv4.SetID(subnetv4.Subnet)
	subnetv4.ZoneName = subnetv4.Name
	subnetv4.SetCreationTimestamp(time.Now())
	log.Println("into dhcprest.go Create, subnetv4: ", subnetv4)
	log.Println("into dhcprest.go Create, subnetv4 ValidLifetime: ", subnetv4.ValidLifetime)
	log.Println("into dhcprest.go Create, subnetv4 name: ", subnetv4.Name)
	if err := h.subnetv4s.CreateSubnetv4(subnetv4); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	} else {
		return subnetv4, nil
	}
}

func (h *subnetv4Handler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Update")

	subnetv4 := ctx.Resource.(*RestSubnetv4)
	if err := h.subnetv4s.UpdateSubnetv4(subnetv4); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	if subnetv4.SubnetId == "" {
		subnetv4.SubnetId = subnetv4.ID
	}

	return subnetv4, nil
}

func (h *subnetv4Handler) Delete(ctx *resource.Context) *goresterr.APIError {
	log.Println("into dhcprest.go Delete")
	subnetv4 := ctx.Resource.(*RestSubnetv4)

	if err := h.subnetv4s.DeleteSubnetv4(subnetv4); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil

}

func (h *subnetv4Handler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dhcprest.go List")
	//filter := ctx.GetFilters()
	var search *SubnetSearch
	// no search now
	all := h.subnetv4s.GetSubnetv4s(search)

	log.Println("in list(), before all")
	return all, nil
}

func (h *subnetv46Handler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dhcprest.go List")
	filters := ctx.GetFilters()
	var search SubnetSearch
	for _, filter := range filters {
		log.Println("filter.name: ", filter.Name)
		log.Println("filter.Values: ", filter.Values)
		if len(filter.Values) > 0 {
			subnet := filter.Values[0]
			search.Subnet = subnet
			//search this subnet
			if strings.Contains(subnet, ":") {
				//serach subnetv6

				search.DhcpVer = "v6"
			} else if strings.Contains(subnet, "/") {
				//search subnetv4

				search.DhcpVer = "v4"
			} else {
				// error occurs
				log.Println("subnet search error")
				return nil, nil
			}
		}
	}
	var all []*RestSubnetv46
	all = h.subnetv46s.GetSubnetv46s(&search)

	log.Println("in list(), before all")
	return all, nil
}
func (h *subnetv4Handler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {

	return h.subnetv4s.GetSubnetv4ById(ctx.Resource.GetID()), nil
}

func (h *subnetv4Handler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	var s4s []*RestSubnetv4
	//var retS4 *RestSubnetv4
	var err error
	log.Println("into Action, ctx.Resource: ", ctx.Resource)

	r := ctx.Resource
	mergesplitData, _ := r.GetAction().Input.(*MergeSplitData)

	log.Println("in Action, name: ", r.GetAction().Name)
	log.Println("in Action, oper: ", mergesplitData.Oper)

	switch r.GetAction().Name {
	case "mergesplit":
		if mergesplitData.Oper == "split" {

			mask := ConvertStringToInt(mergesplitData.Mask)
			log.Println("post mask: ", mask)

			if mask < 1 || mask > 32 {
				log.Println("mask error, mask: ", mask)
				return nil, nil
			}

			var s4 *RestSubnetv4
			s4 = ctx.Resource.(*RestSubnetv4)
			if s4s, err = h.subnetv4s.SplitSubnetv4(s4, mask); err != nil {
				return s4s, goresterr.NewAPIError(goresterr.ServerError, err.Error())
			}

			fmt.Println("Action, in mergesplit, s4s: ", s4s)
			//todo split subnetv4 into new mask
			return s4s, nil
		}
		if mergesplitData.Oper == "merge" {
			var s *RestSubnetv4

			ips := mergesplitData.IPs
			log.Println("post ips: ", ips)

			if s, err = h.subnetv4s.MergeSubnetv4(ips); err != nil {
				return s, goresterr.NewAPIError(goresterr.ServerError, err.Error())
			}
			return s, nil
		}

	}
	return nil, nil
}

func (h *optionNameHandler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dhcprest.go optionNameHandler List")
	//option := ctx.Resource.(*RestOptionName)
	//action := option.GetAction()
	//log.Println("action: ", action)
	return h.GetOptionNames(), nil
}

func (r *optionNameHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go optionNameHandler Create")

	opName := ctx.Resource.(*RestOptionName)

	log.Println("dhcp/dhcprest. optionName: ", opName.OptionName)
	if _, err := r.CreateOptionName(opName); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}
	//log.Println("dhcp/dhcprest. pool.id: ", pool.ID)

	return opName, nil
}
func (r *optionNameHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into rest optionName Update")

	opName := ctx.Resource.(*RestOptionName)
	if err := r.UpdateOptionName(opName); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	return opName, nil
}
func (r *optionNameHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	log.Println("into dhcprest.go optionNameHandler Delete")
	opName := ctx.Resource.(*RestOptionName)

	if err := r.DeleteOptionName(opName); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil

}

func (s *optionNameHandler) CreateOptionName(opName *RestOptionName) (*RestOptionName, error) {
	log.Println("into CreateOptionName: ", opName)

	s.lock.Lock()
	defer s.lock.Unlock()

	//todo check whether opName has been created

	op, err := PGDBConn.OrmCreateOptionName(opName)
	if err != nil {
		return opName, err
	}

	// set newly inserted id
	opName.ID = strconv.Itoa(int(op.ID))
	opName.OptionType = op.OptionType
	opName.OptionVer = op.OptionVer
	opName.OptionId = op.OptionId
	opName.OptionName = op.OptionName
	opName.SetCreationTimestamp(op.CreatedAt)
	log.Println("newly inserted op id: ", op.ID)

	return opName, nil
}

func (s *optionNameHandler) UpdateOptionName(opName *RestOptionName) error {
	log.Println("into dhcp/dhcprest/UpdateOptionName")

	s.lock.Lock()
	defer s.lock.Unlock()

	ormOpName := PGDBConn.getOptionNamebyID(opName.ID)
	if ormOpName.OptionName == "" {
		return fmt.Errorf("OptionName %s not exist", opName.ID)
	}

	err := PGDBConn.OrmUpdateOptionName(opName)
	if err != nil {
		return err
	}

	return nil
}

func (r *optionNameHandler) DeleteOptionName(opName *RestOptionName) error {
	log.Println("into DeleteOptionName, opName.ID: ", opName.ID)
	//log.Println(pool)
	r.lock.Lock()
	defer r.lock.Unlock()

	err := PGDBConn.OrmDeleteOptionName(opName.ID)
	if err != nil {
		return err
	}

	return nil
}

func (h *optionNameHandler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	r := ctx.Resource

	//log.Println("in optionName Action, name: ", r.GetAction().Name)
	//log.Println("in optionName Action, oper: ", mergesplitData.Oper)

	switch r.GetAction().Name {
	case "list":
		// list v4 and v6 option numbers
		ret := PGDBConn.GetOptionNameStatistics()
		log.Println("ret: ", ret)

		return ret, nil
	}

	return nil, nil
}

func (r *PoolHandler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dhcprest.go subnetv4PoolHandler List")
	pool := ctx.Resource.(*RestPool)
	return r.GetPools(pool.GetParent().GetID()), nil
}
func (r *PoolHandler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go PoolHandler Get")
	pool := ctx.Resource.(*RestPool)
	return r.GetSubnetv4Pool(pool.GetParent().GetID(), pool.GetID()), nil
}
func (r *PoolHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go pool Create")

	pool := ctx.Resource.(*RestPool)

	log.Println("in PoolHandler, create(), pool: ", pool)
	log.Println("dhcp/dhcprest. pool.Subnetv4Id: ", pool.Subnetv4Id)
	if _, err := r.CreatePool(pool); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	log.Println("dhcp/dhcprest. pool.id: ", pool.ID)

	return pool, nil
}
func (r *PoolHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into rest pool Update")

	pool := ctx.Resource.(*RestPool)
	if err := r.UpdatePool(pool); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	return pool, nil
}
func (r *PoolHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	pool := ctx.Resource.(*RestPool)

	if err := r.DeletePool(pool); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil
}
func (r *PoolHandler) CreatePool(pool *RestPool) (*RestPool, error) {
	log.Println("into dhcprest/CreatePool, pool: ", pool)

	r.lock.Lock()
	defer r.lock.Unlock()

	//todo check whether it exists

	subnetv4ID := pool.GetParent().GetID()
	log.Println("before CreatePool, subnetv4ID:", subnetv4ID)

	if len(pool.Gateway) > 0 || len(pool.DnsServer) > 0 {
		// set dns or gateway under subnet
		//get Restsubnetv4 or RestSubnetv6
		//log.Println("before CreatePool, pool.DnsServer:", pool.DnsServer)
		if err := r.UpdateSubnetv4Server(subnetv4ID, pool); err != nil {
			return nil, err
		}
	}

	pool2, err := PGDBConn.OrmCreatePool(subnetv4ID, pool)
	if err != nil {
		log.Println("OrmCreatePool error")
		log.Println(err)
		return &RestPool{}, err
	}

	pool.SetID(strconv.Itoa(int(pool2.ID)))
	pool.SetCreationTimestamp(pool2.CreatedAt)
	pool.Subnetv4Id = subnetv4ID

	return pool, nil
}
func (r *PoolHandler) UpdateSubnetv4Server(subnetId string, pool *RestPool) error {
	ormSubnetv4 := PGDBConn.GetSubnetv4ById(subnetId)
	log.Println("into UpdateSubnetv4Server, pool.DnsServer: ", pool.DnsServer)
	//log.Println("into UpdateSubnetv4Server, pool.Gateway: ", pool.Gateway)
	ormSubnetv4.DnsServer = pool.DnsServer
	ormSubnetv4.Gateway = pool.Gateway
	ormSubnetv4.ValidLifetime = pool.ValidLifetime
	ormSubnetv4.MaxValidLifetime = pool.MaxValidLifetime
	var s Dhcpv4
	restSubnetv4 := s.ConvertSubnetv4FromOrmToRest(ormSubnetv4)
	//restSubnetv4.Gateway = pool.Gateway
	//restSubnetv4.DnsServer = pool.DnsServer
	if err := s.UpdateSubnetv4(restSubnetv4); err != nil {
		log.Println("in UpdatePool, update subnetv4 gateway error: ", err)
		return err
	}
	return nil
}
func (r *PoolHandler) UpdatePool(pool *RestPool) error {
	log.Println("into UpdatePool")
	log.Println(pool)

	r.lock.Lock()
	defer r.lock.Unlock()

	subnetId := pool.GetParent().GetID()
	log.Println("in UpdatePool, +++subnetId:", subnetId)

	if err := r.UpdateSubnetv4Server(subnetId, pool); err != nil {
		return err
	}
	pool.Subnetv4Id = subnetId

	err := PGDBConn.OrmUpdatePool(subnetId, pool)
	if err != nil {
		return err
	}

	return nil
}
func (r *PoolHandler) DeletePool(pool *RestPool) error {
	log.Println("into DeletePool, pool.ID: ", pool.ID)
	//log.Println(pool)
	r.lock.Lock()
	defer r.lock.Unlock()

	err := PGDBConn.OrmDeletePool(pool.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReservationHandler) List(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	log.Println("into dhcprest.go subnetv4ReservationHandler List")
	rsv := ctx.Resource.(*RestReservation)
	return r.GetReservations(rsv.GetParent().GetID()), nil
}

func (r *ReservationHandler) Get(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go subnetv4ReservationHandler Get")
	rsv := ctx.Resource.(*RestReservation)
	return r.GetSubnetv4Reservation(rsv.GetParent().GetID(), rsv.GetID()), nil
}

func (r *ReservationHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go rsv Create")

	rsv := ctx.Resource.(*RestReservation)

	if _, err := r.CreateReservation(rsv); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	log.Println("+++rsv. rsv.id", rsv.ID)
	log.Print(rsv)

	return rsv, nil
}

func (r *ReservationHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into rest rsv Update")

	rsv := ctx.Resource.(*RestReservation)
	if err := r.UpdateReservation(rsv); err != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, err.Error())
	}

	return rsv, nil
}

func (r *ReservationHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	rsv := ctx.Resource.(*RestReservation)

	if err := r.DeleteReservation(rsv); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil
}

func (r *ReservationHandler) CreateReservation(rsv *RestReservation) (*RestReservation, error) {
	log.Println("into CreateReservation")

	r.lock.Lock()
	defer r.lock.Unlock()

	//todo check whether it exists

	subnetv4ID := rsv.GetParent().GetID()
	log.Println("before OrmCreateReservation")
	rsv2, err := PGDBConn.OrmCreateReservation(subnetv4ID, rsv)
	if err != nil {
		log.Println("OrmCreateReservation error")
		log.Print(err)
		return &RestReservation{}, err
	}

	rsv.SetID(strconv.Itoa(int(rsv2.ID)))
	rsv.SetCreationTimestamp(rsv2.CreatedAt)

	return rsv, nil
}

func (r *ReservationHandler) UpdateReservation(rsv *RestReservation) error {
	log.Println("into UpdateReservation")
	log.Println("input rsv: ", rsv)

	r.lock.Lock()
	defer r.lock.Unlock()

	subnetId := rsv.GetParent().GetID()
	log.Println("UpdateReservation +++subnetId")
	log.Println("subnetId: ", subnetId)
	err := PGDBConn.OrmUpdateReservation(subnetId, rsv)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReservationHandler) DeleteReservation(rsv *RestReservation) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	err := PGDBConn.OrmDeleteReservation(rsv.ID)
	if err != nil {
		return err
	}

	return nil
}
