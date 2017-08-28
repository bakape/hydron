#pragma once
#include <QList>
#include <QObject>
#include <QString>

namespace go {
#include "libwrapper/libwrapper.h"
}

// Converts C string to QString
QString to_Qstring(char *s);

// Bridges the QML side with with the CGo hydron API
class Bridge : public QObject
{
    Q_OBJECT

signals:
    void error(const QString &err);
    void set_progress_bar(double pos);
    void done(const QString &rec);

public slots:
    QString search(const QString &tags);
    QString completeTag(const QString &tags);
    QString get(const QString &id);
    void remove(const QString &id);
    void importFiles(const QStringList &urls,
                     const QString &tags,
                     bool del,
                     bool fetch);
};

extern Bridge * instance;

void _pass_error(const char *err);
void _set_progress_bar(double pos);
void _pass_done(const char *rec);
