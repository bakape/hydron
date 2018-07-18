package common

type FileType uint8

// Supported file types
const (
	JPEG FileType = iota
	PNG
	GIF
	WEBP
	PDF
	BMP
	PSD
	TIFF
	OGG
	WEBM
	MKV
	MP4
	AVI
	MOV
	WMV
	FLV
)

type MediaType uint8

// File type media container types
const (
	MediaImage MediaType = iota
	MediaVideo
	MediaOther
)

var (
	// Map of MIME types to the constants used internally
	MimeTypes = map[string]FileType{
		"image/jpeg":       JPEG,
		"image/png":        PNG,
		"image/gif":        GIF,
		"image/webp":       WEBP,
		"application/pdf":  PDF,
		"image/bmp":        BMP,
		"image/photoshop":  PSD,
		"image/tiff":       TIFF,
		"application/ogg":  OGG,
		"video/webm":       WEBM,
		"video/x-matroska": MKV,
		"video/mp4":        MP4,
		"video/avi":        AVI,
		"video/quicktime":  MOV,
		"video/x-ms-wmv":   WMV,
		"video/x-flv":      FLV,
	}

	// MIME types allowed to be imported
	AllowedMimes = map[string]bool{
		"image/jpeg":       true,
		"image/png":        true,
		"image/gif":        true,
		"image/webp":       true,
		"application/pdf":  true,
		"image/bmp":        true,
		"image/photoshop":  true,
		"image/tiff":       true,
		"application/ogg":  true,
		"video/webm":       true,
		"video/x-matroska": true,
		"video/mp4":        true,
		"video/avi":        true,
		"video/quicktime":  true,
		"video/x-ms-wmv":   true,
		"video/x-flv":      true,
	}

	// Canonical MIME type extensions
	Extensions = map[FileType]string{
		JPEG: "jpg",
		PNG:  "png",
		GIF:  "gif",
		WEBP: "webp",
		PDF:  "pdf",
		BMP:  "bmp",
		PSD:  "psd",
		TIFF: "tiff",
		OGG:  "ogg",
		WEBM: "webm",
		MKV:  "mkv",
		MP4:  "mp4",
		AVI:  "avi",
		MOV:  "mov",
		WMV:  "wmv",
		FLV:  "flv",
	}

	// Mapping from canonical extensions to internal enum
	RevExtensions = map[string]FileType{
		"jpg":  JPEG,
		"png":  PNG,
		"gif":  GIF,
		"webp": WEBP,
		"pdf":  PDF,
		"bmp":  BMP,
		"psd":  PSD,
		"tiff": TIFF,
		"ogg":  OGG,
		"webm": WEBM,
		"mkv":  MKV,
		"mp4":  MP4,
		"avi":  AVI,
		"mov":  MOV,
		"wmv":  WMV,
		"flv":  FLV,
	}
)

// Map file type to media container type
func GetMediaType(t FileType) MediaType {
	switch t {
	case JPEG, PNG, GIF, WEBP, BMP, TIFF:
		return MediaImage
	case WEBM, OGG, MKV, MP4, AVI, MOV, WMV:
		return MediaVideo
	default:
		return MediaOther
	}
}
