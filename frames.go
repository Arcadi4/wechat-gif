package main

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
)

const (
	maxWidth        = 1000
	maxHeight       = 1000
	maxImageSize    = 5242880
	maxAutoplaySize = 1048576
)

var paletteRgbCompressed = color.Palette{}

func init() {
	for r := 0; r < 32; r++ {
		for g := 0; g < 32; g++ {
			for b := 0; b < 32; b++ {
				paletteRgbCompressed = append(
					paletteRgbCompressed,
					color.RGBA{
						R: uint8(r * 8), G: uint8(g * 8), B: uint8(b * 8),
					},
				)
			}
		}
	}
}

// resizeGifFrames gives a gif with
//   - Same ratio as the original gif
//   - width < x AND height < y
func resizeGifFrames(g *gif.GIF, maxX int, maxY int) (new *gif.GIF) {
	new = &gif.GIF{
		Image:           make([]*image.Paletted, len(g.Image)),
		Delay:           g.Delay,
		Disposal:        g.Disposal,
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
		LoopCount:       g.LoopCount,
	}
	copy(new.Image, g.Image)

	for i, frame := range g.Image {
		x := frame.Stride
		y := len(frame.Pix) / x
		if y > maxY || x > maxX {
			var resizedFrame *image.NRGBA
			if x > y {
				resizedFrame = imaging.Resize(
					frame,
					maxX,
					0,
					imaging.Lanczos,
				)
			} else {
				resizedFrame = imaging.Resize(
					frame,
					0,
					maxY,
					imaging.Lanczos,
				)
			}
			new.Image[i] = nrgbaToPaletted(resizedFrame, frame.Palette)
		}

	}

	updateGifConfig(new)
	return new
}

func updateGifConfig(gif *gif.GIF) {
	largestWidth, largestHeight := 0, 0
	for _, frame := range gif.Image {
		bound := frame.Bounds()
		if bound.Dx() > largestWidth {
			largestWidth = bound.Dx()
		}
		if bound.Dy() > largestHeight {
			largestHeight = bound.Dy()
		}
	}
	// Dynamically adjust the canvas size to fit the largest frame
	gif.Config.Width, gif.Config.Height = largestWidth, largestHeight
}

func nrgbaToPaletted(
nrgba *image.NRGBA,
palette color.Palette,
) (paletted *image.Paletted) {
	if palette == nil {
		palette = paletteRgbCompressed
	}
	paletted = image.NewPaletted(nrgba.Rect, palette)

	draw.FloydSteinberg.Draw(
		paletted,
		paletted.Rect,
		nrgba,
		image.Point{},
	)
	return paletted
}
