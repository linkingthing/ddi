package restfulapi

import (
	//"fmt"

	goresterr "github.com/ben-han-cn/gorest/error"
	"github.com/ben-han-cn/gorest/resource"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	tb "github.com/linkingthing/ddi/dns/cockroachtables"
	"strconv"
)

var (
	version = resource.APIVersion{
		Group:   "linkingthing.com",
		Version: "example/v1",
	}
	aCLKind          = resource.DefaultKindName(ACL{})
	viewKind         = resource.DefaultKindName(View{})
	zoneKind         = resource.DefaultKindName(Zone{})
	rRKind           = resource.DefaultKindName(RR{})
	forwardKind      = resource.DefaultKindName(Forward{})
	redirectionKind  = resource.DefaultKindName(Redirection{})
	defaultDNS64Kind = resource.DefaultKindName(DefaultDNS64{})
	dNS64Kind        = resource.DefaultKindName(DNS64{})
	ipBlackHoleKind  = resource.DefaultKindName(IPBlackHole{})
	recursiveConKind = resource.DefaultKindName(RecursiveConcurrent{})
	sortListKind     = resource.DefaultKindName(SortList{})
	db               *gorm.DB
	FormatError      = goresterr.ErrorCode{"Unauthorized", 400}
)

type View struct {
	resource.ResourceBase `json:",inline"`
	Name                  string         `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	Priority              int            `json:"priority" rest:"required=true,min=1,max=100"`
	ACLIDs                []string       `json:"aclids"`
	Zones                 []*Zone        `json:"-"`
	ACLs                  []*ACL         `json:"acls"`
	ZoneSize              int            `json:"zonesize"`
	Redirections          []*Redirection `json:"-"`
	RPZSize               int            `json:"rpzsize"`
	RedirectSize          int            `json:"redirectsize"`
	DNS64s                []*DNS64       `json:"-"`
	DNS64Size             int            `json:"dns64size"`
}

type Zone struct {
	resource.ResourceBase `json:",inline"`
	Name                  string  `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	rRs                   []*RR   `json:"-"`
	forwards              Forward `json:"-"`
	RRSize                int     `json:"rrsize"`
	ForwarderSize         int     `json:"forwardsize"`
}

type ACL struct {
	resource.ResourceBase `json:",inline"`
	Name                  string   `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	IPs                   []string `json:"IP" rest:"required=true"`
}

type RR struct {
	resource.ResourceBase `json:",inline"`
	Name                  string `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	DataType              string `json:"type" rest:"required=true,minLen=1,maxLen=20"`
	TTL                   uint   `json:"ttl" rest:"required=true"`
	Value                 string `json:"value" rest:"required=true,minLen=1,maxLen=39"`
}

func (z Zone) CreateAction(name string) *resource.Action {
	switch name {
	case "forward":
		return &resource.Action{
			Name:  "forward",
			Input: &ForwardData{},
		}
	default:
		return nil
	}
}

type ForwardData struct {
	resource.ResourceBase `json:",inline"`
	Oper                  string   `json:"oper" rest:"required=true,minLen=1,maxLen=20"`
	ForwardType           string   `json:"type" rest:"required=true,minLen=1,maxLen=20"`
	IPs                   []string `json:"ips" rest:"required=true"`
}

type Forward struct {
	resource.ResourceBase `json:",inline"`
	ForwardType           string   `json:"type" rest:"required=true,minLen=1,maxLen=20"`
	IPs                   []string `json:"ip" rest:"required=true"`
}

type Redirection struct {
	resource.ResourceBase `json:",inline"`
	Name                  string `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	TTL                   uint   `json:"ttl" rest:"required=true"`
	DataType              string `json:"datatype" rest:"required=true,options=A|AAAA|CNAME"`
	RedirectType          string `json:"redirecttype" rest:"required=true,options=rpz|redirect"`
	Value                 string `json:"value" rest:"required=true,minLen=1,maxLen=40"`
}

type DefaultDNS64 struct {
	resource.ResourceBase `json:",inline"`
	Prefix                string `json:"prefix" rest:"required=true,minLen=1,maxLen=39"`
	ClientACL             string `json:"clientacl" rest:"required=true,minLen=1,maxLen=20"`
	ClientACLName         string `json:"clientaclname"`
	AAddress              string `json:"aaddress" rest:"required=true,minLen=1,maxLen=20"`
	AddressName           string `json:"addressname"`
}

type DNS64 struct {
	resource.ResourceBase `json:",inline"`
	Prefix                string `json:"prefix" rest:"required=true,minLen=1,maxLen=39"`
	ClientACL             string `json:"clientacl" rest:"required=true,minLen=1,maxLen=20"`
	ClientACLName         string `json:"clientaclname"`
	AAddress              string `json:"aaddress" rest:"required=true,minLen=1,maxLen=20"`
	AddressName           string `json:"addressname"`
}

type IPBlackHole struct {
	resource.ResourceBase `json:",inline"`
	ACLID                 string `json:"aclid" rest:"required=true"`
	ACLName               string `json:"name"`
}

type RecursiveConcurrent struct {
	resource.ResourceBase `json:",inline"`
	RecursiveClients      int `json:"recursiveClients" rest:"required=true"`
	FetchesPerZone        int `json:"fetchesPerZone" rest:"required=true"`
}

type SortList struct {
	resource.ResourceBase `json:",inline"`
	ACLIDs                []string `json:"aclids" rest:"required=true"`
	ACLs                  []*ACL   `json:"acls"`
}

func (d DNS64) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{View{}}
}

func (h *aCLHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	aCL := ctx.Resource.(*ACL)
	var one tb.ACL
	var err error
	if one, err = DBCon.CreateACL(aCL); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	aCL.SetID(strconv.Itoa(int(one.ID)))
	aCL.SetCreationTimestamp(one.CreatedAt)
	return aCL, nil
}

func (h *aCLHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	aCL := ctx.Resource.(*ACL)
	if err := DBCon.DeleteACL(aCL.GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *aCLHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	aCL := ctx.Resource.(*ACL)
	if _, err := DBCon.GetACL(aCL.GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	if err := DBCon.UpdateACL(aCL); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return aCL, nil
}

func (h *aCLHandler) Get(ctx *resource.Context) resource.Resource {
	var err error
	var aCL *ACL
	if aCL, err = DBCon.GetACL(ctx.Resource.GetID()); err != nil {
		return nil
	}
	return aCL
}
func (h *aCLHandler) List(ctx *resource.Context) interface{} {
	return DBCon.GetACLs()
}

type ACLsState struct {
	ACLs []*ACL
}

func NewACLsState() *ACLsState {
	return &ACLsState{}
}

////////////////////////
type viewHandler struct {
	views *ViewsState
}

func NewViewHandler(s *ViewsState) *viewHandler {
	return &viewHandler{
		views: s,
	}
}

func (h *viewHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	view := ctx.Resource.(*View)
	var one tb.View
	var err error
	if one, err = DBCon.CreateView(view); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	view.SetID(strconv.Itoa(int(one.ID)))
	view.SetCreationTimestamp(one.CreatedAt)
	return view, nil
}

func (h *viewHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	view := ctx.Resource.(*View)
	if err := DBCon.DeleteView(view.GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *viewHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	view := ctx.Resource.(*View)
	if err := DBCon.UpdateView(view); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return view, nil
}

func (h *viewHandler) Get(ctx *resource.Context) resource.Resource {
	var err error
	var view *View
	if view, err = DBCon.GetView(ctx.Resource.GetID()); err != nil {
		return nil
	}
	return view
}

func (h *viewHandler) List(ctx *resource.Context) interface{} {
	return DBCon.GetViews()
}

type ViewsState struct {
	Views []*View
}

func NewViewsState() *ViewsState {
	return &ViewsState{}
}

type zoneHandler struct {
	views *ViewsState
}

func NewZoneHandler(s *ViewsState) *zoneHandler {
	return &zoneHandler{
		views: s,
	}
}

func (h *zoneHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	zone := ctx.Resource.(*Zone)
	var err error
	var dbZone tb.Zone
	if dbZone, err = DBCon.CreateZone(zone, zone.GetParent().GetID()); err != nil {
		return zone, goresterr.NewAPIError(FormatError, err.Error())
	}

	zone.SetID(strconv.Itoa(int(dbZone.ID)))
	zone.SetCreationTimestamp(dbZone.CreatedAt)
	return zone, nil
}

func (h *zoneHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	zone := ctx.Resource.(*Zone)
	if err := DBCon.DeleteZone(zone.GetID(), zone.GetParent().GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	}
	return nil
}

func (h *zoneHandler) List(ctx *resource.Context) interface{} {
	zone := ctx.Resource.(*Zone)
	return DBCon.GetZones(zone.GetParent().GetID())
}

func (h *zoneHandler) Get(ctx *resource.Context) resource.Resource {
	zone := ctx.Resource.(*Zone)
	one := &Zone{}
	var err error
	if one, err = DBCon.GetZone(zone.GetParent().GetID(), zone.GetID()); err != nil {
		return nil
	}
	return one
}

func (h *zoneHandler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	r := ctx.Resource
	var z *Zone
	z = ctx.Resource.(*Zone)
	forwardData, _ := r.GetAction().Input.(*ForwardData)
	switch r.GetAction().Name {
	case "forward":
		if forwardData.Oper == "GET" {
			fw, err := DBCon.GetForward(z.GetID())
			if err != nil {
				return nil, goresterr.NewAPIError(FormatError, err.Error())
			}
			return fw, nil
		}
		if forwardData.Oper == "DEL" {
			if err := DBCon.DeleteForward(z.GetID()); err != nil {
				return nil, goresterr.NewAPIError(FormatError, err.Error())
			}
			return "del success!", nil
		}
		if forwardData.Oper == "MOD" {
			if err := DBCon.UpdateForward(forwardData, z.GetID(), z.GetParent().GetID()); err != nil {
				return nil, goresterr.NewAPIError(FormatError, err.Error())
			}
			return forwardData, nil
		}
	default:
	}
	return nil, nil
}

func (z Zone) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{View{}}
}

func (z Zone) CreateDefaultResource() resource.Resource {
	return &Zone{}
}

type rrHandler struct {
	views *ViewsState
}

func NewRRHandler(s *ViewsState) *rrHandler {
	return &rrHandler{
		views: s,
	}
}

func (h *rrHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	rr := ctx.Resource.(*RR)
	var err error
	var dbRR tb.RR
	if dbRR, err = DBCon.CreateRR(rr, rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return rr, goresterr.NewAPIError(FormatError, err.Error())
	}

	rr.SetID(strconv.Itoa(int(dbRR.ID)))
	rr.SetCreationTimestamp(dbRR.CreatedAt)
	return rr, nil
}

func (h *rrHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	rr := ctx.Resource.(*RR)
	if _, err := DBCon.GetRR(rr.GetID(), rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	if err := DBCon.UpdateRR(rr, rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return rr, nil
}

func (h *rrHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	rr := ctx.Resource.(*RR)
	if err := DBCon.DeleteRR(rr.GetID(), rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	}
	return nil
}
func (h *rrHandler) List(ctx *resource.Context) interface{} {
	rr := ctx.Resource.(*RR)
	var one []*RR
	var err error
	if one, err = DBCon.GetRRs(rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return nil
	}
	return one
}

func (h *rrHandler) Get(ctx *resource.Context) resource.Resource {
	rr := ctx.Resource.(*RR)
	one := &RR{}
	var err error
	if one, err = DBCon.GetRR(rr.GetID(), rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return nil
	}
	return one
}

func (r RR) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{Zone{}}
}

type aCLHandler struct {
	aCLs *ACLsState
}

func NewACLHandler(s *ACLsState) *aCLHandler {
	return &aCLHandler{
		aCLs: s,
	}
}

type ForwardState struct {
	forward *Forward
}

func NewForwardState() *ForwardState {
	return &ForwardState{}
}

type forwardHandler struct {
	forwardState *ForwardState
}

func NewForwardHandler(s *ForwardState) *forwardHandler {
	return &forwardHandler{
		forwardState: s,
	}
}

func (h *forwardHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	forward := ctx.Resource.(*Forward)
	var one tb.DefaultForward
	var err error
	if one, err = DBCon.CreateDefaultForward(forward); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	forward.SetID(strconv.Itoa(int(one.ID)))
	forward.SetCreationTimestamp(one.CreatedAt)
	return forward, nil
}

func (h *forwardHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	forward := ctx.Resource.(*Forward)
	if err := DBCon.DeleteDefaultForward(forward.GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *forwardHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	forward := ctx.Resource.(*Forward)
	if err := DBCon.UpdateDefaultForward(forward); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return forward, nil
}

func (h *forwardHandler) List(ctx *resource.Context) interface{} {
	var one []*Forward
	var err error
	if one, err = DBCon.GetDefaultForwards(); err != nil {
		return nil
	}
	return one
}

func (h *forwardHandler) Get(ctx *resource.Context) resource.Resource {
	fw := ctx.Resource.(*Forward)
	one := &Forward{}
	var err error
	if one, err = DBCon.GetDefaultForward(fw.GetID()); err != nil {
		return nil
	}
	return one
}

type redirectionHandler struct {
	views *ViewsState
}

func NewRedirectionHandler(s *ViewsState) *redirectionHandler {
	return &redirectionHandler{
		views: s,
	}
}

func (r Redirection) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{View{}}
}

func (r *redirectionHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	redirection := ctx.Resource.(*Redirection)
	var one *tb.Redirection
	var err error
	if one, err = DBCon.CreateRedirection(redirection, redirection.GetParent().GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	redirection.SetID(strconv.Itoa(int(one.ID)))
	redirection.SetCreationTimestamp(one.CreatedAt)
	return redirection, nil
}

func (r *redirectionHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	redirection := ctx.Resource.(*Redirection)
	if err := DBCon.DeleteRedirection(redirection.GetID(), redirection.GetParent().GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (r *redirectionHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	redirection := ctx.Resource.(*Redirection)
	if err := DBCon.UpdateRedirection(redirection, redirection.GetParent().GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return redirection, nil
}

func (r *redirectionHandler) List(ctx *resource.Context) interface{} {
	redirection := ctx.Resource.(*Redirection)
	var one []*Redirection
	var err error
	if one, err = DBCon.GetRedirections(redirection.GetParent().GetID()); err != nil {
		return nil
	}
	return one
}

func (r *redirectionHandler) Get(ctx *resource.Context) resource.Resource {
	fw := ctx.Resource.(*Redirection)
	one := &Redirection{}
	var err error
	if one, err = DBCon.GetRedirection(fw.GetID()); err != nil {
		return nil
	}
	return one
}

type DefaultDNS64State struct {
	dns64 *DefaultDNS64
}

func NewDefaultDNS64State() *DefaultDNS64State {
	return &DefaultDNS64State{}
}

type defaultDNS64Handler struct {
	dns64State *DefaultDNS64State
}

func NewDefaultDNS64Handler(s *DefaultDNS64State) *defaultDNS64Handler {
	return &defaultDNS64Handler{
		dns64State: s,
	}
}

func (h *defaultDNS64Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	dns64 := ctx.Resource.(*DefaultDNS64)
	var one *tb.DefaultDNS64
	var err error
	if one, err = DBCon.CreateDefaultDNS64(dns64); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	dns64.SetID(strconv.Itoa(int(one.ID)))
	dns64.SetCreationTimestamp(one.CreatedAt)
	return dns64, nil
}

func (h *defaultDNS64Handler) Delete(ctx *resource.Context) *goresterr.APIError {
	dns64 := ctx.Resource.(*DefaultDNS64)
	if err := DBCon.DeleteDefaultDNS64(dns64.GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *defaultDNS64Handler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	dns64 := ctx.Resource.(*DefaultDNS64)
	if err := DBCon.UpdateDefaultDNS64(dns64); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return dns64, nil
}

func (h *defaultDNS64Handler) List(ctx *resource.Context) interface{} {
	var one []*DefaultDNS64
	var err error
	if one, err = DBCon.GetDefaultDNS64s(); err != nil {
		return nil
	}
	return one
}

func (h *defaultDNS64Handler) Get(ctx *resource.Context) resource.Resource {
	dns64 := ctx.Resource.(*DefaultDNS64)
	one := &DefaultDNS64{}
	var err error
	if one, err = DBCon.GetDefaultDNS64(dns64.GetID()); err != nil {
		return nil
	}
	return one
}

type DNS64Handler struct {
	views *ViewsState
}

func NewDNS64Handler(s *ViewsState) *DNS64Handler {
	return &DNS64Handler{
		views: s,
	}
}

func (h *DNS64Handler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	dns64 := ctx.Resource.(*DNS64)
	var one *tb.DNS64
	var err error
	if one, err = DBCon.CreateDNS64(dns64, dns64.GetParent().GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	dns64.SetID(strconv.Itoa(int(one.ID)))
	dns64.SetCreationTimestamp(one.CreatedAt)
	return dns64, nil
}

func (h *DNS64Handler) Delete(ctx *resource.Context) *goresterr.APIError {
	dns64 := ctx.Resource.(*DNS64)
	if err := DBCon.DeleteDNS64(dns64.GetID(), dns64.GetParent().GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *DNS64Handler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	dns64 := ctx.Resource.(*DNS64)
	if err := DBCon.UpdateDNS64(dns64, dns64.GetParent().GetID()); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return dns64, nil
}

func (h *DNS64Handler) List(ctx *resource.Context) interface{} {
	dns64 := ctx.Resource.(*DNS64)
	var one []*DNS64
	var err error
	if one, err = DBCon.GetDNS64s(dns64.GetParent().GetID()); err != nil {
		return nil
	}
	return one
}

func (h *DNS64Handler) Get(ctx *resource.Context) resource.Resource {
	dns64 := ctx.Resource.(*DNS64)
	one := &DNS64{}
	var err error
	if one, err = DBCon.GetDNS64(dns64.GetID()); err != nil {
		return nil
	}
	return one
}

type IPBlackHoleState struct {
	ipBlackHole *IPBlackHole
}

func NewIPBlackHoleState() *IPBlackHoleState {
	return &IPBlackHoleState{}
}

type ipBlackHoleHandler struct {
	ipBlackHoleState *IPBlackHoleState
}

func NewIPBlackHoleHandler(s *IPBlackHoleState) *ipBlackHoleHandler {
	return &ipBlackHoleHandler{
		ipBlackHoleState: s,
	}
}

func (h *ipBlackHoleHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	ipBlackHole := ctx.Resource.(*IPBlackHole)
	var one *tb.IPBlackHole
	var err error
	if one, err = DBCon.CreateIPBlackHole(ipBlackHole); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	ipBlackHole.SetID(strconv.Itoa(int(one.ID)))
	ipBlackHole.SetCreationTimestamp(one.CreatedAt)
	return ipBlackHole, nil
}

func (h *ipBlackHoleHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	ipBlackHole := ctx.Resource.(*IPBlackHole)
	if err := DBCon.DeleteIPBlackHole(ipBlackHole.GetID()); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}

func (h *ipBlackHoleHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	ipBlackHole := ctx.Resource.(*IPBlackHole)
	if err := DBCon.UpdateIPBlackHole(ipBlackHole); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return ipBlackHole, nil
}

func (h *ipBlackHoleHandler) List(ctx *resource.Context) interface{} {
	var many []*IPBlackHole
	var err error
	if many, err = DBCon.GetIPBlackHoles(); err != nil {
		return nil
	}
	return many
}

func (h *ipBlackHoleHandler) Get(ctx *resource.Context) resource.Resource {
	ipBlackHole := ctx.Resource.(*IPBlackHole)
	one := &IPBlackHole{}
	var err error
	if one, err = DBCon.GetIPBlackHole(ipBlackHole.GetID()); err != nil {
		return nil
	}
	return one
}

////////////
type RecursiveConcurrentState struct {
	con *RecursiveConcurrent
}

func NewRecursiveConcurrentState() *RecursiveConcurrentState {
	return &RecursiveConcurrentState{}
}

type recursiveConcurrentHandler struct {
	conState *RecursiveConcurrentState
}

func NewRecursiveConcurrentHandler(s *RecursiveConcurrentState) *recursiveConcurrentHandler {
	return &recursiveConcurrentHandler{
		conState: s,
	}
}

func (h *recursiveConcurrentHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	con := ctx.Resource.(*RecursiveConcurrent)
	if err := DBCon.UpdateRecursiveConcurrent(con); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return con, nil
}

func (h *recursiveConcurrentHandler) List(ctx *resource.Context) interface{} {
	var many []*RecursiveConcurrent
	var err error
	if many, err = DBCon.GetRecursiveConcurrents(); err != nil {
		return nil
	}
	return many
}

type SortListsState struct {
	SortLists []*SortList
}

func NewSortListsState() *SortListsState {
	return &SortListsState{}
}

type sortListHandler struct {
	sortLists *SortListsState
}

func NewSortListHandler(s *SortListsState) *sortListHandler {
	return &sortListHandler{
		sortLists: s,
	}
}

func (h *sortListHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	sortList := ctx.Resource.(*SortList)
	var one *tb.SortListElement
	var err error
	if one, err = DBCon.CreateSortList(sortList); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	sortList.SetID("1")
	sortList.SetCreationTimestamp(one.CreatedAt)
	return sortList, nil
}

func (h *sortListHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	if err := DBCon.DeleteSortList(); err != nil {
		return goresterr.NewAPIError(FormatError, err.Error())
	} else {
		return nil
	}
}
func (h *sortListHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	sortList := ctx.Resource.(*SortList)
	if err := DBCon.UpdateSortList(sortList); err != nil {
		return nil, goresterr.NewAPIError(FormatError, err.Error())
	}
	return sortList, nil
}

func (h *sortListHandler) Get(ctx *resource.Context) resource.Resource {
	var err error
	var sortList *SortList
	if sortList, err = DBCon.GetSortList(); err != nil {
		return nil
	}
	return sortList
}
