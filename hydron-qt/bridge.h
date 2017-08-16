#pragma once
#include <QList>
#include <QObject>
#include <QString>

namespace go {
    #include "libwrapper/libwrapper.h"
}

// Bridges the QML side with with the CGo hydron API
class Bridge : public QObject
{
    Q_OBJECT

  public slots:
    QString search(const QString &tags);
    QString completeTag(const QString &tags);
    QString get(const QString &id);
    void remove(const QString &id);
};

void print_err(const char * const err);
