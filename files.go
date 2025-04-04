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
