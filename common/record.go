//go:generate easyjson --all $GOFILE

package common

// Record of an image stored in the database
type Record struct {
	Type FileType `json:"type"`
	CompactRecord
	Image
	ID         uint64 `json:"-"`
	ImportTime uint64 `json:"importTime"`
	Size       uint64 `json:"size"`
	Duration   uint64 `json:"duration,omitempty"`
	MD5        string `json:"md5"`
	// Not always defined for performance reasons
	Tags []Tag `json:"tags,omitempty"`
}

// Common fields
type Image struct {
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}

// Thumbnail of image
type Thumbnail struct {
	IsPNG bool `json:"is_png"`
	Image
}

// Only provides the most minimal of fields. Optimal for thumbnail views.
type CompactRecord struct {
	SHA1  string    `json:"sha1"`
	Thumb Thumbnail `json:"thumb"`
}
