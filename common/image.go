//go:generate easyjson --all $GOFILE

package common

// Record of an image stored in the database
type Image struct {
	Type FileType `json:"type"`
	CompactImage
	Dims
	ID         int64  `json:"-"`
	ImportTime int64  `json:"importTime"`
	Size       int    `json:"size"`
	Duration   uint64 `json:"duration,omitempty"`
	MD5        string `json:"md5"`
	// Not always defined for performance reasons
	Tags []Tag `json:"tags,omitempty"`
}

// Common fields
type Dims struct {
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}

// Thumbnail of image
type Thumbnail struct {
	IsPNG bool `json:"is_png"`
	Dims
}

// Only provides the most minimal of fields. Optimal for thumbnail views.
type CompactImage struct {
	SHA1  string    `json:"sha1"`
	Thumb Thumbnail `json:"thumb"`
}
