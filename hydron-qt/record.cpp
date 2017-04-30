#include "record.h"

// Decode tags from C into Qt types
QList<QString> decodeTags(Tags &tags)
{
    QList<QString> list;
    if (tags.tags) {
        list.reserve(tags.len);
        for (int i = 0; i < tags.len; i++) {
            char **pos = tags.tags + (size_t)i;
            list.append(QString::fromUtf8(*pos));
        }
        free(tags.tags);
    }
    return list;
}

// Create new model from a C Record
QRecord::QRecord(Record &r)
{
    selected = false;
    pngThumb = r.pngThumb;
    noThumb = r.noThumb;
    importTime = r.importTime;
    size = r.size;
    width = r.width;
    height = r.height;
    length = r.length;
    sha1 = QString::fromUtf8(r.sha1);
    type = QString::fromUtf8(r.type);
    if (r.md5) {
        md5 = QString::fromUtf8(r.md5);
    }
    tags = decodeTags(r.tags);
}

// Decode an array of C Records
QList<QObject *> decodeRecords(Record *recs, int len)
{
    QList<QObject *> list;
    if (recs) {
        list.reserve(len);
        const size_t size = sizeof(Record);
        for (int i = 0; i < len; i++) {
            Record r = *(recs + size * (size_t)i);
            list.append(new QRecord(r));
        }
        free(recs);
    }
    return list;
}
