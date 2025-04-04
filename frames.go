package main

import (
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"sync"
)

const (
	MaxWidth        = 1000
	MaxHeight       = 1000
	MaxImageSize    = 5242880
	MaxAutoplaySize = 1048576
)

var defaultPalette = color.Palette{}

// Helper Functions
func initializePalette() {
	defaultPalette = color.Palette{
		color.RGBA{A: 255},                         // Black
		color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White
		color.RGBA{R: 255, A: 255},                 // Red
		color.RGBA{G: 255, A: 255},                 // Green
		color.RGBA{B: 255, A: 255},                 // Blue
		color.RGBA{R: 255, G: 255, A: 255},         // Yellow
		color.RGBA{G: 255, B: 255, A: 255},         // Cyan
		color.RGBA{R: 255, B: 255, A: 255},         // Magenta
	}

	for i := len(defaultPalette); i < 256; i++ {
		defaultPalette = append(
			defaultPalette,
			color.RGBA{R: uint8(i), G: uint8(i), B: uint8(i), A: 255},
		)
	}
}

// isGoodGif checks for the following conditions (AND):
//  1. Height < 1000px
//  3. Width < 1000px
//  2. File Size < 1MB
func isGoodGif(gif *gif.GIF, f *os.File) (good bool, err error) {
	for _, frame := range gif.Image {
		bound := frame.Bounds()
		if bound.Dx() > MaxWidth || bound.Dy() > MaxHeight {
			return false, nil
		}
	}

	stat, err := f.Stat()
	if err != nil {
		return false, err
	}
	if stat.Size() >= MaxImageSize {
		return false, nil
	}
	return true, nil
}

// resizeGifFrames gives a gif with
//   - Same ratio as the original gif
//   - width < x AND height < y
func resizeGifFrames(g *gif.GIF, x int, y int) (new *gif.GIF) {
	new = &gif.GIF{
		Image:           make([]*image.Paletted, len(g.Image)),
		Delay:           g.Delay,
		Disposal:        g.Disposal,
		BackgroundIndex: g.BackgroundIndex,
		Config:          g.Config,
		LoopCount:       g.LoopCount,
	}
	copy(new.Image, g.Image)

	var wg sync.WaitGroup
	for i, frame := range g.Image {
		wg.Add(1)
		go func(i int, frame *image.Paletted) {
			defer wg.Done()
			bound := frame.Bounds()
			if bound.Dx() > x {
				resizedFrame := imaging.Resize(frame, x, 0, imaging.Lanczos)
				new.Image[i] = convertNrgbaPaletted(resizedFrame, frame.Palette)
			}
			if bound.Dy() > y {
				resizedFrame := imaging.Resize(frame, 0, y, imaging.Lanczos)
				new.Image[i] = convertNrgbaPaletted(resizedFrame, frame.Palette)
			}
		}(i, frame)
	}
	wg.Wait()
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

func convertNrgbaPaletted(
nrgba *image.NRGBA,
palette color.Palette,
) (paletted *image.Paletted) {
	if palette == nil {
		paletted = image.NewPaletted(nrgba.Rect, defaultPalette)
	} else {
		paletted = image.NewPaletted(nrgba.Rect, palette)
	}
	draw.FloydSteinberg.Draw(
		paletted,
		paletted.Rect,
		nrgba,
		image.Point{},
	)
	return paletted
}
