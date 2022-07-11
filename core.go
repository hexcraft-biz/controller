package controller

import (
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
				case 1062:
					c.JSON(http.StatusConflict, gin.H{"message": http.StatusText(http.StatusConflict)})
				case 1452:
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
