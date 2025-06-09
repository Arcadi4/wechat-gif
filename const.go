package main

import "image/color"

// WeChat constraints for sending GIFs as images
const (
	maxWidth        = 1000    // Maximum width in pixels
	maxHeight       = 1000    // Maximum height in pixels
	maxImageSize    = 5242880 // Maximum file size (5MiB) for sending as image
	maxAutoplaySize = 1048576 // Maximum file size (1MiB) for autoplay
)

// Pre-computed compressed color palette for better performance
var paletteRgbCompressed = color.Palette{}

func init() {
	// Generate a compressed RGB palette (32x32x32 = 32768 colors)
	for r := 0; r < 32; r++ {
		for g := 0; g < 32; g++ {
			for b := 0; b < 32; b++ {
				paletteRgbCompressed = append(
					paletteRgbCompressed,
					color.RGBA{
						R: uint8(r * 8),
						G: uint8(g * 8),
						B: uint8(b * 8),
					},
				)
			}
		}
	}
}
