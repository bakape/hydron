#pragma once

#include <cstddef>
#include <cstdint>
#include <libavformat/avformat.h>

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
    unsigned long width, height, duration;
    uint8_t md5[16], sha1[20];
} Source;

// Result of a thumbnailing operation
typedef struct {
    Source src;
    Image thumb;
    char* error;
} Result;

// Generate a file thumbnail
extern "C" Result thumbnail_file(const Buffer src, FileType type);

// Performs the thumbnailing operation
class Thumbnailer {
public:
    Thumbnailer(const Buffer src, FileType type)
        : src_buf(src)
        , src_type(type)
    {
    }

    ~Thumbnailer()
    {
        if (avf_ctx) {
            av_free(avf_ctx->pb->buffer);
            avf_ctx->pb->buffer = NULL;
            av_free(avf_ctx->pb);
            av_free(avf_ctx);
        }
        if (avc_ctx) {
            avcodec_free_context(&avc_ctx);
        }
    }

    // Process the file buffer
    Result process();

    // Read source file buffer into passed buffer
    int read_source(unsigned char* buf, int size);

    // Seek source file buffer to offset
    int64_t seek_source(int64_t offset, int whence);

private:
    // Source file buffer
    const Buffer src_buf;

    // Source file type
    const FileType src_type;

    // Position in source file buffer FFmpeg is currently seeked to
    int64_t src_pos = 0;

    AVFormatContext* avf_ctx = NULL;
    AVCodecContext* avc_ctx = NULL;

    Result process_video();
};
