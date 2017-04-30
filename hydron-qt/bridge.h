#pragma once
#include "libwrapper/libwrapper.h"
#include "libwrapper/types.h"
#include "record.h"
#include <QList>
#include <QObject>
#include <QString>

// Bridges the QML side with with the CGo hydron API
class Bridge : public QObject
{
    Q_OBJECT

  public slots:
    QList<QObject *> search(const QString &tags);
};
