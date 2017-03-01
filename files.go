package main

type fileType uint8

// Supported file types
const (
	jpeg fileType = iota
	png
	gif
	webp
	pdf
	bmp
	psd
	tiff
	ico
	mp3
	aac
	wave
	flac
	midi
	ogg
	webm
	mkv
	mp4
	avi
	mov
	wmv
	flv
)

var (
	// Map of MIME types to the constants used internally
	mimeTypes = map[string]fileType{
		"image/jpeg":       jpeg,
		"image/png":        png,
		"image/gif":        gif,
		"image/webp":       webp,
		"application/pdf":  pdf,
		"image/bmp":        bmp,
		"image/photoshop":  psd,
		"image/tiff":       tiff,
		"image/x-icon":     ico,
		"audio/mpeg":       mp3,
		"audio/aac":        aac,
		"audio/wave":       wave,
		"audio/x-flac":     flac,
		"audio/midi":       midi,
		"application/ogg":  ogg,
		"video/webm":       webm,
		"video/x-matroska": mkv,
		"video/mp4":        mp4,
		"video/avi":        avi,
		"video/quicktime":  mov,
		"video/x-ms-wmv":   wmv,
		"video/x-flv":      flv,
	}

	// Canonical MIME type extensions
	extensions = map[fileType]string{
		jpeg: "jpg",
		png:  "png",
		gif:  "gif",
		webp: "webp",
		pdf:  "pdf",
		bmp:  "bmp",
		psd:  "psd",
		tiff: "tiff",
		ico:  "ico",
		mp3:  "mp3",
		aac:  "aac",
		wave: "wave",
		flac: "flac",
		midi: "midi",
		ogg:  "ogg",
		webm: "webm",
		mkv:  "mkv",
		mp4:  "mp4",
		avi:  "avi",
		mov:  "mov",
		wmv:  "wmv",
		flv:  "flv",
	}
)
