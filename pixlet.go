package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"go.starlark.net/starlark"
	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
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

func (wrapper *AppletWrapper) Render(appName string, config map[string]string) ([]byte, error) {

	appPath := fmt.Sprintf("%s/%[2]s/%[2]s.star", wrapper.appsPath, appName)

	fmt.Println("loading script")
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

	fmt.Println("Config", config)

	roots, err := wrapper.Run(config, threadInitializer)

	if err != nil {
		return []byte{}, err
	}

	screens := encode.ScreensFromRoots(roots)

	encodedWebP, err := screens.EncodeWebP(0)

	image, err := webp.Decode(bytes.NewReader(encodedWebP))

	writer := &bytes.Buffer{}

	err = bmp.Encode(writer, image)

	if err != nil {
		return []byte{}, fmt.Errorf("error rendering")
	}

	return writer.Bytes(), nil
}

func (applet AppletWrapper) loadScript(filepath string, appID string, filename string) error {
	src, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	applet.Load(appID, filename, src, nil)

	return nil
}
