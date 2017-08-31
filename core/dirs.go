package core

import (
	"os"
	"path/filepath"
	"runtime"
)

const hexStr = "0123456789abcdef"

// Root directory path
var rootPath, ImageRoot, ThumbRoot string

// Determine root dirs
func init() {
	if runtime.GOOS == "windows" {
		rootPath = filepath.Join(os.Getenv("APPDATA"), "hydron")
	} else {
		rootPath = filepath.Join(os.Getenv("HOME"), ".hydron")
	}
	sep := string(filepath.Separator)
	ImageRoot = concatStrings(rootPath, sep, "images", sep)
	ThumbRoot = concatStrings(rootPath, sep, "thumbs", sep)
}

func initDirs() error {
	if _, err := os.Stat(rootPath); !os.IsNotExist(err) {
		return err
	}

	// Create source file and thumbnail directories
	const dirMode = os.ModeDir | 0700
	for _, dir := range [...]string{ImageRoot, ThumbRoot} {
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

// traverse recursively traverses an array of file and/or directory paths
func traverse(paths []string) (files []string, err error) {
	files = make([]string, 0, 64)

	visit := func(path string, info os.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case !info.IsDir():
			files = append(files, path)
		}
		return nil
	}

	for _, p := range paths {
		// Don't walk network paths
		if IsFetchable(p) {
			files = append(files, p)
			continue
		}

		err = filepath.Walk(p, visit)
		if err != nil {
			return
		}
	}
	return
}

// Returns the path to the source file
func SourcePath(id string, typ FileType) string {
	name := concatStrings(id, ".", Extensions[typ])
	return filepath.Join(ImageRoot, id[:2], name)
}

// Returns the path to the thumbnail
func ThumbPath(id string, isPNG bool) string {
	var ext string
	if isPNG {
		ext = "png"
	} else {
		ext = "jpg"
	}
	return filepath.Join(ThumbRoot, id[:2], concatStrings(id, ".", ext))
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

// Write a file to disk. If a file already exists, because of an interrupted
// write or something, overwrite it.
func writeFile(path string, buf []byte) (err error) {
	const flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	f, err := os.OpenFile(path, flags, 0660)
	switch {
	case err == nil:
	case os.IsExist(err):
		err = os.Remove(path)
		if err != nil {
			return
		}
		f, err = os.OpenFile(path, flags, 0660)
		if err != nil {
			return
		}
	default:
		return
	}
	defer f.Close()

	_, err = f.Write(buf)
	return
}
