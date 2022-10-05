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
// Rest: Insert
//================================================================
func (ctrl *Controller) BindPatternInsert(c *gin.Context, b *Binding) error {
	if err := b.BindRole(c, ctrl); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if b.ResourceKeys != nil {
		if err := c.ShouldBindUri(b.ResourceKeys); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
			return err
		}
	}

	if err := c.ShouldBindJSON(b.Payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
		return err
	}

	return nil
}

func (ctrl *Controller) RestInsert(c *gin.Context, b *Binding) error {
	if _, err := b.ModelWrite.Insert(b.Payload); err != nil {
		MysqlErrDefaultResponse(c, err)
		return err
	}

	c.JSON(http.StatusCreated, gin.H{"message": http.StatusText(http.StatusCreated), "results": b.Payload})
	return nil
}

//================================================================
// Rest: List
//================================================================
func (ctrl *Controller) BindPatternList(c *gin.Context, b *Binding) error {
	if err := b.BindRole(c, ctrl); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if b.ResourceKeys != nil {
		if err := c.ShouldBindUri(b.ResourceKeys); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
			return err
		}
	}

	if err := c.ShouldBindQuery(b.QueryParameters); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return err
	}

	return nil
}

func (ctrl *Controller) RestList(c *gin.Context, b *Binding, paginate bool) error {
	dest := b.ModelOutput.NewRows()
	if err := b.ModelOutput.FetchRows(dest, b.ResourceKeys, b.QueryParameters, paginate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return err
	}

	c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	return nil
}

//================================================================
// Rest: Get
//================================================================
func (ctrl *Controller) BindPatternGet(c *gin.Context, b *Binding) error {
	if err := b.BindRole(c, ctrl); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if err := c.ShouldBindUri(b.ResourceKeys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	return nil
}

func (ctrl *Controller) RestGet(c *gin.Context, b *Binding) error {
	dest := b.ModelOutput.NewRow()
	if err := b.ModelOutput.FetchRow(dest, b.ResourceKeys); err != nil {
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
func (ctrl *Controller) RestGetByKey(c *gin.Context, b *Binding, rii ResourceIdentityInterface) error {
	dest := b.ModelOutput.NewRow()
	if err := b.ModelOutput.FetchByKey(dest, rii.GetIdentity()); err != nil {
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
func (ctrl *Controller) BindPatternUpdate(c *gin.Context, b *Binding) error {
	if err := b.BindRole(c, ctrl); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if err := c.ShouldBindUri(b.ResourceKeys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	if err := c.ShouldBindJSON(b.Payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
		return err
	}

	return nil
}

func (ctrl *Controller) RestUpdate(c *gin.Context, b *Binding) error {
	if _, err := b.ModelWrite.Update(b.ResourceKeys, b.Payload); err != nil {
		MysqlErrDefaultResponse(c, err)
		return err
	}

	c.JSON(http.StatusNoContent, nil)
	return nil
}

//================================================================
// Rest: Delete
//================================================================
func (ctrl *Controller) BindPatternDelete(c *gin.Context, b *Binding) error {
	if err := b.BindRole(c, ctrl); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return err
	}

	if err := c.ShouldBindUri(b.ResourceKeys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	return nil
}

func (ctrl *Controller) RestDelete(c *gin.Context, b *Binding) error {
	if result, err := b.ModelResource.Delete(b.ResourceKeys); err != nil {
		MysqlErrDefaultResponse(c, err)
		return err
	} else if affectedRows, err := result.RowsAffected(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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
