package main

import (
	"bufio"
	"fmt"
	"image/gif"
	"math"
	"os"
	"path"
	"path/filepath"
	"sync"
)

func processGifs(objs []*gifImg, autoplay bool, maxWorkers int) {
	// Limit concurrent processing to avoid overwhelming the system
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	if len(objs) < maxWorkers {
		maxWorkers = len(objs)
	}

	semaphore := make(chan struct{}, maxWorkers)
	wg := sync.WaitGroup{}

	for _, obj := range objs {
		wg.Add(1)
		go func(obj *gifImg) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			good, err := isGoodGif(obj.decode, obj.info)
			if err != nil {
				fmt.Printf(
					"❌ Failed checking '%s': %s\n",
					obj.path,
					err.Error(),
				)
				return
			}
			if !good {
				if autoplay {
					obj.decode = compressGif(obj, maxAutoplaySize)
				} else {
					obj.decode = compressGif(obj, maxImageSize)
				}

				obj.decode = resizeGifFramesUpTo(
					obj.decode,
					maxWidth,
					maxHeight,
				)
				saveGif(obj)
			} else {
				fmt.Printf("🟢 '%s' is already good\n", obj.path)
			}
		}(obj)
	}
	wg.Wait()
}

func saveGif(obj *gifImg) {
	outPath := filepath.Join(
		path.Dir(obj.path),
		"WeChat_"+path.Base(obj.path),
	)
	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("❌ Failed creating output file: %s\n", err.Error())
		return
	}
	defer out.Close()

	// Use buffered writer for better I/O performance
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	err = gif.EncodeAll(writer, obj.decode)
	if err != nil {
		fmt.Printf("❌ Failed saving '%s': %s\n", obj.path, err.Error())
		return
	}
	fmt.Printf("🟢 Saved resized image '%s'\n", path.Base(outPath))
}

func compressGif(img *gifImg, maxSize int) (new *gif.GIF) {
	decode := img.decode

	// Early return if already small enough
	size := img.info.Size()
	if size <= int64(maxSize) {
		return decode
	}

	// Improved compression ratio calculation
	// Using a more conservative approach to ensure we don't over-compress
	sizeRatio := float64(maxSize) / float64(size)
	areaRatio := math.Sqrt(sizeRatio) * 0.8

	// Ensure minimum quality threshold
	if areaRatio < 0.3 {
		areaRatio = 0.3
	}

	targetX := int(float64(decode.Config.Width) * areaRatio)
	targetY := int(float64(decode.Config.Height) * areaRatio)

	// Ensure minimum dimensions
	if targetX < 64 {
		targetX = 64
	}
	if targetY < 64 {
		targetY = 64
	}

	return resizeGifFramesUpTo(decode, targetX, targetY)
}

// isGoodGif checks for the following conditions (AND):
//  1. Height < 1000px
//  3. Width < 1000px
//  2. File Size < maxImageSize
func isGoodGif(gif *gif.GIF, fileInfo os.FileInfo) (good bool, err error) {
	for _, frame := range gif.Image {
		bound := frame.Bounds()
		if bound.Dx() > maxWidth || bound.Dy() > maxHeight {
			return false, nil
		}
	}

	if fileInfo.Size() >= maxImageSize {
		return false, nil
	}
	return true, nil
}
