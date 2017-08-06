#include "bridge.h"

// TODO: Error reporting to the GUI

// Bridges the QML side with with the CGo hydron API
QString Bridge::search(const QString &tags)
{
    auto re = searchByTags(tags.toUtf8().data());
    if (re.r1) {
        throw re.r1;
    }
    return QString::fromUtf8(re.r0, -1);
}
