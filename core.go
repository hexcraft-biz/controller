package controller

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/hexcraft-biz/model"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type Controller struct {
	ConfigInterface
}

func New(cfg ConfigInterface) *Controller {
	return &Controller{ConfigInterface: cfg}
}

type ConfigInterface interface {
	GetDB() *sqlx.DB
	GetHeaderAffix() string
	GetGinMode() string
}

//================================================================
// NewRequest
//================================================================
func (ctrl *Controller) NewRequest(role Role, resourceKeys, payload interface{}, qpi model.QueryParametersInterface, meResource, meWrite, meOutput model.EngineInterface) *Request {
	// NOTE: "*" means must have
	return &Request{
		Role:            role,         // *insert, *list, *get, *update, *delete
		ResourceKeys:    resourceKeys, // insert, list, *get, *update, *delete
		Payload:         payload,      // *insert, *update
		QueryParameters: qpi,          // *list
		ModelResource:   meResource,   // insert, list, update, *delete
		ModelWrite:      meWrite,      // *insert, *update
		ModelOutput:     meOutput,     // *list, *get
	}
}

//================================================================
// BindPattern: Insert
//================================================================
func (ctrl *Controller) BindPatternInsert(c *gin.Context, r *Request) error {
	if err := r.BindRole(c, ctrl.GetHeaderAffix()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if r.ResourceKeys != nil {
		if err := c.ShouldBindUri(r.ResourceKeys); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
			return err
		}
	}

	if err := c.ShouldBindJSON(r.Payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
		return err
	}

	return nil
}

//================================================================
// BindPattern: List
//================================================================
func (ctrl *Controller) BindPatternList(c *gin.Context, r *Request) error {
	if err := r.BindRole(c, ctrl.GetHeaderAffix()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if r.ResourceKeys != nil {
		if err := c.ShouldBindUri(r.ResourceKeys); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
			return err
		}
	}

	if err := c.ShouldBindQuery(r.QueryParameters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return err
	}

	return nil
}

//================================================================
// BindPattern: Get
//================================================================
func (ctrl *Controller) BindPatternGet(c *gin.Context, r *Request) error {
	if err := r.BindRole(c, ctrl.GetHeaderAffix()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if err := c.ShouldBindUri(r.ResourceKeys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	return nil
}

//================================================================
// BindPattern: Update
//================================================================
func (ctrl *Controller) BindPatternUpdate(c *gin.Context, r *Request) error {
	if err := r.BindRole(c, ctrl.GetHeaderAffix()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if err := c.ShouldBindUri(r.ResourceKeys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	if err := c.ShouldBindJSON(r.Payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
		return err
	}

	return nil
}

//================================================================
// BindPattern: Delete
//================================================================
func (ctrl *Controller) BindPatternDelete(c *gin.Context, r *Request) error {
	if err := r.BindRole(c, ctrl.GetHeaderAffix()); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if err := c.ShouldBindUri(r.ResourceKeys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	return nil
}

//================================================================
// Rest: Insert
//================================================================
func (ctrl *Controller) RestInsert(c *gin.Context, r *Request) error {
	if _, err := r.ModelWrite.Insert(r.Payload); err != nil {
		MysqlErrDefaultResponse(c, err)
		return err
	}

	c.JSON(http.StatusCreated, gin.H{"message": http.StatusText(http.StatusCreated), "results": r.Payload})
	return nil
}

//================================================================
// Rest: List
//================================================================
func (ctrl *Controller) RestList(c *gin.Context, r *Request, paginate bool) error {
	dest := r.ModelOutput.NewPrototypeList()
	if err := r.ModelOutput.List(dest, r.ResourceKeys, r.QueryParameters, paginate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	return nil
}

//================================================================
// Rest: Get
//================================================================
func (ctrl *Controller) RestGet(c *gin.Context, r *Request) error {
	dest := r.ModelOutput.NewPrototype()
	if err := r.ModelOutput.GetByPrimaryKeys(dest, r.ResourceKeys); err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	return nil
}

//================================================================
// Rest: GetByKey
//================================================================
func (ctrl *Controller) RestGetByKey(c *gin.Context, r *Request, rii ResourceIdentityInterface) error {
	dest := r.ModelOutput.NewPrototype()
	if err := r.ModelOutput.GetByKey(dest, rii.GetIdentity()); err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	return nil
}

//================================================================
// Rest: Update
//================================================================
func (ctrl *Controller) RestUpdate(c *gin.Context, r *Request) error {
	if _, err := r.ModelWrite.UpdateByPrimaryKeys(r.ResourceKeys, r.Payload); err != nil {
		MysqlErrDefaultResponse(c, err)
		return err
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}

//================================================================
// Rest: Delete
//================================================================
func (ctrl *Controller) RestDelete(c *gin.Context, r *Request) error {
	if affectedRows, err := r.ModelResource.DeleteByPrimaryKeys(r.ResourceKeys); err != nil {
		MysqlErrDefaultResponse(c, err)
		return err
	} else if affectedRows <= 0 {
		err := errors.New(http.StatusText(http.StatusNotFound))
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return err
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}

//================================================================
// MysqlErrDefaultResponse
//================================================================
func MysqlErrDefaultResponse(c *gin.Context, err error) {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case model.MysqlErrCodeDuplicateEntry:
			c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
		case model.MysqlErrCodeForeignKeyConstraintFailsCreate:
			c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
		case model.MysqlErrCodeForeignKeyConstraintFailsDelete:
			c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}
