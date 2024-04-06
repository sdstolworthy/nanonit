package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sdstolworthy/nanonit/gif_manipulator"
	"google.golang.org/api/option"
	"net/http"
)

type ImageRenderer interface {
	Render(deviceID string, appName string) string
}

const DEVICE_IMAGE_PATH = "/render/:deviceID"

func DeviceImageMiddleware(deviceSettingsGetter DeviceSettingsGetter, renderer *AppletWrapper, cache map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.Param("deviceID")
		deviceSettings, err := deviceSettingsGetter.GetSettingsForDeviceByID(deviceID)

		fmt.Println("Device settings: ", deviceSettings)

		if deviceSettings.AppName == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("App not found %s", deviceSettings.AppName),
			})
			return
		}

		fmt.Println(deviceSettings)
		f, err := renderer.Render(deviceSettings.AppName, deviceSettings.AppConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error rendering image %s", err),
			})
			return
		}
		f, err = gif_manipulator.Darken(f)
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
	config, err := GetConfig()
	if err != nil {
		panic(err)
	}
	renderer := NewAppletWrapper(config.AppsPath)
	imageHashes := make(map[string]string)

	var settings DeviceSettingsGetter

	credentials, err := config.GetGoogleApplicationCredentials()
	fmt.Println("hello,credentials")
	fmt.Println("Credentials ", len(credentials))

	if len(credentials) > 0 {
		opt := option.WithCredentialsJSON(credentials)
		app, err := firebase.NewApp(context.Background(), nil, opt)
		client, err := app.Firestore(context.Background())
		if err != nil {
			panic(err)
		}
		settings = NewDeviceSettings(*client)
	} else {
		settings = &FakeDeviceSettings{}
	}

	imageRenderMiddleware := DeviceImageMiddleware(settings, renderer, imageHashes)
	imageCacheMiddleware := ImageCachingMiddleware(imageHashes)

	r.GET(DEVICE_IMAGE_PATH, imageRenderMiddleware, imageCacheMiddleware, func(c *gin.Context) {
		image := c.MustGet("image").([]byte)
		c.Data(http.StatusOK, "image/gif", image)
	})

	r.HEAD(DEVICE_IMAGE_PATH, imageRenderMiddleware, imageCacheMiddleware)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
