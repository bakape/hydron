#pragma once
#include "libwrapper/types.h"
#include <QList>
#include <QObject>
#include <QString>

// Decode tags from C into Qt types
QList<QString> decodeTags(Tags &tags);

class QRecord : public QObject
{
    Q_OBJECT

  public:
    bool selected, pngThumb, noThumb;
    unsigned int importTime, size, width, height, length;
    QString sha1, md5, type;
    QList<QString> tags;

    // Create new model from a C Record
    QRecord(Record &r);
};

// Decode an array of C Records
QList<QObject *> decodeRecords(Record *recs, int len);
