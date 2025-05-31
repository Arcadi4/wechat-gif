package main

import (
	"fmt"
	"image/gif"
	"os"
	"path/filepath"
)

type gifImg struct {
	decode *gif.GIF
	file   *os.File
	info   os.FileInfo
	path   string
}

func flagDirectory(args []string) []*gifImg {
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

func flagFiles(args []string) []*gifImg {
	var objs []*gifImg
	for _, arg := range args {
		obj := readPath(filepath.Join(arg))
		if obj != nil {
			objs = append(objs, obj)
		}
	}
	return objs
}

func readPath(p string) *gifImg {
	file, err := os.OpenFile(p, os.O_RDONLY, 0o644)
	if err != nil {
		fmt.Printf(
			"❌ Failed opening '%s': %s\n",
			p,
			err.Error(),
		)
		return nil
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		fmt.Printf(
			"❌ Failed getting file info '%s': %s\n",
			p,
			err.Error(),
		)
		return nil
	}

	if info.IsDir() {
		file.Close()
		fmt.Printf("❌ '%s' is a directory, use -d flag instead\n", p)
		return nil
	}

	decode, err := gif.DecodeAll(file)
	if err != nil {
		file.Close()
		fmt.Printf("❌ Failed decoding '%s': %s\n", p, err.Error())
		return nil
	}

	// Close file immediately after decoding since we have all data in memory
	file.Close()

	return &gifImg{
		decode: decode,
		file:   nil, // We no longer need the file handle
		info:   info,
		path:   p, // Store path for later use
	}
}
