package files

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bakape/hydron/v3/common"
	"github.com/bakape/hydron/v3/util"
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
	ImageRoot = filepath.Join(RootPath, "images")
	ThumbRoot = filepath.Join(RootPath, "thumbs")
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

	var home string
	for _, p := range paths {
		// Don't walk network paths
		if util.IsFetchable(p) {
			files = append(files, p)
			continue
		}

		// Expand home directory token
		if strings.HasPrefix(p, "~/") {
			if home == "" {
				var u *user.User
				u, err = user.Current()
				if err != nil {
					return
				}
				home = u.HomeDir
			}
			p = filepath.Join(home, p[2:])
		}

		p, err = filepath.Abs(p)
		if err != nil {
			return
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
	return filepath.Join(ImageRoot, id[:2],
		fmt.Sprintf("%s.%s", id, common.Extensions[typ]))
}

// Returns the path to the thumbnail
func ThumbPath(id string) string {
	return filepath.Join(ThumbRoot, id[:2], id+".webp")
}

// Net URL to thumbnail path
func NetThumbPath(id string) string {
	return fmt.Sprintf("/thumbs/%s.webp", id)
}

// Net URL to source file path
func NetSourcePath(id string, typ common.FileType) string {
	return fmt.Sprintf("/files/%s.%s", id, common.Extensions[typ])
}
