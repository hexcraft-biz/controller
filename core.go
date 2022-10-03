package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-sql-driver/mysql"
	"github.com/hexcraft-biz/misc/xuuid"
	"github.com/hexcraft-biz/model"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type Prototype struct {
	Name string
	DB   *sqlx.DB
}

func New(name string, db *sqlx.DB) *Prototype {
	return &Prototype{
		Name: name,
		DB:   db,
	}
}

type ReqUri struct {
	ID xuuid.UUID `uri:"id" binding:"required" db:"id"`
}

//================================================================
// Insert
//================================================================
func (p Prototype) RestfulInsert(c *gin.Context, me, meHasUri model.EngineInterface, req, hasUriPkCols interface{}) {
	if meHasUri != nil {
		if exists, err := meHasUri.Has(hasUriPkCols); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		} else if !exists {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
			return
		}
	}

	if _, err := me.Insert(req); err != nil {
		MysqlErrDefaultResponse(c, err, nil)
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": http.StatusText(http.StatusCreated), "results": req})
	}
}

//================================================================
// Read
//================================================================

// List
type ReqList struct {
	Query  string `form:"q" binding:"omitempty"`
	Offset uint64 `form:"pos" binding:"omitempty,numeric,min=0"`
	Length uint64 `form:"len" binding:"omitempty,numeric,min=1,max=400"`
}

func (p Prototype) RestfulList(c *gin.Context, me model.EngineInterface, dest, pkCols interface{}, searchCols []string, orderby string) {
	req := new(ReqList)
	if err := c.ShouldBindWith(req, binding.Query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
	} else if err := me.List(dest, pkCols, orderby, req.Query, searchCols, model.NewPagination(req.Offset, req.Length)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	}
}

// Get
func (p Prototype) RestfulGet(c *gin.Context, me model.EngineInterface, dest, pkCols interface{}) {
	if err := me.GetByPrimaryKeys(dest, pkCols); err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	}
}

func (p Prototype) RestfulGetByKey(c *gin.Context, me model.EngineInterface, dest interface{}) {
	uri := new(ReqUri)
	if err := c.ShouldBindUri(uri); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
	} else if err := me.GetByKey(dest, uri.ID.String()); err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	}
}

func (p Prototype) RestfulGetByHook(c *gin.Context, hook func(dest, id interface{}) error, dest interface{}) {
	uri := new(ReqUri)
	if err := c.ShouldBindUri(uri); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
	} else if err := hook(dest, uri.ID.String()); err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	}
}

//================================================================
// Update
//================================================================
func (p Prototype) RestfulUpdateByID(c *gin.Context, me model.EngineInterface, pkCols, req interface{}) {
	if err := c.ShouldBindWith(req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
	} else if exists, err := me.Has(pkCols); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
	} else if _, err := me.UpdateByPrimaryKeys(pkCols, req); err != nil {
		MysqlErrDefaultResponse(c, err, nil)
	} else {
		c.JSON(http.StatusNoContent, nil)
	}
}

//================================================================
// Delete
//================================================================
func (p Prototype) RestfulDeleteByID(c *gin.Context, me model.EngineInterface, pkCols interface{}) {
	if affectedRows, err := me.DeleteByPrimaryKeys(pkCols); err != nil {
		MysqlErrDefaultResponse(c, err, nil)
	} else if affectedRows <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
	} else {
		c.JSON(http.StatusNoContent, nil)
	}
}

//================================================================
// MysqlErrDefaultResponse
//================================================================
func MysqlErrDefaultResponse(c *gin.Context, err error, hook func(*gin.Context, *mysql.MySQLError)) {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case model.MysqlErrCodeDuplicateEntry:
			c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
		case model.MysqlErrCodeForeignKeyConstraintFailsCreate:
			c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
		case model.MysqlErrCodeForeignKeyConstraintFailsDelete:
			c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
		default:
			if hook == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			} else {
				hook(c, mysqlErr)
			}
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}
