#include "bridge.h"

// Cast JSON C string to QString and handle errors
// TODO: Error reporting to the GUI
#define CAST_JSON \
    if (re.r1) { \
        throw re.r1; \
    } \
    return QString::fromUtf8(re.r0, -1); \

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
