package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
)

type ImageRenderer interface {
	Render(deviceID string, appName string) string
}

func main() {
	r := gin.Default()
	renderer := NewAppletWrapper(os.Getenv("APPS_PATH"))
	app, err := firebase.NewApp(context.Background(), nil)
	client, err := app.Firestore(context.Background())
	if err != nil {
		panic(err)
	}

	r.GET("/render/:deviceID", func(c *gin.Context) {
		deviceID := c.Param("deviceID")
		deviceSettings := NewDeviceSettings(deviceID, *client)
		deviceSettings.LoadDeviceSettings()
		appID := "metar"

		fmt.Println("Device settings: ", deviceSettings)
		if deviceSettings.appName != "" {
			appID = deviceSettings.appName
		}

		f, err := renderer.Render(appID, deviceSettings.appConfig)
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		c.Data(http.StatusOK, "image/bmp", f)
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
