#include "record.h"

// Decode tags from C into Qt types
QList<QString> decodeTags(Tags &tags)
{
    QList<QString> list;
    if (tags.tags) {
        list.reserve(tags.len);
        for (int i = 0; i < tags.len; i++) {
            list << QString::fromUtf8(*(tags.tags + i));
        }
        free(tags.tags);
    }
    return list;
}

// Create new model from a C Record
QRecord::QRecord(Record &r)
{
    selected = false;
    type = r.type;
    importTime = r.importTime;
    size = r.size;
    width = r.width;
    height = r.height;
    length = r.length;
    std::copy(std::begin(r.sha1), std::end(r.sha1), std::begin(sha1));
    std::copy(std::begin(r.md5), std::end(r.md5), std::begin(md5));
    if (r.sourcePath) {
        sourcePath = QString::fromUtf8(r.sourcePath);
    }
    if (r.thumbPath) {
        sourcePath = QString::fromUtf8(r.thumbPath);
    }
    tags = decodeTags(r.tags);
}

void QRecord::setSelected(bool val)
{
    selected = val;
}

bool QRecord::getSelected()
{
    return selected;
}

QString QRecord::getSourcePath()
{
    return sourcePath;
}

// Decode an array of C Records
QList<QObject *> decodeRecords(Record *recs, int len)
{
    QList<QObject *> list;
    if (recs) {
        list.reserve(len);
        for (int i = 0; i < len; i++) {
            list << new QRecord(*(recs + i));
        }
        free(recs);
    }
    return list;
}
