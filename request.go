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

type Request struct {
	Role Role
	*Admin
	*User
	ResourceKeys    interface{}
	Payload         interface{}
	QueryParameters model.QueryParametersInterface // Only for List()
	ModelResource   model.EngineInterface
	ModelWrite      model.EngineInterface
	ModelOutput     model.EngineInterface
}

type ResourceIdentityInterface interface {
	GetIdentity() interface{}
}

//----------------------------------------------------------------
// Role
//----------------------------------------------------------------
func (r *Request) BindRole(c *gin.Context, headerAffix string) error {
	switch r.Role {
	case RoleAdmin:
		return r.BindRoleAdmin(c)
	case RoleUser:
		return r.BindRoleUser(c, headerAffix)
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

func (r *Request) BindRoleAdmin(c *gin.Context) error {
	var err error
	pieces := strings.Split(c.GetHeader("X-Goog-Authenticated-User-Email"), ":")

	if len(pieces) != 2 {
		err = errors.New("Invalid user.")
	} else {
		r.Admin = &Admin{
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

func (r *Request) BindRoleUser(c *gin.Context, headerAffix string) error {
	var err error
	id := c.GetHeader("X-" + headerAffix + "-Authenticated-User-Id")
	identity := c.GetHeader("X-" + headerAffix + "-Authenticated-User-Email")

	if id == "" || identity == "" {
		err = errors.New("Invalid user.")
	} else {
		r.User = &User{
			ID:       id,
			Identity: identity,
		}
	}

	return err
}

//----------------------------------------------------------------
// HasResource
//----------------------------------------------------------------
func (r *Request) HasResource() (bool, error) {
	if r.ModelResource != nil {
		return r.ModelResource.Has(r.ResourceKeys)
	}
	return true, nil
}
