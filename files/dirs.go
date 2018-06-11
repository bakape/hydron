package files

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/bakape/hydron/common"
	"github.com/bakape/hydron/util"
)

// Root directory paths
var RootPath, ImageRoot, ThumbRoot string

// Determine root dirs
func init() {
	if runtime.GOOS == "windows" {
		RootPath = filepath.Join(os.Getenv("APPDATA"), "hydron")
	} else {
		RootPath = filepath.Join(os.Getenv("HOME"), ".hydron")
	}
	sep := string(filepath.Separator)
	ImageRoot = concatStrings(RootPath, sep, "images", sep)
	ThumbRoot = concatStrings(RootPath, sep, "thumbs", sep)
}

func Init() error {
	if _, err := os.Stat(RootPath); !os.IsNotExist(err) {
		return err
	}

	const hexStr = "0123456789abcdef"

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

// Recursively traverses an array of file and/or directory paths
func Traverse(paths []string) (files []string, err error) {
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
		if util.IsFetchable(p) {
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
func SourcePath(id string, typ common.FileType) string {
	name := concatStrings(id, ".", common.Extensions[typ])
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
