package controller

import (
	"errors"
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
	Role              RoleType
	ResourceKeys      interface{}
	Payload           interface{}
	QueryParameters   model.QueryParametersInterface // Only for List()
	ModelReadResource model.EngineInterface
	ModelWrite        model.EngineInterface
	ModelOutput       model.EngineInterface
	*RoleService
	*RoleAdmin
	*RoleUser
}

type ResourceIdentityInterface interface {
	GetIdentity() interface{}
}

//----------------------------------------------------------------
// Role
//----------------------------------------------------------------
func (b *Binding) BindRole(c *gin.Context, cfg ConfigInterface) error {
	switch b.Role {
	case RoleTypeService:
		return b.BindRoleService(c, cfg)
	case RoleTypeAdmin:
		return b.BindRoleAdmin(c, cfg)
	case RoleTypeUser:
		return b.BindRoleUser(c, cfg)
	default:
		return nil
	}
}

//----------------------------------------------------------------
// RoleService
//----------------------------------------------------------------
type RoleService struct {
	ClientID string
}

func (b *Binding) BindRoleService(c *gin.Context, cfg ConfigInterface) error {
	var err error
	clientID := c.GetHeader("X-" + cfg.GetHeaderAffix() + "-Authenticated-Client-Id")

	// TODO: Might add more verification.
	if clientID == "" {
		err = errors.New("Invalid user.")
	} else {
		b.RoleService = &RoleService{ClientID: clientID}
	}

	return err
}

//----------------------------------------------------------------
// RoleAdmin
//----------------------------------------------------------------
type RoleAdmin struct {
	Authenticator string
	Email         string
}

func (b *Binding) BindRoleAdmin(c *gin.Context, cfg ConfigInterface) error {
	var err error
	pieces := strings.Split(c.GetHeader("X-Goog-Authenticated-User-Email"), ":")

	if len(pieces) != 2 {
		err = errors.New("Invalid user.")
	} else {
		b.RoleAdmin = &RoleAdmin{
			Authenticator: pieces[0],
			Email:         pieces[1],
		}
	}

	return err
}

//----------------------------------------------------------------
// RoleUser
//----------------------------------------------------------------
type RoleUser struct {
	ID       interface{} `db:"user_id"`
	Identity string
}

func (b *Binding) BindRoleUser(c *gin.Context, cfg ConfigInterface) error {
	var err error
	headerAffix := cfg.GetHeaderAffix()
	id := c.GetHeader("X-" + headerAffix + "-Authenticated-User-Id")
	identity := c.GetHeader("X-" + headerAffix + "-Authenticated-User-Email")

	if id == "" || identity == "" {
		err = errors.New("Invalid user.")
	} else {
		b.RoleUser = &RoleUser{
			ID:       id,
			Identity: identity,
		}
	}

	return err
}

//----------------------------------------------------------------
// Fetch
//----------------------------------------------------------------
func (b *Binding) HasResource() (bool, error) {
	return b.ModelReadResource.Has(b.ResourceKeys)
}

func (b *Binding) GetResource(dest interface{}) error {
	return b.ModelReadResource.FetchRow(dest, b.ResourceKeys)
}
