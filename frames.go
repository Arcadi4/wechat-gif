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
func resizeGifFrames(g *gif.GIF, targetX int, targetY int) (resized *gif.GIF) {
	resized = &gif.GIF{
		Image:           make([]*image.Paletted, len(g.Image)),
		Delay:           g.Delay,
		Disposal:        g.Disposal,
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
		LoopCount:       g.LoopCount,
	}
	copy(resized.Image, g.Image)
	sizeX := resized.Config.Width
	sizeY := resized.Config.Height

	for i, frame := range g.Image {
		resampleFilter := imaging.Lanczos
		bound := frame.Bounds()
		if bound.Min.X != 0 || bound.Min.Y != 0 || bound.Max.X != sizeX || bound.Max.Y != sizeY {
			// If the frame is not aligned to the gif canvas, we need to
			// draw it on a new canvas before resizing. This might increase
			// the file size, but it is necessary to ensure that the gif is
			// correctly displayed after resizing.
			backgroundBound := image.Rect(0, 0, sizeX, sizeY)
			frame = addTransparentBackground(backgroundBound, frame)
			resampleFilter = imaging.NearestNeighbor
		}

		var resizedFrame *image.NRGBA
		if sizeX > sizeY {
			resizedFrame = imaging.Resize(frame, targetX, 0, resampleFilter)
		} else {
			resizedFrame = imaging.Resize(frame, 0, targetY, resampleFilter)
		}

		resized.Image[i] = nrgbaToPaletted(resizedFrame, frame.Palette)
	}

	updateGifConfig(resized)
	return resized
}

func addTransparentBackground(
rect image.Rectangle,
paletted *image.Paletted,
) *image.Paletted {
	background := image.NewPaletted(rect, paletted.Palette)
	draw.Draw(
		background,
		background.Bounds(),
		image.Transparent,
		image.Point{},
		draw.Src,
	)
	draw.Draw(background, background.Rect, paletted, image.Point{}, draw.Over)
	return background
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
