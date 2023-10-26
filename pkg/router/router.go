package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

func PromHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func DBMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func InitRouter(db *gorm.DB) (e *gin.Engine, g *gin.RouterGroup) {
	router := gin.Default()
	router.Use(DBMiddleware(db))
	v1 := router.Group("/api/v1")
	v1.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})
	v1.Handle("GET", "/metrics", PromHandler())
	return router, v1
}

func Routes(g *gin.RouterGroup) {
	createUserRouter(g)
}
