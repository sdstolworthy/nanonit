package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type ImageRenderer interface {
	Render(deviceID string, appName string) string
}

const DEVICE_IMAGE_PATH = "/render/:deviceID"

func DeviceImageMiddleware(client *firestore.Client, renderer *AppletWrapper, cache map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error rendering image %s", err),
			})
			return
		}
		c.Set("image", f)
	}
}

func ImageCachingMiddleware(cache map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.Param("deviceID")
		image := c.MustGet("image").([]byte)
		md5 := md5.New()
		md5.Write(image)
		hash := hex.EncodeToString(md5.Sum(nil))
		fmt.Println("Hash: ", hash)
		fmt.Println("Cache: ", cache[deviceID])
		if hash == cache[deviceID] {
			fmt.Println(c.Writer.Status())
			c.Status(http.StatusNotModified)
			return
		}
		fmt.Println("Caching image for deviceID: ", deviceID)
		cache[deviceID] = hash
	}
}

func main() {
	r := gin.Default()
	renderer := NewAppletWrapper(os.Getenv("APPS_PATH"))
	sdk, _ := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	opt := option.WithCredentialsJSON(sdk)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	client, err := app.Firestore(context.Background())
	if err != nil {
		panic(err)
	}
	imageHashes := make(map[string]string)

	imageRenderMiddleware := DeviceImageMiddleware(client, renderer, imageHashes)
	imageCacheMiddleware := ImageCachingMiddleware(imageHashes)

	r.GET(DEVICE_IMAGE_PATH, imageRenderMiddleware, imageCacheMiddleware, func(c *gin.Context) {
		image := c.MustGet("image").([]byte)
		c.Data(http.StatusOK, "image/bmp", image)
	})

	r.HEAD(DEVICE_IMAGE_PATH, imageRenderMiddleware, imageCacheMiddleware)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
