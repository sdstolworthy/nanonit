package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"

	"github.com/teacat/noire"
)

func darkenFrame(img image.Image) *image.Paletted {
	newImage := image.NewPaletted(img.Bounds(), palette.Plan9)
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			fmt.Println(uint8(a), uint8(g), uint8(b), a)
			n := noire.NewRGB(float64(uint8(r)), float64(uint8(g)), float64(uint8(b)))
			newColor := n.Shade(0.4)
			newR, newG, newB, _ := newColor.RGBA()
			newImage.Set(x, y, color.RGBA{uint8(newR), uint8(newG), uint8(newB), 1})
		}
	}
	return newImage
}

func Darken(gifImage []byte) ([]byte, error) {
	reader := bytes.NewReader(gifImage)

	img, err := gif.DecodeAll(reader)

	if err != nil {
		return nil, err
	}

	writer := &bytes.Buffer{}

	frameCount := len(img.Image)
	for frame := 0; frame < frameCount; frame++ {
		img.Image[frame] = darkenFrame(img.Image[frame])
	}

	gif.EncodeAll(writer, img)

	return writer.Bytes(), nil
}
