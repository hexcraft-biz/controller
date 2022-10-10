package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/hexcraft-biz/model"
	"strings"
)

type RoleType uint8

const (
	RoleTypeNone RoleType = iota
	RoleTypeService
	RoleTypeAdmin
	RoleTypeUser
)

type Binding struct {
	Role RoleInterface

	// For URI or row status checking
	Anchor *Resource

	// Insert, Update, Delete
	Write *Resource

	// List, Get
	Output *Resource
}

func NewBinding(c *gin.Context, cfg ConfigInterface, role RoleType) *Binding {
	b := new(Binding)
	switch role {
	case RoleTypeService:
		b.Role = bindRoleService(c, cfg)
	case RoleTypeAdmin:
		b.Role = bindRoleAdmin(c, cfg)
	case RoleTypeUser:
		b.Role = bindRoleUser(c, cfg)
	}

	return b
}

func NewResource(keys interface{}, model model.EngineInterface, qp model.QueryParametersInterface) *Resource {
	return &Resource{
		Keys:            keys,
		Model:           model,
		QueryParameters: qp,
	}
}

func (b *Binding) SetAnchor(keys interface{}, model model.EngineInterface) *Binding {
	b.Anchor = NewResource(keys, model, nil)
	return b
}

func (b *Binding) SetWrite(keys interface{}, model model.EngineInterface) *Binding {
	b.Write = NewResource(keys, model, nil)
	return b
}

func (b *Binding) SetOutput(keys interface{}, model model.EngineInterface, qp model.QueryParametersInterface) *Binding {
	b.Output = NewResource(keys, model, qp)
	return b
}

func (b *Binding) Insert() (sql.Result, error) {
	return b.Write.Model.Insert(b.Write.Keys)
}

func (b *Binding) AnchorHas() (bool, error) {
	return b.Anchor.Model.Has(b.Anchor.Keys)
}

func (b *Binding) AnchorFetchRow(dest interface{}) error {
	return b.Anchor.Model.FetchRow(dest, b.Anchor.Keys)
}

func (b *Binding) OutputRows(paginate bool) (interface{}, error) {
	rows := b.Output.Model.NewRows()
	err := b.Output.Model.FetchRows(rows, b.Output.Keys, b.Output.QueryParameters, paginate)
	return rows, err
}

func (b *Binding) OutputRow() (interface{}, error) {
	row := b.Output.Model.NewRow()
	err := b.Output.Model.FetchRow(row, b.Output.Keys)
	return row, err
}

func (b *Binding) OutputRowByKey(rii ResourceIdentityInterface) (interface{}, error) {
	row := b.Output.Model.NewRow()
	err := b.Output.Model.FetchByKey(row, rii.GetIdentity())
	return row, err
}

func (b *Binding) Update(conds interface{}) (sql.Result, error) {
	return b.Write.Model.Update(conds, b.Write.Keys)
}

func (b *Binding) Delete(conds interface{}) (sql.Result, error) {
	return b.Write.Model.Delete(conds)
}

//----------------------------------------------------------------
// Resource
//----------------------------------------------------------------
type Resource struct {
	Keys            interface{}
	Model           model.EngineInterface
	QueryParameters model.QueryParametersInterface // Only for List()
}

type ResourceIdentityInterface interface {
	GetIdentity() interface{}
}

//----------------------------------------------------------------
// Role
//----------------------------------------------------------------
type RoleInterface interface {
	GetRole() RoleType
	IsLegit() bool
	GetIdentity() string
}

//----------------------------------------------------------------
// RoleService
//----------------------------------------------------------------
type RoleService struct {
	Identity string
}

func (r *RoleService) GetRole() RoleType {
	return RoleTypeService
}

func (r *RoleService) IsLegit() bool {
	return r.Identity != ""
}

func (r *RoleService) GetIdentity() string {
	return r.Identity
}

func bindRoleService(c *gin.Context, cfg ConfigInterface) RoleInterface {
	return &RoleService{
		Identity: c.GetHeader(cfg.GetSchedulerHeader()),
	}
}

//----------------------------------------------------------------
// RoleAdmin
//----------------------------------------------------------------
type RoleAdmin struct {
	Authenticator string
	Identity      string
}

func (r *RoleAdmin) GetRole() RoleType {
	return RoleTypeAdmin
}

func (r *RoleAdmin) IsLegit() bool {
	return r.Identity != ""
}

func (r *RoleAdmin) GetIdentity() string {
	return r.Identity
}

func (r *RoleAdmin) GetAuthenticator() string {
	return r.Authenticator
}

func bindRoleAdmin(c *gin.Context, cfg ConfigInterface) RoleInterface {
	pieces := strings.Split(c.GetHeader("X-Goog-Authenticated-User-Email"), ":")
	r := new(RoleAdmin)

	if len(pieces) == 2 {
		r.Authenticator = pieces[0]
		r.Identity = pieces[1]
	}

	return r
}

//----------------------------------------------------------------
// RoleUser
//----------------------------------------------------------------
type RoleUser struct {
	Identity string
	ID       string `db:"user_id"`
}

func (r *RoleUser) GetRole() RoleType {
	return RoleTypeUser
}

func (r *RoleUser) IsLegit() bool {
	return r.ID != "" && r.Identity != ""
}

func (r *RoleUser) GetIdentity() string {
	return r.Identity
}

func (r *RoleUser) GetID() string {
	return r.ID
}

func bindRoleUser(c *gin.Context, cfg ConfigInterface) RoleInterface {
	headerAffix := cfg.GetHeaderAffix()
	return &RoleUser{
		Identity: c.GetHeader("X-" + headerAffix + "-Authenticated-User-Email"),
		ID:       c.GetHeader("X-" + headerAffix + "-Authenticated-User-Id"),
	}
}
