package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"os"
	"sync"

	"github.com/disintegration/imaging"
)

const (
	maxWidth        = 1000
	maxHeight       = 1000
	maxImageSize    = 5242880
	maxAutoplaySize = 1048576
)

var rgb24Palette = color.Palette{}

func init() {
	for r := range 256 {
		for g := range 256 {
			for b := range 256 {
				rgb24Palette = append(rgb24Palette, color.RGBA{uint8(r), uint8(g), uint8(b), 0})
			}
		}
	}
}

// isGoodGif checks for the following conditions (AND):
//  1. Height < 1000px
//  3. Width < 1000px
//  2. File Size < 1MB
func isGoodGif(gif *gif.GIF, f *os.File) (good bool, err error) {
	for _, frame := range gif.Image {
		bound := frame.Bounds()
		if bound.Dx() > maxWidth || bound.Dy() > maxHeight {
			return false, nil
		}
	}

	stat, err := f.Stat()
	if err != nil {
		return false, err
	}
	if stat.Size() >= maxImageSize {
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
				new.Image[i] = convertNrgbaToPaletted(resizedFrame, frame.Palette)
			}
			if bound.Dy() > y {
				resizedFrame := imaging.Resize(frame, 0, y, imaging.Lanczos)
				new.Image[i] = convertNrgbaToPaletted(resizedFrame, frame.Palette)
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

func convertNrgbaToPaletted(
	nrgba *image.NRGBA,
	palette color.Palette,
) (paletted *image.Paletted) {
	if palette == nil {
		paletted = image.NewPaletted(nrgba.Rect, rgb24Palette)
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
