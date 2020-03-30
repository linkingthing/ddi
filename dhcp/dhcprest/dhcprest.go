package dhcprest

import (
	"fmt"
	"time"

	"log"
	"strconv"

	goresterr "github.com/ben-han-cn/gorest/error"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	"github.com/linkingthing/ddi/cmd/websocket/server"
	"strings"
)

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

	log.Println("in dhcp/dhcprest CreateSubnetv4, subnetv4: ", subnetv4)
	s4, err := PGDBConn.CreateSubnetv4(subnetv4.Name, subnetv4.Subnet, subnetv4.ValidLifetime)
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
	var err error

	ormS4 := PGDBConn.GetSubnetv4ById(s4.GetID())
	log.Println("ormS4.subnet: ", ormS4.Subnet)

	out := strings.Split(ormS4.Subnet, "/")
	log.Println("out: ", out)
	curMask := 0
	if len(out) > 0 {
		ip := out[0]
		log.Println("ip: ", ip)
		ipLong := Ip2long(ip)
		log.Println("ipLong: ", ipLong)

		longIP := Long2ip(ipLong)
		log.Println("longIP: ", longIP)

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
	s4s, err = PGDBConn.OrmSplitSubnetv4(ormS4, newMask)
	if err != nil {
		return s4s, err
	}

	return s4s, nil
}
func (s *Dhcpv4) MergeSubnetv4(s4 *RestSubnetv4, newMask uint) (*RestSubnetv4, error) {
	var newS4 *RestSubnetv4

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

	v4 := s.convertSubnetv4FromOrmToRest(v)
	return v4
}

func (s *Dhcpv4) getSubnetv4BySubnet(subnet string) *RestSubnetv4 {
	log.Println("In dhcprest getSubnetv4BySubnet, subnet: ", subnet)

	v := PGDBConn.getSubnetv4BySubnet(subnet)
	if v.ID == 0 {
		return nil
	}
	v4 := s.convertSubnetv4FromOrmToRest(v)

	return v4
}

func (s *Dhcpv4) GetSubnetv4s() []*RestSubnetv4 {
	log.Println("into GetSubnetv4s()")
	s.lock.Lock()
	defer s.lock.Unlock()

	//var subnetv4Fronts []*dhcporm.OrmSubnetv4Front
	//var subnetv4Front *dhcporm.OrmSubnetv4Front

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

	list := PGDBConn.Subnetv4List()
	var v4 []*RestSubnetv4
	for _, v := range list {
		//log.Println("v.name: ", v.Name)
		//log.Println("v.ID: ", v.ID)
		//log.Println("v.Subnet: ", v.Subnet)
		//log.Println("v.CreatedAt: ", v.CreatedAt)

		var subnet *RestSubnetv4
		subnet = s.convertSubnetv4FromOrmToRest(&v)

		subnet.SubnetTotal = "0"
		subnet.SubnetUsage = "0.0"
		//subnet.Name = ""
		if _, ok := getUsages[v.Subnet]; ok {
			//存在

			//log.Println("---- v.name: ", v.Name)
			//log.Println("---- v.subnet: ", v.Subnet)
			//subnet.Name = getUsages[v.Subnet].Name
			subnet.SubnetTotal = strconv.Itoa(getUsages[v.Subnet].Total)
			subnet.SubnetUsage = fmt.Sprintf("%.2f", getUsages[v.Subnet].Usage)
			//subnet.SubnetUsage = strconv.Itoa(int((collector.Decimal(getUsages[v.Subnet].Usage))))
			//strconv.FormatFloat(getUsages[v.Subnet].Usage, 'f', 5, 64)

			log.Println("--- subnet.Subnet: ", subnet.Subnet)
			log.Println("--- subnet.subnetTotal: ", subnet.SubnetTotal)
			log.Println("--- subnet.SubnetUsage: ", subnet.SubnetUsage)
		}

		v4 = append(v4, subnet)

		//subnetv4Front.DbS4 = *subnet

	}

	//log.Println("GetSubnetv4s, v4: ", v4)
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

type subnetv4PoolHandler struct {
	subnetv4s *Dhcpv4
}

func (h *subnetv4Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	log.Println("into dhcprest.go Create")

	subnetv4 := ctx.Resource.(*RestSubnetv4)
	//subnetv4.SetID(subnetv4.Subnet)
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

func (h *subnetv4Handler) List(ctx *resource.Context) interface{} {
	log.Println("into dhcprest.go List")

	return h.subnetv4s.GetSubnetv4s()
}

func (h *subnetv4Handler) Get(ctx *resource.Context) resource.Resource {

	return h.subnetv4s.GetSubnetv4ById(ctx.Resource.GetID())
}

func (h *subnetv4Handler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	var s4s []*RestSubnetv4
	var retS4 *RestSubnetv4
	var err error
	log.Println("into Action, ctx.Resource: ", ctx.Resource)

	r := ctx.Resource
	var s4 *RestSubnetv4
	s4 = ctx.Resource.(*RestSubnetv4)
	mergesplitData, _ := r.GetAction().Input.(*MergeSplitData)

	log.Println("in Action, name: ", r.GetAction().Name)
	log.Println("in Action, oper: ", mergesplitData.Oper)

	mask := ConvertStringToInt(mergesplitData.Mask)
	log.Println("post mask: ", mask)

	if mask < 1 || mask > 32 {
		log.Println("mask error, mask: ", mask)
		return nil, nil
	}

	switch r.GetAction().Name {
	case "mergesplit":
		if mergesplitData.Oper == "split" {

			if s4s, err = h.subnetv4s.SplitSubnetv4(s4, mask); err != nil {
				return s4s, goresterr.NewAPIError(goresterr.ServerError, err.Error())
			}

			//todo split subnetv4 into new mask
			return s4s, nil
		}
		if mergesplitData.Oper == "merge" {

			return retS4, nil
		}

	}
	return nil, nil
}

func (r *PoolHandler) List(ctx *resource.Context) interface{} {
	log.Println("into dhcprest.go subnetv4PoolHandler List")
	pool := ctx.Resource.(*RestPool)
	return r.GetPools(pool.GetParent().GetID())
}
func (r *PoolHandler) Get(ctx *resource.Context) resource.Resource {
	log.Println("into dhcprest.go PoolHandler Get")
	pool := ctx.Resource.(*RestPool)
	return r.GetSubnetv4Pool(pool.GetParent().GetID(), pool.GetID())
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
	log.Println("before CreatePool, subnetv4ID, pool.getparent.getid: ", subnetv4ID)
	pool2, err := PGDBConn.OrmCreatePool(subnetv4ID, pool)
	if err != nil {
		log.Println("OrmCreatePool error")
		log.Println(err)
		return &RestPool{}, err
	}

	pool.SetID(strconv.Itoa(int(pool2.ID)))
	pool.SetCreationTimestamp(pool2.CreatedAt)

	return pool, nil
}
func (r *PoolHandler) UpdatePool(pool *RestPool) error {
	log.Println("into UpdatePool")
	log.Println(pool)

	r.lock.Lock()
	defer r.lock.Unlock()

	subnetId := pool.GetParent().GetID()
	log.Println("+++subnetId")
	log.Println(subnetId)
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

func (r *ReservationHandler) List(ctx *resource.Context) interface{} {
	log.Println("into dhcprest.go subnetv4ReservationHandler List")
	rsv := ctx.Resource.(*RestReservation)
	return r.GetReservations(rsv.GetParent().GetID())
}

func (r *ReservationHandler) Get(ctx *resource.Context) resource.Resource {
	log.Println("into dhcprest.go subnetv4ReservationHandler Get")
	rsv := ctx.Resource.(*RestReservation)
	return r.GetSubnetv4Reservation(rsv.GetParent().GetID(), rsv.GetID())
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
	log.Println(rsv)

	r.lock.Lock()
	defer r.lock.Unlock()

	subnetId := rsv.GetParent().GetID()
	log.Println("UpdateReservation +++subnetId")
	log.Println(subnetId)
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
