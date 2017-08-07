#include "bridge.h"

// TODO: Error reporting to the GUI

// Bridges the QML side with with the CGo hydron API
QString Bridge::search(const QString &tags)
{
    auto re = go::searchByTags(tags.toUtf8().data());
    if (re.r1) {
        throw re.r1;
    }
    return QString::fromUtf8(re.r0, -1);
}

QString Bridge::completeTag(const QString &tags)
{
    auto re = go::completeTag(tags.toUtf8().data());
    if (re.r1) {
        throw re.r1;
    }
    return QString::fromUtf8(re.r0, -1);
}
