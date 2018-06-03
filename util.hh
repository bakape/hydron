#pragma once

#include <libavformat/avformat.h>

// Format ffmpeg error code to string message
char const* format_error(const int code);
