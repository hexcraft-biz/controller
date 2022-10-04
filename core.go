package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
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

type ReqList struct {
	Query string `form:"q" binding:"omitempty"`
	Pos   uint16 `form:"pos" binding:"omitempty"`
	Len   uint16 `form:"len" binding:"omitempty"`
}

//================================================================
// Read
//================================================================

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

func (p Prototype) RestfulGetByKey(c *gin.Context, me model.EngineInterface, dest, key string) {
	if err := me.GetByKey(dest, key); err != nil {
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
	if exists, err := me.Has(pkCols); err != nil {
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
