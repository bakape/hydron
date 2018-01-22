#include "thumbnailer.hh"
#include <cstdlib>
#include <cstring>
#include <exception>

extern "C" Result thumbnail_file(const Buffer src, FileType type)
{
    try {
        auto t = Thumbnailer(src, type);
        return t.process();
    } catch (const std::exception& e) {
        Result res;
        res.error = (char*)malloc(sizeof(char) * strlen(e.what()));
        strcpy(res.error, e.what());
        return res;
    }
}

Result Thumbnailer::process()
{
    switch (src_type) {
    case JPEG:
    case PNG:
    case GIF:
    case WEBP:
        throw "image thumbnailing unimplemented";
        return Result{};
    case WEBM:
    case MP4:
        return process_video();
    default:
        return Result{};
    }
}
