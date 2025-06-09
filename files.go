package main

import (
	"fmt"
	"image/gif"
	"os"
	"path/filepath"
)

// gifImg represents a GIF image with its metadata
type gifImg struct {
	decode *gif.GIF
	info   os.FileInfo
	path   string
}

// loadGifsFromDirectory loads all GIF files from specified directories
func loadGifsFromDirectory(args []string) []*gifImg {
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
			obj := loadGifFromPath(filepath.Join(arg, entry.Name()))
			if obj != nil {
				objs = append(objs, obj)
			}
		}
	}
	return objs
}

// loadGifsFromPaths loads GIF files from specified file paths
func loadGifsFromPaths(args []string) []*gifImg {
	var objs []*gifImg
	for _, arg := range args {
		obj := loadGifFromPath(arg)
		if obj != nil {
			objs = append(objs, obj)
		}
	}
	return objs
}

// loadGifFromPath loads a single GIF file from the given path
func loadGifFromPath(path string) *gifImg {
	file, err := os.OpenFile(path, os.O_RDONLY, 0o644)
	if err != nil {
		fmt.Printf("❌ Failed opening '%s': %s\n", path, err.Error())
		return nil
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		fmt.Printf("❌ Failed getting file info '%s': %s\n", path, err.Error())
		return nil
	}

	if info.IsDir() {
		fmt.Printf("❌ '%s' is a directory, use -d flag instead\n", path)
		return nil
	}

	decode, err := gif.DecodeAll(file)
	if err != nil {
		fmt.Printf("❌ Failed decoding '%s': %s\n", path, err.Error())
		return nil
	}

	return &gifImg{
		decode: decode,
		info:   info,
		path:   path,
	}
}
