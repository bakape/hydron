#include "thumbnailer.hh"
#include <cstdlib>
#include <cstring>
#include <exception>

// Stub for testing
int main() { return 0; }

extern "C" Result thumbnail_file(const Buffer src)
{
    try {
        auto t = Thumbnailer(src);
        return t.process();
    } catch (const std::exception& e) {
        Result res;
        res.error = (char*)malloc(sizeof(char) * strlen(e.what()));
        strcpy(res.error, e.what());
        return res;
    }
}

Result Thumbnailer::process() {}
