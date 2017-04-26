package core

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
	ICO
	MP3
	AAC
	WAVE
	FLAC
	MIDI
	OGG
	WEBM
	MKV
	MP4
	AVI
	MOV
	WMV
	FLV
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
		"image/x-icon":     ICO,
		"audio/mpeg":       MP3,
		"audio/aac":        AAC,
		"audio/wave":       WAVE,
		"audio/x-flac":     FLAC,
		"audio/midi":       MIDI,
		"application/ogg":  OGG,
		"video/webm":       WEBM,
		"video/x-matroska": MKV,
		"video/mp4":        MP4,
		"video/avi":        AVI,
		"video/quicktime":  MOV,
		"video/x-ms-wmv":   WMV,
		"video/x-flv":      FLV,
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
		ICO:  "ico",
		MP3:  "mp3",
		AAC:  "aac",
		WAVE: "wave",
		FLAC: "flac",
		MIDI: "midi",
		OGG:  "ogg",
		WEBM: "webm",
		MKV:  "mkv",
		MP4:  "mp4",
		AVI:  "avi",
		MOV:  "mov",
		WMV:  "wmv",
		FLV:  "flv",
	}
)
