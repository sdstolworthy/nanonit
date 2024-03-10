package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type DeviceSettings struct {
	appName   string
	deviceID  string
	appConfig map[string]string
	firestore firestore.Client
}

type firestoreAppConfig struct {
	App    string            `firestore:"app"`
	Config map[string]string `firestore:"config"`
}

func (deviceSettings *DeviceSettings) String() string {
	return fmt.Sprintf("DeviceSettings{appName: %s, deviceID: %s, appConfig: %v}", deviceSettings.appName, deviceSettings.deviceID, deviceSettings.appConfig)
}

func NewDeviceSettings(deviceID string, app firestore.Client) *DeviceSettings {
	return &DeviceSettings{"", deviceID, map[string]string{}, app}
}

func (deviceSettings *DeviceSettings) LoadDeviceSettings() error {

	doc, err := deviceSettings.firestore.Collection("devices").Doc(deviceSettings.deviceID).Get(context.Background())
	if !doc.Exists() {
		defaultAppConfig := firestoreAppConfig{"metar", map[string]string{"icao": "KJFK"}}
		_, err := deviceSettings.firestore.Collection("devices").Doc(deviceSettings.deviceID).Set(context.Background(), &defaultAppConfig, firestore.MergeAll)
		if err != nil {
			fmt.Println(err)
			return err
		}
		deviceSettings.appName = "metar"
		deviceSettings.appConfig = map[string]string{"icao": "KJFK"}
		return nil
	}
	var appConfig firestoreAppConfig

	fmt.Println("doc.Data(): ", doc.Data())

	doc.DataTo(&appConfig)

	fmt.Println("appConfig: ", appConfig)

	deviceSettings.appName = appConfig.App
	deviceSettings.appConfig = appConfig.Config

	if err != nil {
		return err
	}
	return nil
}