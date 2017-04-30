#include "bridge.h"

// TODO: Error reporting to the GUI

// Bridges the QML side with with the CGo hydron API
QList<QObject *> Bridge::search(const QString &tags)
{
    auto re = searchByTags(tags.toUtf8().data());
    if (re.r2) {
        throw re.r2;
    }
    return decodeRecords(re.r0, re.r1);
}
