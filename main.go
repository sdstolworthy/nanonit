package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"github.com/sdstolworthy/nanonit/gif_manipulator"
	"google.golang.org/api/option"
)

type ImageRenderer interface {
	Render(deviceID string, appName string) string
}

const deviceImagePath = "/render/:deviceID"
const imageContext = "image"

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

		f, err := renderer.Render(deviceSettings.AppName, deviceSettings.AppConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error rendering image %s", err),
			})
			return
		}
		c.Set(imageContext, f)
	}
}

func ImageManipulationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		image := c.MustGet(imageContext).([]byte)
		gif, err := gif_manipulator.FromBytes(image)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to load image",
			})
		}
		rawShade := c.Query("shade")
		if rawShade != "" {
			shade, err := strconv.ParseFloat(rawShade, 64)
			if err != nil {
				fmt.Println(fmt.Sprintf("Failed to convert shade: %s", rawShade))
			} else {
				fmt.Println(fmt.Sprintf("Applying shade: %v", shade))
				gif.Darken(shade)
			}
		} else {
			fmt.Println("Did not get shade param")
		}
		imgBytes, err := gif.ToBytes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to manipulate image",
			})
		}

		c.Set(imageContext, imgBytes)
	}
}

func ImageCachingMiddleware(cache map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.Param("deviceID")
		image := c.MustGet(imageContext).([]byte)
		md5 := md5.New()
		md5.Write(image)
		hash := hex.EncodeToString(md5.Sum(nil))
		fmt.Println("Hash: ", hash)
		fmt.Println("Cache: ", cache[deviceID])
		if hash == cache[deviceID] {
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
		settings = &FakeDeviceSettings{AppName: "detailedmetar", Config: map[string]string{"airports": "KPDX", "icao": "KPDX"}}
	}

	imageRenderMiddleware := DeviceImageMiddleware(settings, renderer, imageHashes)
	imageCacheMiddleware := ImageCachingMiddleware(imageHashes)
	imageManipulationMiddleware := ImageManipulationMiddleware()

	r.GET(deviceImagePath, imageRenderMiddleware, imageManipulationMiddleware, imageCacheMiddleware, func(c *gin.Context) {
		image := c.MustGet(imageContext).([]byte)
		c.Data(http.StatusOK, "image/gif", image)
	})

	r.HEAD(deviceImagePath, imageRenderMiddleware, imageCacheMiddleware)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
