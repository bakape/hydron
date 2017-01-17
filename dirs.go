package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const hexStr = "0123456789abcdef"

var (
	// maps internal file types to their canonical file extensions
	extensions = map[fileType]string{
		JPEG: "jpg",
		PNG:  "png",
		GIF:  "gif",
		MP4:  "mp4",
		WEBM: "webm",
		OGG:  "ogg",
		PDF:  "pdf",
	}

	rootPath, imageRoot, thumbRoot string
)

// Determine root dirs
func init() {
	var key string
	if runtime.GOOS == "windows" {
		key = "HOMEPATH"
	} else {
		key = "HOME"
	}
	rootPath = filepath.Join(os.Getenv(key), ".hydron")
	sep := string(os.PathSeparator)
	imageRoot = rootPath + sep + "images" + sep
	thumbRoot = rootPath + sep + "thumbs" + sep
}

func initDirs() error {
	stderr.Printf("initializing root directory %s\n", rootPath)
	if _, err := os.Stat(rootPath); !os.IsNotExist(err) {
		return err
	}

	// Create source file and thumbnail directories
	const dirMode = os.ModeDir | 0700
	for _, dir := range [...]string{imageRoot, thumbRoot} {
		err := os.MkdirAll(dir, dirMode)
		if err != nil {
			return err
		}
		for _, ch1 := range hexStr {
			for _, ch2 := range hexStr {
				d := string([]rune{ch1, ch2})
				err := os.Mkdir(filepath.Join(dir, d), dirMode)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Recursively traverse array of files and/or directories
func traverse(paths []string) (files []string, err error) {
	i := 0
	defer stderr.Print("\n")
	for _, p := range paths {
		err = filepath.Walk(p, func(
			path string,
			info os.FileInfo,
			err error,
		) error {
			switch {
			case err != nil:
				return err
			case !info.IsDir():
				files = append(files, path)
				i++
				fmt.Fprintf(os.Stderr, "\rgathering files: %d", i)
			}
			return nil
		})
		if err != nil {
			return
		}
	}
	return
}

func sourcePath(id string, typ fileType) string {
	return concatStrings(imageRoot, id[:2], "/", id, ".", extensions[typ])
}

func thumbPath(id string, isPNG bool) string {
	var ext string
	if isPNG {
		ext = "png"
	} else {
		ext = "jpg"
	}
	return concatStrings(thumbRoot, id[:2], "/", id, ".", ext)
}

func concatStrings(s ...string) string {
	l := 0
	for _, s := range s {
		l += len(s)
	}

	b := make([]byte, 0, l)
	for _, s := range s {
		b = append(b, s...)
	}

	return string(b)
}
