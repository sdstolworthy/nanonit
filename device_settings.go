package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
)

type DeviceSettings struct {
	AppName   string
	DeviceID  string
	AppConfig map[string]string
}

type DeviceSettingsGetter interface {
	GetSettingsForDeviceByID(deviceID string) (*DeviceSettings, error)
}
type FakeDeviceSettings struct {
	AppName string
	Config  map[string]string
}

func (f *FakeDeviceSettings) GetSettingsForDeviceByID(deviceID string) (*DeviceSettings, error) {
	return &DeviceSettings{AppName: f.AppName, AppConfig: f.Config}, nil
}

type FirebaseDeviceSettings struct {
	firestore firestore.Client
}

type firestoreAppConfig struct {
	App    string            `firestore:"app"`
	Config map[string]string `firestore:"config"`
}

func NewDeviceSettings(app firestore.Client) *FirebaseDeviceSettings {
	return &FirebaseDeviceSettings{firestore: app}
}

func (firebaseSettings *FirebaseDeviceSettings) SaveDeviceSettings(deviceID string, deviceSettings *DeviceSettings) error {
	config := &firestoreAppConfig{App: deviceSettings.AppName, Config: deviceSettings.AppConfig}
	_, err := firebaseSettings.firestore.Collection("devices").Doc(deviceID).Set(context.Background(), map[string]interface{}{
		"app":    config.App,
		"config": config.Config,
	})
	if err != nil {
		return err
	}
	return nil
}

func (firebaseSettings *FirebaseDeviceSettings) GetSettingsForDeviceByID(deviceID string) (*DeviceSettings, error) {
	doc, err := firebaseSettings.firestore.Collection("devices").Doc(deviceID).Get(context.Background())
	if err != nil {
		return nil, err
	}
	var settings DeviceSettings
	if err != nil {
		settings.AppName = "metar"
		settings.AppConfig = map[string]string{"icao": "KPDX,KSLC,KBNA"}
		firebaseSettings.SaveDeviceSettings(deviceID, &settings)
		return &settings, nil
	}

	var appConfig firestoreAppConfig

	fmt.Println("doc.Data(): ", doc.Data())

	doc.DataTo(&appConfig)

	fmt.Println("appConfig: ", appConfig)

	settings.AppName = appConfig.App
	settings.AppConfig = appConfig.Config
	return &settings, nil
}
