#include "thumbnailer.hh"

const int BUFFER_SIZE = 4 << 10;

extern "C" int read_callback(void* thumbnailer, unsigned char* buf, int size)
{
    return static_cast<Thumbnailer*>(thumbnailer)->read_source(buf, size);
}

extern "C" int64_t seek_callback(void* thumbnailer, int64_t offset, int whence)
{
    return static_cast<Thumbnailer*>(thumbnailer)->seek_source(offset, whence);
}

int Thumbnailer::read_source(unsigned char* buf, int size)
{
    if (src_pos < 0 || src_pos >= (int64_t)src_buf.size) {
        return -1;
    }
    int n = size;
    if (src_pos + size >= (int64_t)src_buf.size) {
        n = src_buf.size - src_pos;
    }
    memcpy(buf, (void*)(&src_buf.data[src_pos]), n);
    return n;
}

int64_t Thumbnailer::seek_source(int64_t offset, int whence)
{
    switch (whence) {
    case 0:
        src_pos = 0;
        break;
    case 1:
        break;
    case 2:
        src_pos = (int64_t)src_buf.size;
        break;
    }

    src_pos += offset;
    return src_pos;
}

// Format ffmpeg error code to string message
static char* format_error(const int code)
{
    char* buf = (char*)malloc(1024);
    av_strerror(code, buf, 1024);
    return buf;
}

Result Thumbnailer::process_video()
{
    unsigned char* buf = (unsigned char*)malloc(BUFFER_SIZE);
    avf_ctx = avformat_alloc_context();
    avf_ctx->pb = avio_alloc_context(
        buf, BUFFER_SIZE, 0, this, &read_callback, NULL, &seek_callback);
    avf_ctx->flags |= AVFMT_FLAG_CUSTOM_IO;

    int err = avformat_open_input(&avf_ctx, NULL, NULL, NULL);
    if (err < 0) {
        throw format_error(err);
    }

    // TODO: The rest
}
