package main

import (
	"fmt"
	"image"
	"image/gif"
	"math"
	"os"
	"path"
	"path/filepath"
	"sync"
)

func processGifs(objs []*gifImg, autoplay bool) {
	wg := sync.WaitGroup{}
	for _, obj := range objs {
		wg.Add(1)
		go func(obj *gifImg) {
			defer wg.Done()
			good, err := isGoodGif(obj.decode, obj.file)
			if err != nil {
				fmt.Printf(
					"❌ Failed checking '%s': %s\n",
					obj.file.Name(),
					err.Error(),
				)
				return
			}
			if !good {
				obj.decode = resizeGifFrames(obj.decode, maxWidth, maxHeight)
				if autoplay {
					obj.decode = compressGif(obj, maxAutoplaySize)
				} else {
					obj.decode = compressGif(obj, maxImageSize)
				}
				saveGif(obj)
			} else {
				fmt.Printf("🟢 '%s' is already good\n", obj.file.Name())
			}
		}(obj)
	}
	wg.Wait()
}

func saveGif(obj *gifImg) {
	outPath := filepath.Join(
		path.Dir(obj.file.Name()),
		"WeChat_"+path.Base(obj.file.Name()),
	)
	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("❌ Failed creating output file: %s\n", err.Error())
		return
	}
	defer out.Close()

	err = gif.EncodeAll(out, obj.decode)
	if err != nil {
		fmt.Printf("❌ Failed saving '%s': %s\n", obj.file.Name(), err.Error())
		return
	}
	fmt.Printf("🟢 Saved resized image '%s'\n", path.Base(outPath))
}

func compressGif(gifImg *gifImg, maxSize int) (new *gif.GIF) {
	decode := gifImg.decode
	// Stretching one edge of the gif by factor x will expand the size by roughly
	// x^2 times. Note that it's the same when x < 1.0. So we can use the square
	// root of the size ratio to scale the gif to estimate the ratio to resize
	// one edge of the gif. Then we minus the ratio by 0.02 for safety in edge
	// cases.
	rate := math.Sqrt(float64(maxSize)/float64(gifImg.info.Size())) - 0.02
	// Rate > 1.0 indicates that the gif is already smaller than the target size.
	// So we return the original gif directly.
	if rate > 1.0 {
		return decode
	}

	new = &gif.GIF{
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
