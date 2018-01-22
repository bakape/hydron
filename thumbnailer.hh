#pragma once

#include <cstddef>
#include <cstdint>

// Helper for passing buffers over FFI
typedef struct {
    uint8_t* data;
    std::size_t size;
} Buffer;

// Generic image data
typedef struct {
    Buffer buf;
    unsigned long width, height;
} Image;

// Supported input file types
enum FileType { JPEG, PNG, GIF, WEBP, WEBM, MP4 };

// Data of file being thumbnailed
typedef struct {
    FileType type;
    unsigned long width, height, duration;
} Source;

// Result of a thumbnailing operation
typedef struct {
    Source src;
    Image thumb;
    char* error;
} Result;

// Generate a file thumbnail
extern "C" Result thumbnail_file(const Buffer);

// Performs the thumbnailing operation
class Thumbnailer {
public:
    Thumbnailer(const Buffer src)
        : src_buf(src)
    {
    }

    // Process the file buffer
    Result process();

private:
    const Buffer src_buf;
};
