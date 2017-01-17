#ifndef CGO_THUMBNAILER_H
#define CGO_THUMBNAILER_H

#include <magick/api.h>
#include <stdbool.h>
#include <stdint.h>

struct Thumbnail {
	bool isPNG;
	void *buf;
	size_t size;
};

int thumbnail(const void *src,
			  const size_t size,
			  struct Thumbnail *thumb,
			  ExceptionInfo *ex);
static int writeThumb(Image *img, struct Thumbnail *thumb, ExceptionInfo *ex);
static int hasTransparency(const Image *img, bool *needPNG, ExceptionInfo *ex);
#endif
