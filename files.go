package main

import (
	"fmt"
	"github.com/urfave/cli/v3"
	"image"
	"image/gif"
	"math"
	"os"
	"path"
	"path/filepath"
	"sync"
)

func processDirectory(args []string) []*gifImg {
	var objs []*gifImg
	for _, arg := range args {
		entries, err := os.ReadDir(arg)
		if err != nil {
			fmt.Printf("❌ Failed reading '%s': %s\n", arg, err.Error())
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".gif" {
				continue
			}
			obj := readPath(filepath.Join(arg, entry.Name()))
			if obj != nil {
				objs = append(objs, obj)
			}
		}
	}
	return objs
}

func processFiles(args []string) []*gifImg {
	var objs []*gifImg
	for _, arg := range args {
		obj := readPath(filepath.Join(arg))
		if obj != nil {
			objs = append(objs, obj)
		}
	}
	return objs
}

func processGifs(objs []*gifImg, c *cli.Command) {
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
				obj.decode = resizeGifFrames(obj.decode, MaxWidth, MaxHeight)
				if c.Bool("autoplay") {
					obj.decode = compressGif(obj, MaxAutoplaySize)
				} else {
					obj.decode = compressGif(obj, MaxImageSize)
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

// File Handling
type gifImg struct {
	decode *gif.GIF
	file   *os.File
	info   os.FileInfo
}

func readPath(p string) *gifImg {
	file, err := os.OpenFile(p, os.O_RDONLY, 0o644)
	info, _ := file.Stat()
	if info.IsDir() {
		fmt.Printf("❌ '%s' is a directory, use -d flag instead\n", p)
		return nil
	}
	if err != nil {
		fmt.Printf(
			"❌ Failed opening '%s': %s\n",
			p,
			err.Error(),
		)
		return nil
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf(
			"❌ Failed opening '%s': %s\n",
			p,
			err.Error(),
		)
		return nil
	}
	var obj gifImg
	obj.file = file
	obj.info = fileInfo
	decode, err := gif.DecodeAll(obj.file)
	if err != nil {
		fmt.Printf("❌ Faild decoding '%s': %s\n", p, err.Error())
		return nil
	}
	obj.decode = decode
	return &obj
}
