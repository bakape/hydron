//go:generate easyjson --all $GOFILE

package common

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
)

// Record of an image stored in the database
type Image struct {
	CompactImage
	Dims
	ID         int64  `json:"-"`
	ImportTime int64  `json:"import_time"`
	Size       int    `json:"size"`
	Duration   uint64 `json:"duration,omitempty"`
	MD5        string `json:"md5"`
	Name       string `json:"name"`
	// Not always defined for performance reasons
	Tags []Tag `json:"tags,omitempty"`
}

// Common fields
type Dims struct {
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}

// Only provides the most minimal of fields. Optimal for thumbnail views.
type CompactImage struct {
	Type  FileType `json:"type"`
	SHA1  string   `json:"sha1"`
	Thumb Dims     `json:"thumb"`
}

// TODO: Use these types instead of hex strings

// MD5 hash that encodes itself to hex
type MD5 [16]byte

func (h *MD5) MarshalJSON() ([]byte, error) {
	b := make([]byte, 32+2)
	b[0] = '"'
	b[33] = '"'
	hex.Encode(b[1:], h[:])
	return b, nil
}

func (h *MD5) Value() (driver.Value, error) {
	return h[:], nil
}

func (h *MD5) Scan(src interface{}) error {
	switch src.(type) {
	case string:
		return h.scan([]byte(src.(string)))
	case []byte:
		return h.scan(src.([]byte))
	default:
		return formatConvertError(*h, src)
	}
}

func (h *MD5) scan(src []byte) error {
	if len(src) != 16 {
		return formatConvertError(*h, src)
	}
	copy(h[:], src)
	return nil
}

func formatConvertError(dst, src interface{}) error {
	return fmt.Errorf("cannot convert %T to %T", src, dst)
}

// MD5 hash that encodes itself to hex
type SHA1 [20]byte

func (h *SHA1) MarshalJSON() ([]byte, error) {
	b := make([]byte, 40+2)
	b[0] = '"'
	b[41] = '"'
	hex.Encode(b[1:], h[:])
	return b, nil
}

func (h *SHA1) Value() (driver.Value, error) {
	return h[:], nil
}

func (h *SHA1) Scan(src interface{}) error {
	switch src.(type) {
	case string:
		return h.scan([]byte(src.(string)))
	case []byte:
		return h.scan(src.([]byte))
	default:
		return formatConvertError(*h, src)
	}
}

func (h *SHA1) scan(src []byte) error {
	if len(src) != 20 {
		return formatConvertError(*h, src)
	}
	copy(h[:], src)
	return nil
}
