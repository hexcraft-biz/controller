package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/hexcraft-biz/model"
	"strings"
)

type Role uint8

const (
	RoleNone Role = iota
	RoleAdmin
	RoleUser
)

type Binding struct {
	Role            Role
	ResourceKeys    interface{}
	Payload         interface{}
	QueryParameters model.QueryParametersInterface // Only for List()
	ModelResource   model.EngineInterface
	ModelWrite      model.EngineInterface
	ModelOutput     model.EngineInterface
	*Admin
	*User
}

type ResourceIdentityInterface interface {
	GetIdentity() interface{}
}

//----------------------------------------------------------------
// Role
//----------------------------------------------------------------
func (b *Binding) BindRole(c *gin.Context, headerAffix string) error {
	switch b.Role {
	case RoleAdmin:
		return b.BindRoleAdmin(c)
	case RoleUser:
		return b.BindRoleUser(c, headerAffix)
	default:
		return nil
	}
}

//----------------------------------------------------------------
// Admin
//----------------------------------------------------------------
type Admin struct {
	Authenticator string
	Email         string
}

func (b *Binding) BindRoleAdmin(c *gin.Context) error {
	var err error
	pieces := strings.Split(c.GetHeader("X-Goog-Authenticated-User-Email"), ":")

	if len(pieces) != 2 {
		err = errors.New("Invalid user.")
	} else {
		b.Admin = &Admin{
			Authenticator: pieces[0],
			Email:         pieces[1],
		}
	}

	return err
}

//----------------------------------------------------------------
// User
//----------------------------------------------------------------
type User struct {
	ID       interface{} `db:"user_id"`
	Identity string
}

func (b *Binding) BindRoleUser(c *gin.Context, headerAffix string) error {
	var err error
	id := c.GetHeader("X-" + headerAffix + "-Authenticated-User-Id")
	identity := c.GetHeader("X-" + headerAffix + "-Authenticated-User-Email")

	if id == "" || identity == "" {
		err = errors.New("Invalid user.")
	} else {
		b.User = &User{
			ID:       id,
			Identity: identity,
		}
	}

	return err
}

//----------------------------------------------------------------
// HasResource
//----------------------------------------------------------------
func (b *Binding) HasResource() (bool, error) {
	if b.ModelResource != nil {
		return b.ModelResource.Has(b.ResourceKeys)
	}
	return true, nil
}
