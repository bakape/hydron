#pragma once

extern "C" {
#include "thumbnailer.h"
}
#include <cstddef>
#include <cstdint>
#include <libavformat/avformat.h>

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
        if (png_avc_ctx) {
            avcodec_free_context(&png_avc_ctx);
        }
        if (selected_frame) {
            av_frame_free(&selected_frame);
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
    // Codec context used for encoding PNG thumbnails
    AVCodecContext* png_avc_ctx = NULL;

    // Frame being thumbnailed
    AVFrame* selected_frame = NULL;
};
