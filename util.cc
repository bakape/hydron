#include "util.hh"

char const* format_error(const int code)
{
    char* buf = (char*)malloc(1024);
    av_strerror(code, buf, 1024);
    return buf;
}
