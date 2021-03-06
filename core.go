package controller

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-sql-driver/mysql"
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

//================================================================
// Insert
//================================================================
func (p Prototype) RestfulInsert(c *gin.Context, req model.PrototypeInterface, me model.EngineInterface) {
	if err := c.ShouldBindWith(req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
	} else {
		req.Init()
		if _, err := me.Insert(req); err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				switch mysqlErr.Number {
				case model.MysqlErrCodeConflict:
					c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
				case model.MysqlErrCodeForeignKeyConstraintFails:
					c.JSON(http.StatusUnprocessableEntity, gin.H{"message": http.StatusText(http.StatusUnprocessableEntity)})
				default:
					c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": req})
		}
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

func (p Prototype) RestfulList(c *gin.Context, me model.EngineInterface, dest interface{}, searchCols []string) {
	req := new(ReqList)
	if err := c.ShouldBindWith(req, binding.Query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
	} else if err := me.List(dest, req.Query, searchCols, model.NewPagination(req.Offset, req.Length)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": http.StatusText(http.StatusOK), "results": dest})
	}
}

// Get
func (p Prototype) RestfulGet(c *gin.Context, hook func(interface{}, string) error, dest interface{}, key string) {
	if err := hook(dest, key); err != nil {
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
func (p Prototype) RestfulUpdateByID(c *gin.Context, req interface{}, me model.EngineInterface, id string) {
	if err := c.ShouldBindWith(req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": http.StatusText(http.StatusBadRequest)})
	} else {
		if exists, err := me.Has(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else if !exists {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		} else if _, err := me.UpdateByID(id, req); err != nil {
			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				switch mysqlErr.Number {
				case model.MysqlErrCodeConflict:
					c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
				case model.MysqlErrCodeForeignKeyConstraintFails:
					c.JSON(http.StatusUnprocessableEntity, gin.H{"message": http.StatusText(http.StatusUnprocessableEntity)})
				default:
					c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			}
		} else {
			c.JSON(http.StatusNoContent, nil)
		}
	}
}

//================================================================
// Delete
//================================================================
func (p Prototype) RestfulDeleteByID(c *gin.Context, me model.EngineInterface, id string) {
	if exists, err := me.Has(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
	} else if _, err := me.DeleteByID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	} else {
		c.JSON(http.StatusNoContent, nil)
	}
}
