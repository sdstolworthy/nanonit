package main

import (
	"bytes"
	"context"
	"fmt"
	"image/gif"
	"io/ioutil"
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

	fmt.Println("Loading appName: ", appName)
	appPath := fmt.Sprintf("%s/%[2]s/%[2]s.star", wrapper.appsPath, appName)

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

	webpResult, err := screens.EncodeWebP(0)
	os.WriteFile("output.webp", webpResult, 0644)
	if err != nil {
		return []byte{}, fmt.Errorf("error rendering")
	}

	// return a bitmap from the rendered webp

	decoded, err := webp.Decode(bytes.NewReader(webpResult))

	writer := new(bytes.Buffer)
	err = bmp.Encode(writer, decoded)
  fmt.Println(writer.Bytes())
	if err != nil {
		return []byte{}, fmt.Errorf("error encoding bmp")
	}

	// frames, err := gif.DecodeAll(bytes.NewReader(encodedGIF))

	//	for i, frame := range frames.Image {
	//		filename := fmt.Sprintf("frame-%d.bmp", i)
	//		file, err := os.Create(filename)
	//		if err != nil {
	//			return []byte{}, fmt.Errorf("error creating file")
	//		}
	//		defer file.Close()
	//    bmp.Encode(file, frame)
	//	}

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
