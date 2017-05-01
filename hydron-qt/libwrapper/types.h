#pragma once
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

typedef struct {
	char **tags;
	int len;
} Tags;

enum FileType {
	JPEG,
	PNG,
	GIF,
	WEBP,
	PDF,
	BMP,
	PSD,
	TIFF,
	ICO,
	MP3,
	AAC,
	WAVE,
	FLAC,
	MIDI,
	OGG,
	WEBM,
	MKV,
	MP4,
	AVI,
	MOV,
	WMV,
	FLV
};

typedef struct {
	uint64_t importTime, size, width, height, length;
	char sha1[20];
	char md5[16];
	enum FileType type;
	char *thumbPath, *sourcePath;
	Tags tags;
} Record;
