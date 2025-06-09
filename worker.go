package main

import (
	"bufio"
	"fmt"
	"image/gif"
	"os"
	"path"
	"path/filepath"
	"sync"
)

// processGifs processes multiple GIF files concurrently
func processGifs(gifs []*gifImg, autoplay bool, maxWorkers int) {
	if len(gifs) == 0 {
		return
	}

	// Limit concurrent processing
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	if len(gifs) < maxWorkers {
		maxWorkers = len(gifs)
	}

	// Use semaphore to control concurrency
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, gif := range gifs {
		wg.Add(1)
		go func(gif *gifImg) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			processGif(gif, autoplay)
		}(gif)
	}

	wg.Wait()
}

// processGif processes a single GIF file
func processGif(gif *gifImg, autoplay bool) {
	// Check if GIF already meets requirements
	if isGoodGif(gif.decode, gif.info) {
		fmt.Printf("🟢 '%s' is already good\n", gif.path)
		return
	}

	// Determine target size based on autoplay requirement
	var maxSize int
	if autoplay {
		maxSize = maxAutoplaySize
	} else {
		maxSize = maxImageSize
	}

	// Compress and resize the GIF
	gif.decode = compressGif(gif, maxSize)
	gif.decode = resizeGif(gif.decode, maxWidth, maxHeight)

	// Save the processed GIF
	saveGif(gif)
}

// saveGif saves the processed GIF to a new file
func saveGif(img *gifImg) {
	// Generate output path with "WeChat_" prefix
	outPath := filepath.Join(
		path.Dir(img.path),
		"WeChat_"+path.Base(img.path),
	)

	// Create output file
	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("❌ Failed creating output file: %s\n", err.Error())
		return
	}
	defer out.Close()

	// Use buffered writer for better I/O performance
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	// Encode and save the GIF
	err = gif.EncodeAll(writer, img.decode)
	if err != nil {
		fmt.Printf("❌ Failed saving '%s': %s\n", img.path, err.Error())
		return
	}

	fmt.Printf("🟢 Saved resized image '%s'\n", path.Base(outPath))
}
