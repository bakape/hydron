#include "thumbnailer.hh"
#include "util.hh"
#include "video.hh"
#include <cstdlib>
#include <cstring>
#include <exception>
#include <libavutil/imgutils.h>
#include <libswscale/swscale.h>
#include <mutex>
#include <unordered_map>
#include <vector>

// Guards against concurent access to codec lookup
static std::mutex codec_mu;

// Initializes libs right before processin, but only once
static std::once_flag init_flag;

static void init()
{
#if LIBAVCODEC_VERSION_INT < AV_VERSION_INT(58, 9, 100)
    av_register_all();
#endif
#if LIBAVCODEC_VERSION_INT < AV_VERSION_INT(58, 10, 100)
    avcodec_register_all();
#endif
    av_log_set_level(16);
}

extern "C" Result thumbnail_file(const Buffer src, FileType type)
{
    try {
        auto t = Thumbnailer(src, type);
        return t.process();
    } catch (char const* e) {
        Result res;
        res.error = e;
        return res;
    } catch (const char* e) {
        Result res;
        char* buf = (char*)malloc(strlen(e) * sizeof(char));
        strcpy(buf, e);
        res.error = buf;
        return res;
    }
}

// Size of one thumbnail side
static const double THUMB_SIZE = 150;

// Scale frame and record dimentions
static void scale_frame(const AVFrame* frame, Result* re)
{
    // Compute and persist all required dimentions. Maintain aspect ratio.
    re->src.height = frame->height;
    re->src.width = frame->width;
    double scale;
    if (re->src.width >= re->src.height) {
        scale = (double)re->src.width / THUMB_SIZE;
    } else {
        scale = (double)re->src.height / THUMB_SIZE;
    }
    re->thumb.width = (unsigned long)((double)re->src.width / scale);
    re->thumb.height = (unsigned long)((double)re->src.height / scale);

    struct SwsContext* ctx = sws_getContext(re->src.width, re->src.height,
        (AVPixelFormat)frame->format, re->thumb.width, re->thumb.height,
        AV_PIX_FMT_RGBA, SWS_BICUBIC, 0, 0, 0);
    if (!ctx) {
        throw "could not allocate SwsContext";
    }

    int dst_linesize[1];
    uint8_t* dst_data[1];
    re->thumb.buf.size = (size_t)av_image_get_buffer_size(
        AV_PIX_FMT_RGBA, re->thumb.width, re->thumb.height, 1);
    re->thumb.buf.data = dst_data[0]
        = (uint8_t*)malloc(re->thumb.buf.size); // RGB have one plane
    dst_linesize[0] = 4 * re->thumb.width; // RGBA stride

    sws_scale(ctx, (const uint8_t* const*)frame->data, frame->linesize, 0,
        frame->height, (uint8_t * const*)dst_data, dst_linesize);
    sws_freeContext(ctx);
}

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

// Simply read the first frame
static AVFrame* read_image_to_frame(
    AVFormatContext* avfc, AVCodecContext* avcc, const int stream)
{
    int err = 0;
    AVPacket pkt;
    AVFrame* frame = av_frame_alloc();

    while (1) {
        err = av_read_frame(avfc, &pkt);
        switch (err) {
        case 0:
            break;
        default:
            goto end;
        }

        if (pkt.stream_index == stream) {
            err = avcodec_send_packet(avcc, &pkt);
            if (err < 0) {
                goto end;
            }
            err = avcodec_receive_frame(avcc, frame);
            switch (err) {
            case 0:
                goto end;
            case AVERROR(EAGAIN):
                av_packet_unref(&pkt);
                continue;
            default:
                goto end;
            }
        } else {
            av_packet_unref(&pkt);
        }
    }

end:
    av_packet_unref(&pkt);
    if (err != 0) {
        av_frame_free(&frame);
        throw format_error(err);
    }
    return frame;
}

Result Thumbnailer::process()
{
    std::call_once(init_flag, init);

    const int BUFFER_SIZE = 4 << 10;
    unsigned char* buf = (unsigned char*)malloc(BUFFER_SIZE);
    avf_ctx = avformat_alloc_context();
    avf_ctx->pb = avio_alloc_context(
        buf, BUFFER_SIZE, 0, this, &read_callback, NULL, &seek_callback);
    avf_ctx->flags |= AVFMT_FLAG_CUSTOM_IO;

    int err = avformat_open_input(&avf_ctx, NULL, NULL, NULL);
    if (err < 0) {
        throw format_error(err);
    }
    if (!avf_ctx) {
        throw "unknown context creation error";
    }

    // Calls avcodec_open2 internally, so needs locking
    codec_mu.lock();
    err = avformat_find_stream_info(avf_ctx, NULL);
    codec_mu.unlock();
    if (err < 0) {
        throw format_error(err);
    }

    // Check for webm video stream presence and retrieve codec context

    AVStream* st = nullptr;
    AVCodec* codec = nullptr;

    int stream
        = av_find_best_stream(avf_ctx, AVMEDIA_TYPE_VIDEO, -1, -1, NULL, 0);
    if (stream < 0) {
        throw format_error(stream);
    }
    st = avf_ctx->streams[stream];

    const AVCodecID codec_id = st->codecpar->codec_id;

    // ffvp8/9 doesn't support alpha channel, so force libvpx
    switch (codec_id) {
    case AV_CODEC_ID_VP8:
        codec = avcodec_find_decoder_by_name("libvpx");
        break;
    case AV_CODEC_ID_VP9:
        codec = avcodec_find_decoder_by_name("libvpx-vp9");
        break;
    default:
        codec = avcodec_find_decoder(codec_id);
    }
    if (!codec) {
        throw "unsuported file format";
    }

    avc_ctx = avcodec_alloc_context3(codec);
    if (!avc_ctx) {
        throw "could not allocate codec context";
    }
    err = avcodec_parameters_to_context(avc_ctx, st->codecpar);
    if (err < 0) {
        throw format_error(err);
    }

    // Not thread safe. Needs lock.
    codec_mu.lock();
    err = avcodec_open2(avc_ctx, codec, NULL);
    codec_mu.unlock();
    if (err < 0) {
        throw format_error(err);
    }

    Result re;
    switch (codec_id) {
    case AV_CODEC_ID_MJPEG:
    case AV_CODEC_ID_PNG:
        selected_frame = read_image_to_frame(avf_ctx, avc_ctx, stream);
        re.src.duration = 0;
        break;
    case AV_CODEC_ID_APNG:
    case AV_CODEC_ID_GIF:
    case AV_CODEC_ID_VP8:
    case AV_CODEC_ID_VP9:
        selected_frame = extract_video_frame(avf_ctx, avc_ctx, stream);
        re.src.duration = avf_ctx->duration;
    }

    scale_frame(selected_frame, &re);
    return re;
}
