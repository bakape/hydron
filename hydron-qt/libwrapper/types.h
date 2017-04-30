#pragma once
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

typedef struct {
	char **tags;
	int len;
} Tags;

typedef struct {
	bool pngThumb, noThumb;
	uint64_t importTime, size, width, height, length;
	char *sha1, *md5, *type;
	Tags tags;
} Record;
