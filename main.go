package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ImageRenderer interface {
	Render(deviceID string, appName string) string
}

func main() {
	r := gin.Default()
	renderer := NewAppletWrapper()

	r.GET("/render/:deviceID/:appID", func(c *gin.Context) {
		deviceID := c.Param("deviceID")
		appID := c.Param("appID")
		f, err := renderer.Render(deviceID, appID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		c.JSON(http.StatusOK, gin.H{"Path": f})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
