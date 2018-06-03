#pragma once

#include <stddef.h>
#include <stdint.h>

// Helper for passing buffers over FFI
typedef struct {
    uint8_t* data;
    size_t size;
} Buffer;

// THumbnail data
typedef struct {
    Buffer buf;
    unsigned long width, height;
} Thumb;

// Supported input file types
enum FileType { JPEG, PNG, GIF, WEBM };

// Data of file being thumbnailed
typedef struct {
    unsigned long width, height, duration;
} Source;

// Result of a thumbnailing operation
typedef struct {
    Source src;
    Thumb thumb;
    char const* error;
} Result;

// Generate a file thumbnail
Result thumbnail_file(const Buffer src, FileType type);
