package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type ImageRenderer interface {
	Render(deviceID string, appName string) string
}

func main() {
	r := gin.Default()
	renderer := NewAppletWrapper(os.Getenv("APPS_PATH"))
	sdk, _ := base64.StdEncoding.DecodeString(os.Getenv("FIREBASE_SDK"))
	opt := option.WithCredentialsJSON(sdk)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	client, err := app.Firestore(context.Background())
	if err != nil {
		panic(err)
	}

	r.GET("/render/:deviceID", func(c *gin.Context) {
		deviceID := c.Param("deviceID")
		deviceSettings := NewDeviceSettings(deviceID, *client)
		deviceSettings.LoadDeviceSettings()

		fmt.Println("Device settings: ", deviceSettings)

		if deviceSettings.appName == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("App not found %s", deviceSettings.appName),
			})
			return
		}

		f, err := renderer.Render(deviceSettings.appName, deviceSettings.appConfig)
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
