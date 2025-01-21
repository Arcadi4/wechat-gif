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
	"math"
	"os"
	"path"
	"path/filepath"
	"sync"
)

const MaxWidth = 1000
const MaxHeight = 1000
const MaxImageSize = 5242880
const MaxAutoplaySize = 1048576

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
		&cli.BoolFlag{
			Name:    "autoplay",
			Usage:   "Compress gif so it can autoplay",
			Aliases: []string{"a"},
		},
	},
	Action: action,
}

func action(ctx context.Context, c *cli.Command) (err error) {
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

	args := c.Args().Slice()
	if len(args) == 0 {
		cli.ShowAppHelpAndExit(c, 0)
	}

	var objs []*gifImg
	if c.Bool("dir") {
		var paths []string
		for _, arg := range args {
			entries, err := os.ReadDir(arg)
			if err != nil {
				fmt.Printf(
					"❌ Failed reading '%s': %s\n",
					arg,
					err.Error(),
				)
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() || filepath.Ext(entry.Name()) != ".gif" {
					continue
				}
				paths = append(paths, filepath.Join(arg, entry.Name()))
			}
		}
		objs = readArgs(paths)
	} else {
		objs = readArgs(args)
	}

	for _, obj := range objs {
		good, err := isGoodGif(obj.decode, obj.file)
		if err != nil {
			fmt.Printf(
				"❌ Failed checking '%s': %s\n",
				obj.file.Name(),
				err.Error(),
			)
			continue
		}
		if !good {
			obj.decode = resizeGifFrames(obj.decode, MaxWidth, MaxHeight)
			if c.Bool("autoplay") {
				obj.decode = compressGif(obj, MaxAutoplaySize)
			} else {
				obj.decode = compressGif(obj, MaxImageSize)
			}
			outPath := filepath.Join(
				path.Dir(obj.file.Name()),
				"WeChat_"+path.Base(obj.file.Name()),
			)
			out, err := os.Create(outPath)
			err = stlgif.EncodeAll(out, obj.decode)
			if err != nil {
				fmt.Printf(
					"❌ Failed saving '%s': %s\n",
					obj.file.Name(),
					err.Error(),
				)
				continue
			}
			fmt.Printf("🟢 Saved resized image '%s'\n", path.Base(outPath))
		} else {
			fmt.Printf("🟢 '%s' is already good\n", obj.file.Name())
		}
	}

	return nil
}

type gifImg struct {
	decode *stlgif.GIF
	file   *os.File
	size   int
}

func readArgs(paths []string) []*gifImg {
	var objs []*gifImg
	for _, p := range paths {
		file, err := os.OpenFile(p, os.O_RDONLY, 0644)
		if err != nil {
			fmt.Printf(
				"❌ Failed opening '%s': %s\n",
				p,
				err.Error(),
			)
			continue
		}
		stat, err := file.Stat()
		if err != nil {
			fmt.Printf(
				"❌ Failed opening '%s': %s\n",
				p,
				err.Error(),
			)
			continue
		}
		var obj gifImg
		obj.file = file
		obj.size = int(stat.Size())
		decode, err := stlgif.DecodeAll(obj.file)
		if err != nil {
			fmt.Printf("❌ Faild decoding '%s': %s", p, err.Error())
			continue
		}
		obj.decode = decode
		objs = append(objs, &obj)
	}
	return objs
}

// isGoodGif checks for the following conditions (AND):
//  1. Height < 1000px
//  3. Width < 1000px
//  2. File Size < 1MB
func isGoodGif(gif *stlgif.GIF, f *os.File) (good bool, err error) {
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

	var wg sync.WaitGroup
	for i, frame := range gif.Image {
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

func updateGifConfig(gif *stlgif.GIF) {
	largestX := 0
	largestY := 0
	for _, frame := range gif.Image {
		bound := frame.Bounds()
		if bound.Dx() > largestX {
			largestX = bound.Dx()
		}
		if bound.Dy() > largestY {
			largestY = bound.Dy()
		}
	}
	gif.Config.Width = largestX
	gif.Config.Height = largestY
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

func compressGif(gif *gifImg, maxSize int) (new *stlgif.GIF) {
	decode := gif.decode
	// Stretching one edge of the gif by factor x will expand the size by roughly
	// x^2 times. Note that it's the same when x < 1.0. So we can use the square
	// root of the size ratio to scale the gif to estimate the ratio to resize
	// one edge of the gif. Then we minus the ratio by 0.02 for safety in edge
	// cases.
	rate := math.Sqrt(float64(maxSize)/float64(gif.size)) - 0.02
	// Rate > 1.0 indicates that the gif is already smaller than the target size.
	// So we return the original gif directly.
	if rate > 1.0 {
		return decode
	}

	new = &stlgif.GIF{
		Image:           make([]*image.Paletted, len(decode.Image)),
		Delay:           decode.Delay,
		Disposal:        decode.Disposal,
		BackgroundIndex: decode.BackgroundIndex,
		Config:          decode.Config,
		LoopCount:       decode.LoopCount,
	}
	copy(new.Image, decode.Image)

	return resizeGifFrames(
		new,
		int(float64(decode.Config.Width)*rate),
		decode.Config.Height,
	)
}
