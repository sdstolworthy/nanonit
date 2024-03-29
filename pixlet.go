package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.starlark.net/starlark"
	"gopkg.in/yaml.v3"
	"tidbyt.dev/pixlet/encode"
	"tidbyt.dev/pixlet/runtime"
	"tidbyt.dev/pixlet/starlarkutil"
)

type AppletWrapper struct {
	*runtime.Applet
	appsPath string
}

func NewAppletWrapper(appsPath string) *AppletWrapper {
	return &AppletWrapper{&runtime.Applet{}, appsPath}
}

type Manifest struct {
	Filename    string `yaml:"fileName"`
	Name        string `yaml:"name"`
	Description string `yaml:"desc"`
	PackageName string `yaml:"packageName"`
}

func (manifest *Manifest) String() string {
	return fmt.Sprintf("Manifest{Filename: %s, Name: %s, Description: %s, PackageName: %s}", manifest.Filename, manifest.Name, manifest.Description, manifest.PackageName)
}

func (wrapper *AppletWrapper) LoadManifest(appName string) (*Manifest, error) {
	manifestPath := fmt.Sprintf("%s/%[2]s/manifest.yaml", wrapper.appsPath, appName)
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	manifestData := &Manifest{}

	err = yaml.Unmarshal(manifest, manifestData)
	if err != nil {
		return nil, err
	}
	return manifestData, nil
}

func (wrapper *AppletWrapper) Render(appName string, config map[string]string) ([]byte, error) {
	fmt.Println("Loading appName: ", appName)
	manifest, err := wrapper.LoadManifest(appName)
	if err != nil {
		return []byte{}, err
	}
	appPath := fmt.Sprintf("%s/%s/%s", wrapper.appsPath, appName, manifest.Filename)
	wrapper.loadScript(appPath, appName, appName)
	timeout := 15000
	threadInitializer := func(thread *starlark.Thread) *starlark.Thread {
		ctx, _ := context.WithTimeoutCause(
			context.Background(),
			time.Duration(timeout)*time.Millisecond,
			fmt.Errorf("timeout after %dms", timeout),
		)
		starlarkutil.AttachThreadContext(ctx, thread)
		return thread
	}

	roots, err := wrapper.Run(config, threadInitializer)

	if err != nil {
		return []byte{}, err
	}

	screens := encode.ScreensFromRoots(roots)

	fmt.Println("Encoding GIF")
	encodedGif, err := screens.EncodeGIF(0)

	if err != nil {
		return []byte{}, fmt.Errorf("error rendering")
	}

	return encodedGif, nil
}

func (applet AppletWrapper) loadScript(filepath string, appID string, filename string) error {
	src, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	applet.Load(appID, filename, src, nil)

	return nil
}
