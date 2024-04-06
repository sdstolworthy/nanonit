package gif_manipulator

import (
	"bytes"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"

	"github.com/teacat/noire"
)

func NewGifManipulator(gif gif.GIF) *GifManipulator {
	return &GifManipulator{gif: gif}
}

type GifManipulator struct {
	gif gif.GIF
}

func FromBytes(image []byte) (*GifManipulator, error) {
	reader := bytes.NewReader(image)

	img, err := gif.DecodeAll(reader)

	if err != nil {
		return nil, err
	}
	return &GifManipulator{*img}, nil
}

func toStraightAlpha(c color.Color) (uint8, uint8, uint8, uint8) {
	r, g, b, a := c.RGBA()
	alphaFloat := float64(a)

	nr := float64(r) / alphaFloat * 255
	ng := float64(g) / alphaFloat * 255
	nb := float64(b) / alphaFloat * 255
	uint16Max := ^uint16(0)
	na := uint16(alphaFloat) / uint16Max * 255

	return uint8(nr), uint8(ng), uint8(nb), uint8(na)
}

func darkenFrame(img image.Image, shade float64) *image.Paletted {
	newImage := image.NewPaletted(img.Bounds(), palette.WebSafe)
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			c := img.At(x, y)
			r, g, b, a := toStraightAlpha(c)

			n := noire.Color{Red: float64(r), Green: float64(g), Blue: float64(b), Alpha: float64(a) / 255}
			n = n.Shade(shade)
			nr, ng, nb, na := n.RGBA()

			newImage.Set(x, y, color.RGBA{uint8(nr), uint8(ng), uint8(nb), uint8(na * 255)})
		}
	}
	return newImage
}

func (g *GifManipulator) ToBytes() ([]byte, error) {
	writer := &bytes.Buffer{}
	gif.EncodeAll(writer, &g.gif)

	return writer.Bytes(), nil
}

func (g *GifManipulator) Darken(shade float64) error {
	frameCount := len(g.gif.Image)
	for frame := 0; frame < frameCount; frame++ {
		g.gif.Image[frame] = darkenFrame(g.gif.Image[frame], shade)
	}
	return nil
}
