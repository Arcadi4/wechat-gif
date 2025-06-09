package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"

	"github.com/disintegration/imaging"
)

// isGoodGif checks if a GIF meets WeChat's requirements:
// 1. Width and height both < 1000px
// 2. File size < maxImageSize
func isGoodGif(g *gif.GIF, fileInfo os.FileInfo) bool {
	// Check dimensions
	for _, frame := range g.Image {
		bound := frame.Bounds()
		if bound.Dx() > maxWidth || bound.Dy() > maxHeight {
			return false
		}
	}

	// Check file size
	return fileInfo.Size() < maxImageSize
}

// compressGif compresses a GIF to fit within the specified size limit
func compressGif(img *gifImg, maxSize int) *gif.GIF {
	decode := img.decode

	// Early return if already small enough
	if img.info.Size() <= int64(maxSize) {
		return decode
	}

	// Calculate compression ratio
	sizeRatio := float64(maxSize) / float64(img.info.Size())
	areaRatio := math.Sqrt(sizeRatio) * 0.8 // Conservative approach

	// Ensure minimum quality threshold
	if areaRatio < 0.3 {
		areaRatio = 0.3
	}

	// Calculate target dimensions
	targetX := int(float64(decode.Config.Width) * areaRatio)
	targetY := int(float64(decode.Config.Height) * areaRatio)

	// Ensure minimum dimensions
	if targetX < 64 {
		targetX = 64
	}
	if targetY < 64 {
		targetY = 64
	}

	return resizeGif(decode, targetX, targetY)
}

// resizeGif resizes a GIF to the specified maximum dimensions
func resizeGif(g *gif.GIF, targetX, targetY int) *gif.GIF {
	canvaX, canvaY := g.Config.Width, g.Config.Height

	// No need to resize if already within limits
	if canvaX <= targetX && canvaY <= targetY {
		return g
	}

	// Create new GIF structure
	resized := &gif.GIF{
		Image:           make([]*image.Paletted, len(g.Image)),
		Delay:           g.Delay,
		Disposal:        g.Disposal,
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
		LoopCount:       g.LoopCount,
	}

	// Process each frame
	for i, frame := range g.Image {
		resized.Image[i] = resizeFrame(frame, canvaX, canvaY, targetX, targetY)
	}

	updateGifConfig(resized)
	return resized
}

// resizeFrame resizes a single frame of the GIF
func resizeFrame(frame *image.Paletted, canvaX, canvaY, targetX, targetY int) *image.Paletted {
	resampleFilter := imaging.Lanczos
	bound := frame.Bounds()

	// Check if frame needs background alignment
	if bound.Min.X != 0 || bound.Min.Y != 0 || bound.Max.X != canvaX || bound.Max.Y != canvaY {
		// Add transparent background to align frame to canvas
		backgroundBound := image.Rect(0, 0, canvaX, canvaY)
		frame = addTransparentBackground(backgroundBound, frame)
		// Use NearestNeighbor to preserve edges after background addition
		resampleFilter = imaging.NearestNeighbor
	}

	// Resize frame maintaining aspect ratio
	var resizedFrame *image.NRGBA
	if canvaX > canvaY {
		resizedFrame = imaging.Resize(frame, targetX, 0, resampleFilter)
	} else {
		resizedFrame = imaging.Resize(frame, 0, targetY, resampleFilter)
	}

	return nrgbaToPaletted(resizedFrame, frame.Palette)
}

// addTransparentBackground adds a transparent background to align the frame
func addTransparentBackground(rect image.Rectangle, paletted *image.Paletted) *image.Paletted {
	background := image.NewPaletted(rect, paletted.Palette)

	// Fill with transparent background
	draw.Draw(background, background.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Draw the original frame on top
	draw.Draw(background, background.Rect, paletted, image.Point{}, draw.Over)

	return background
}

// updateGifConfig updates the GIF configuration to match the largest frame
func updateGifConfig(g *gif.GIF) {
	largestWidth, largestHeight := 0, 0

	for _, frame := range g.Image {
		bound := frame.Bounds()
		if bound.Dx() > largestWidth {
			largestWidth = bound.Dx()
		}
		if bound.Dy() > largestHeight {
			largestHeight = bound.Dy()
		}
	}

	// Update canvas size to fit the largest frame
	g.Config.Width = largestWidth
	g.Config.Height = largestHeight
}

// nrgbaToPaletted converts NRGBA image to paletted image
func nrgbaToPaletted(nrgba *image.NRGBA, palette color.Palette) *image.Paletted {
	if palette == nil {
		palette = paletteRgbCompressed
	}

	paletted := image.NewPaletted(nrgba.Rect, palette)
	bounds := nrgba.Bounds()

	// Direct color mapping for better performance
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			paletted.Set(x, y, nrgba.At(x, y))
		}
	}

	return paletted
}
