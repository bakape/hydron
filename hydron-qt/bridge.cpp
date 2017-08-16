#include "bridge.h"
#include <stdio.h>

// Cast JSON C string to QString and handle errors
// TODO: Error reporting to the GUI
#define CAST_JSON \
    if (re.r1) { \
        print_err(re.r1); \
    } \
    return QString::fromUtf8(re.r0, -1); \

void print_err(const char * const err){
    fprintf(stderr, "error: %s\n", err);
}

// Bridges the QML side with with the CGo hydron API
QString Bridge::search(const QString &tags)
{
    auto re = go::searchByTags(tags.toUtf8().data());
    CAST_JSON
}

QString Bridge::completeTag(const QString &tags)
{
    auto re = go::completeTag(tags.toUtf8().data());
    CAST_JSON
}

QString Bridge::get(const QString &id)
{
    auto re = go::getRecord(id.toUtf8().data());
    CAST_JSON
}

void Bridge::remove(const QString &id)
{
    auto err = go::remove(id.toUtf8().data());
    if (err) {
        print_err(err);
    }
}
