#pragma once

#include <libavformat/avformat.h>

// Find and extract best video frame to thumbnail
AVFrame* extract_video_frame(
    AVFormatContext* avfc, AVCodecContext* avcc, const int stream);
