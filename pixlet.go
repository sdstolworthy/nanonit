package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"go.starlark.net/starlark"
	"tidbyt.dev/pixlet/encode"
	"tidbyt.dev/pixlet/runtime"
	"tidbyt.dev/pixlet/starlarkutil"
)

type AppletWrapper struct {
	*runtime.Applet
}

func NewAppletWrapper() *AppletWrapper {
	return &AppletWrapper{&runtime.Applet{}}
}

func (wrapper *AppletWrapper) Render(deviceID string, appName string) (string, error) {

	appPath := fmt.Sprintf("/home/sdstolworthy/dev/tidbytcommunity/apps/%s/%[1]s.star", appName)

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

	roots, err := wrapper.Run(map[string]string{}, threadInitializer)

	if err != nil {
		return "", err
	}

	screens := encode.ScreensFromRoots(roots)

	webp, err := screens.EncodeWebP(0)

	if err != nil {
		return "", fmt.Errorf("error rendering")
	}

	return base64.StdEncoding.EncodeToString(webp), nil
}

func (applet AppletWrapper) loadScript(filepath string, appID string, filename string) error {
	src, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	applet.Load(appID, filename, src, nil)

	return nil
}
