package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
	"testing"
)

type Config struct {
	DB          *sqlx.DB
	HeaderAffix string
	GinMode     string
}

func (c *Config) GetDB() *sqlx.DB {
	return c.DB
}

func (c *Config) GetHeaderAffix() string {
	return c.HeaderAffix
}

func (c *Config) GetGinMode() string {
	return c.GinMode
}

type ReqRK struct {
	ID     string `uri:"id" binding:"required" json:"id"`
	Status string `json:"status"`
}

type TestController struct {
	*Controller
}

func NewTestController(cfg *Config) *TestController {
	return &TestController{Controller: New(cfg)}
}

func (tc *TestController) Get() gin.HandlerFunc {
	return func(c *gin.Context) {
		rk := new(ReqRK)
		if err := c.ShouldBindUri(rk); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": http.StatusText(http.StatusNotFound)})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Pass", "results": rk})
		}
	}
}

func TestRest(t *testing.T) {
	route := gin.Default()
	cfg := &Config{DB: nil, HeaderAffix: "Test", GinMode: "testing"}
	tc := NewTestController(cfg)

	route.GET("/rewards/:id", tc.Get())
	route.Run(":8088")
}
