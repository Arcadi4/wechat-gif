package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"

	"github.com/disintegration/imaging"
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

func resizeGifFramesUpTo(
	g *gif.GIF,
	targetX int,
	targetY int,
) (resized *gif.GIF) {
	canvaX := g.Config.Width
	canvaY := g.Config.Height
	if canvaX <= targetX && canvaY <= targetY {
		return g
	}

	// Pre-calculate scale ratios to avoid repeated calculations
	var scaleX, scaleY float64
	if canvaX > canvaY {
		scaleX = float64(targetX) / float64(canvaX)
		scaleY = scaleX
	} else {
		scaleY = float64(targetY) / float64(canvaY)
		scaleX = scaleY
	}

	resized = &gif.GIF{
		Image:           make([]*image.Paletted, len(g.Image)),
		Delay:           g.Delay,    // Share slice instead of copying
		Disposal:        g.Disposal, // Share slice instead of copying
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
		LoopCount:       g.LoopCount,
	}

	for i, frame := range g.Image {
		resampleFilter := imaging.Lanczos
		bound := frame.Bounds()
		if bound.Min.X != 0 || bound.Min.Y != 0 || bound.Max.X != canvaX || bound.Max.Y != canvaY {
			// If the frame is not aligned to the gif canvas, we need to
			// draw it on a new canvas before resizing. This might increase
			// the file size, but it is necessary to ensure that the gif is
			// correctly displayed after resizing.
			backgroundBound := image.Rect(0, 0, canvaX, canvaY)
			frame = addTransparentBackground(backgroundBound, frame)
			// NearestNeighbor preserves harsh edges, this would prevent
			// glitched frames (as a transparent background is added).
			resampleFilter = imaging.NearestNeighbor
		}

		// Resize frame efficiently
		var resizedFrame *image.NRGBA
		if canvaX > canvaY {
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

	// For performance, use direct color mapping instead of Floyd-Steinberg
	// This is faster but may reduce quality slightly
	bounds := nrgba.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			paletted.Set(x, y, nrgba.At(x, y))
		}
	}
	return paletted
}
