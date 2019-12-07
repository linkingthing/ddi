package dnscontroller

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
	zoneKind = resource.DefaultKindName(Zone{})
	viewKind = resource.DefaultKindName(View{})
	aCLKind  = resource.DefaultKindName(ACL{})
	rRKind   = resource.DefaultKindName(RR{})
	db       *gorm.DB
)

type View struct {
	resource.ResourceBase `json:",inline"`
	Name                  string   `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	Priority              int      `json:"priority" rest:"required=true,min=1,max=100"`
	ACLIDs                []string `json:"aclids" rest:"required=true"`
	zones                 []*Zone  `json:"-"`
}

type Zone struct {
	resource.ResourceBase `json:",inline"`
	Name                  string `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	ZoneFile              string `json:"zone file" rest:"required=true,minLen=1,maxLen=20"`
	rRs                   []*RR  `json:"-"`
}

type ACL struct {
	resource.ResourceBase `json:",inline"`
	Name                  string   `json:"name" rest:"required=true,minLen=1,maxLen=20"`
	IPs                   []string `json:"IP" rest:"required=true"`
}

type RR struct {
	resource.ResourceBase `json:",inline"`
	Data                  string `json:"name" rest:"required=true,minLen=1,maxLen=20"`
}

type aCLHandler struct {
	aCLs *ACLsState
}

func NewACLHandler(s *ACLsState) *aCLHandler {
	return &aCLHandler{
		aCLs: s,
	}
}

func (h *aCLHandler) Create(ctx *resource.Context) (resource.Resource, *goresterr.APIError) {
	aCL := ctx.Resource.(*ACL)
	var one tb.DBACL
	var err error
	if one, err = DBCon.CreateACL(aCL); err != nil {
		return nil, goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	aCL.SetID(strconv.Itoa(int(one.ID)))
	aCL.SetCreationTimestamp(one.CreatedAt)
	return aCL, nil
}

func (h *aCLHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	aCL := ctx.Resource.(*ACL)
	if err := DBCon.DeleteACL(aCL.GetID()); err != nil {
		return goresterr.NewAPIError(goresterr.NotFound, err.Error())
	} else {
		return nil
	}
}

func (h *aCLHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) { //全量
	aCL := ctx.Resource.(*ACL)
	if _, err := DBCon.GetACL(aCL.GetID()); err != nil {
		return nil, goresterr.NewAPIError(goresterr.NotFound, err.Error())
	}
	if err := DBCon.UpdateACL(aCL); err != nil {
		return nil, goresterr.NewAPIError(goresterr.ServerError, err.Error())
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
	var one tb.DBView
	var err error
	if one, err = DBCon.CreateView(view); err != nil {
		return nil, goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	view.SetID(strconv.Itoa(int(one.ID)))
	view.SetCreationTimestamp(one.CreatedAt)
	return view, nil
}

func (h *viewHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	view := ctx.Resource.(*View)
	if err := DBCon.DeleteView(view.GetID()); err != nil {
		return goresterr.NewAPIError(goresterr.NotFound, err.Error())
	} else {
		return nil
	}
}

/*func (h *viewHandler) Update(ctx *resource.Context) (resource.Resource, *goresterr.APIError) { //全量
	view := ctx.Resource.(*View)
	if _, err := DBCon.GetView(view.GetID()); err != nil {
		return nil, goresterr.NewAPIError(goresterr.NotFound, err.Error())
	}
	if err := DBCon.UpdateView(view); err != nil {
		return nil, goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return view, nil
}*/

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
	var dbZone tb.DBZone
	if dbZone, err = DBCon.CreateZone(zone, zone.GetParent().GetID()); err != nil {
		return zone, goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}

	zone.SetID(strconv.Itoa(int(dbZone.ID)))
	zone.SetCreationTimestamp(dbZone.CreatedAt)
	return zone, nil
}

func (h *zoneHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	zone := ctx.Resource.(*Zone)
	if err := DBCon.DeleteZone(zone.GetID(), zone.GetParent().GetID()); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}
	return nil
}
func (h *zoneHandler) List(ctx *resource.Context) interface{} {
	zone := ctx.Resource.(*Zone)
	return DBCon.GetZones(zone.GetParent().GetID())
}

func (h *zoneHandler) Get(ctx *resource.Context) interface{} {
	zone := ctx.Resource.(*Zone)
	one := &Zone{}
	var err error
	if one, err = DBCon.GetZone(zone.GetParent().GetID(), zone.GetID()); err != nil {
		return nil
	}
	return one
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
	var dbRR tb.DBRR
	if dbRR, err = DBCon.CreateRR(rr, rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return rr, goresterr.NewAPIError(goresterr.ServerError, err.Error())
	}

	rr.SetID(strconv.Itoa(int(dbRR.ID)))
	rr.SetCreationTimestamp(dbRR.CreatedAt)
	return rr, nil
}

func (h *rrHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	rr := ctx.Resource.(*RR)
	if err := DBCon.DeleteRR(rr.GetID(), rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return goresterr.NewAPIError(goresterr.ServerError, err.Error())
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

func (h *rrHandler) Get(ctx *resource.Context) interface{} {
	rr := ctx.Resource.(*RR)
	one := &RR{}
	var err error
	if one, err = DBCon.GetRR(rr.GetID(), rr.GetParent().GetID(), rr.GetParent().GetParent().GetID()); err != nil {
		return nil
	}
	return one
}
