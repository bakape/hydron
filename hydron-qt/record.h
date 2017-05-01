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
    Q_PROPERTY(bool selected READ getSelected WRITE setSelected)
    Q_PROPERTY(QString sourcePath READ getSourcePath)

  public:
    bool selected;
    FileType type;
    unsigned long importTime, size, width, height, length;
    QString sourcePath, thumbPath;
    QList<QString> tags;
    QRecord(Record &); // Create new model from a C Record
    void setSelected(bool);
    bool getSelected();
    QString getSourcePath();

  private:
    char sha1[20];
    char md5[16];
};

// Decode an array of C Records
QList<QObject *> decodeRecords(Record *recs, int len);
