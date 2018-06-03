#include "thumbnailer.hh"
#include "util.hh"
#include <libswscale/swscale.h>
#include <mutex>

/**
 * Potential thumbnail lookup filter to reduce the risk of an inappropriate
 * selection (such as a black frame) we could get with an absolute seek.
 *
 * Simplified version of algorithm by Vadim Zaliva <lord@crocodile.org>.
 * http://notbrainsurgery.livejournal.com/29773.html
 *
 * Adapted by Janis Petersons <bakape@gmail.com>
 */
static const int HIST_SIZE = 3 * 256;
static const int MAX_FRAMES = 100;

// Compute sum-square deviation to estimate "closeness"
static double compute_error(
    const int hist[HIST_SIZE], const double median[HIST_SIZE])
{
    double sum_sq_err = 0;
    for (int i = 0; i < HIST_SIZE; i++) {
        const double err = median[i] - (double)hist[i];
        sum_sq_err += err * err;
    }
    return sum_sq_err;
}

// Select best frame based on RGB histograms
static int select_best_frame(AVFrame* frames[])
{
    // RGB color distribution histograms of the frames. Reused between all calls
    // to
    // avoid allocations. Need lock of hist_mu.
    static int hists[MAX_FRAMES][HIST_SIZE];
    static std::mutex hist_mu;

    hist_mu.lock();

    // First rezero after last use
    memset(hists, 0, sizeof(int) * MAX_FRAMES * HIST_SIZE);

    // Compute each frame's histogram
    int frame_i;
    for (frame_i = 0; frame_i < MAX_FRAMES; frame_i++) {
        const AVFrame* f = frames[frame_i];
        if (!f || !f->data) {
            frame_i--;
            break;
        }
        uint8_t* p = f->data[0];
        for (int j = 0; j < f->height; j++) {
            for (int i = 0; i < f->width; i++) {
                for (int k = 0; k < 3; k++) {
                    hists[frame_i][k * 256 + p[i * 3 + k]]++;
                }
            }
            p += f->linesize[0];
        }
    }

    // Average histograms of up to 100 frames
    double average[HIST_SIZE] = { 0 };
    for (int j = 0; j <= frame_i; j++) {
        for (int i = 0; i <= frame_i; i++) {
            average[j] = (double)hists[i][j];
        }
        average[j] /= frame_i + 1;
    }

    // Find the frame closer to the average using the sum of squared errors
    double min_sq_err = -1;
    int best_i = 0;
    for (int i = 0; i <= frame_i; i++) {
        const double sq_err = compute_error(hists[i], average);
        if (i == 0 || sq_err < min_sq_err) {
            best_i = i;
            min_sq_err = sq_err;
        }
    }

    hist_mu.unlock();
    return best_i;
}

AVFrame* extract_video_frame(
    AVFormatContext* avfc, AVCodecContext* avcc, const int stream)
{
    int err = 0;
    int frame_i = 0;
    AVPacket pkt;
    AVFrame* frames[MAX_FRAMES] = { 0 };

    // Read up to 100 frames
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

            if (!frames[frame_i]) {
                frames[frame_i] = av_frame_alloc();
            }
            err = avcodec_receive_frame(avcc, frames[frame_i]);
            switch (err) {
            case 0:
                if (++frame_i == MAX_FRAMES) {
                    goto end;
                }
                av_packet_unref(&pkt);
                continue;
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
    int best_frame = -1;
    switch (err) {
    case AVERROR_EOF:
        err = 0;
        [[fallthrough]];
    case 0:
        best_frame = select_best_frame(frames);
        break;
    }
    av_packet_unref(&pkt);
    for (int i = 0; i < MAX_FRAMES; i++) {
        if (!frames[i]) {
            break;
        }
        if (i != best_frame) {
            av_frame_free(&frames[i]);
        }
    }
    if (err) {
        throw format_error(err);
    }
    return frames[best_frame];
}
