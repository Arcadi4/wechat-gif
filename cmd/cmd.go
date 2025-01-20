package cmd

import (
	"context"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/urfave/cli/v3"
	"image"
	"image/color"
	"image/draw"
	stlgif "image/gif"
	"log"
	"os"
	"path"
	"path/filepath"
)

const maxX = 1000
const maxY = 1000
const maxFileSize = 1000000

var defaultPalette = color.Palette{}

var Cmd = &cli.Command{
	Name:  "wechat-gif",
	Usage: "Compress gif so it can be sent via WeChat",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "dir",
			Usage:   "Process all gif files in given directory",
			Aliases: []string{"d"},
		},
	},
	Action: action,
}

func action(context.Context, *cli.Command) (err error) {
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

	args := os.Args[1:]

	if len(args) == 0 {
		log.Println("Specify gifDecode file(s) to process it")
	}

	gifs, files := readArgs(args)

	for i, gifDecode := range gifs {
		if gifDecode == nil {
			continue
		}

		good, err := isGoodGif(gifDecode, files[i])
		if err != nil {
			fmt.Printf("❌ Failed checking '%s': %s\n", args[i], err.Error())
			continue
		}
		if !good {
			resized := resizeGifFrames(gifDecode, maxX, maxY)
			outPath := filepath.Join(
				path.Dir(files[i].Name()),
				"WeChat_"+path.Base(files[i].Name()),
			)
			out, err := os.Create(outPath)
			err = stlgif.EncodeAll(out, resized)
			if err != nil {
				fmt.Printf("❌ Failed saving '%s': %s\n", args[i], err.Error())
				continue
			}
			fmt.Printf("🟢 Saved resized image '%s'\n", outPath)
		} else {
			fmt.Printf("🟢 '%s' is already good\n", args[i])
		}
	}

	return nil
}

func readArgs(args []string) ([]*stlgif.GIF, []*os.File) {
	// The indices of args, paths, files, and gifs align with each others.
	var paths []string
	for _, arg := range args {
		path, err := filepath.Abs(arg)
		if err != nil {
			fmt.Printf("❌ Invalid path '%s'\n", arg)
			paths = append(paths, "")
			continue
		}
		paths = append(paths, path)
	}

	var files []*os.File
	for i, path := range paths {
		if path == "" {
			files = append(files, nil)
			continue
		}

		file, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			fmt.Printf(
				"❌ Failed opening '%s': %s\n",
				args[i],
				err.Error(),
			)
			files = append(files, nil)
			continue
		}
		files = append(files, file)
	}

	var gifs []*stlgif.GIF
	for i, file := range files {
		if file == nil {
			gifs = append(gifs, nil)
			continue
		}

		gifDecode, err := stlgif.DecodeAll(file)
		if err != nil {
			fmt.Printf("❌ Faild decoding '%s': %s", args[i], err.Error())
			gifs = append(gifs, nil)
			continue
		}
		gifs = append(gifs, gifDecode)
	}

	return gifs, files
}

// isGoodGif checks for the following conditions (AND):
//  1. Height < 1000px
//  3. Width < 1000px
//  2. File Size < 1MB
func isGoodGif(gif *stlgif.GIF, f *os.File) (good bool, err error) {
	for _, frame := range gif.Image {
		bound := frame.Bounds()
		if bound.Dx() > maxX || bound.Dy() > maxY {
			return false, nil
		}
	}

	stat, err := f.Stat()
	if err != nil {
		return false, err
	}
	if stat.Size() >= maxFileSize {
		return false, nil
	}

	return true, nil
}

// resizeGifFrames gives a gif with
//   - Same ratio as the original gif
//   - width < x AND height < y
func resizeGifFrames(gif *stlgif.GIF, x int, y int) (new *stlgif.GIF) {
	new = &stlgif.GIF{
		Image:           make([]*image.Paletted, len(gif.Image)),
		Delay:           gif.Delay,
		Disposal:        gif.Disposal,
		BackgroundIndex: gif.BackgroundIndex,
		Config:          gif.Config,
		LoopCount:       gif.LoopCount,
	}
	copy(new.Image, gif.Image)

	largestX := 0
	largestY := 0

	for i, frame := range gif.Image {
		bound := frame.Bounds()
		if bound.Dx() > x {
			resizedFrame := imaging.Resize(frame, x, 0, imaging.Lanczos)
			new.Image[i] = convertNrgbaPaletted(resizedFrame, frame.Palette)
		}
		if bound.Dy() > y {
			resizedFrame := imaging.Resize(frame, 0, y, imaging.Lanczos)
			new.Image[i] = convertNrgbaPaletted(resizedFrame, frame.Palette)
		}
		if new.Image[i].Bounds().Dx() > largestX {
			largestX = new.Image[i].Bounds().Dx()
		}
		if new.Image[i].Bounds().Dy() > largestY {
			largestY = new.Image[i].Bounds().Dy()
		}
	}

	new.Config.Width = largestX
	new.Config.Height = largestY

	return new
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
