package controller

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/hexcraft-biz/model"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type Controller struct {
	*sqlx.DB
	ConfigInterface
}

func New(cfg ConfigInterface) *Controller {
	return &Controller{
		DB:              cfg.GetDB(),
		ConfigInterface: cfg,
	}
}

type ConfigInterface interface {
	GetDB() *sqlx.DB
	GetHeaderAffix() string
	GetGinMode() string
}

func (ctrl *Controller) bindRole(c *gin.Context, b *Binding) error {
	if b.Role == nil || !b.Role.IsLegit() {
		return fmt.Errorf("Invalid role.")
	}

	return nil
}

//================================================================
// Rest: Insert
//================================================================
func (ctrl *Controller) BindPatternInsert(c *gin.Context, b *Binding) error {
	if err := ctrl.bindRole(c, b); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": http.StatusText(http.StatusUnauthorized)})
		return err
	}

	if b.Anchor != nil {
		if err := c.ShouldBindUri(b.Anchor.Keys); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
			return err
		}
	}

	if b.Write.Keys != nil {
		if err := c.ShouldBindJSON(b.Write.Keys); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return err
		}
	}

	return nil
}

func (ctrl *Controller) RestInsert(c *gin.Context, b *Binding) error {
	if _, err := b.Insert(); err != nil {
		MysqlErrDefaultResponse(c, err)
		return err
	}

	c.JSON(http.StatusCreated, gin.H{"message": http.StatusText(http.StatusCreated), "results": b.Write.Keys})
	return nil
}

//================================================================
// Rest: List
//================================================================
func (ctrl *Controller) BindPatternList(c *gin.Context, b *Binding) error {
	if err := ctrl.bindRole(c, b); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": http.StatusText(http.StatusUnauthorized)})
		return err
	}

	if b.Anchor != nil {
		if err := c.ShouldBindUri(b.Anchor.Keys); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
			return err
		}
	}

	if b.Output != nil {
		if err := c.ShouldBindQuery(b.Output.QueryParameters); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return err
		}
	}

	return nil
}

func (ctrl *Controller) RestList(c *gin.Context, b *Binding) error {
	rows, err := b.OutputRows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": rows})
	}
	return err
}

//================================================================
// Rest: Get
//================================================================
func (ctrl *Controller) BindPatternGet(c *gin.Context, b *Binding) error {
	if err := ctrl.bindRole(c, b); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": http.StatusText(http.StatusUnauthorized)})
		return err
	}

	if err := c.ShouldBindUri(b.Anchor.Keys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	return nil
}

func (ctrl *Controller) RestGet(c *gin.Context, b *Binding) error {
	if row, err := b.OutputRow(); err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return err
	} else {
		c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": row})
		return nil
	}
}

//================================================================
// Rest: Update
//================================================================
func (ctrl *Controller) BindPatternUpdate(c *gin.Context, b *Binding) error {
	if err := ctrl.bindRole(c, b); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": http.StatusText(http.StatusUnauthorized)})
		return err
	}

	if err := c.ShouldBindUri(b.Anchor.Keys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	if b.Write.Keys != nil {
		if err := c.ShouldBindJSON(b.Write.Keys); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
			return err
		}
	}

	return nil
}

func (ctrl *Controller) RestUpdate(c *gin.Context, b *Binding, conds interface{}) error {
	if _, err := b.Update(conds); err != nil {
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
	if err := ctrl.bindRole(c, b); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": http.StatusText(http.StatusUnauthorized)})
		return err
	}

	if err := c.ShouldBindUri(b.Anchor.Keys); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		return err
	}

	return nil
}

func (ctrl *Controller) RestDelete(c *gin.Context, b *Binding, conds interface{}) error {
	if _, err := b.Delete(conds); err != nil {
		MysqlErrDefaultResponse(c, err)
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
