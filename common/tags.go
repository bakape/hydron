package common

type TagSource uint8

const (
	User TagSource = iota
	Gelbooru
	Hydrus
)

type TagType uint8

const (
	Undefined TagType = iota
	Author
	Character
	Series
	Rating
)

// Tag of an image
type Tag struct {
	Type   TagType   `json:"type"`
	Source TagSource `json:"source"`
	ID     uint64    `json:"-"`
	Tag    string    `json:"tag"` // Not always defined for performance reasons
}
