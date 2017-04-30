#pragma once
#include "libwrapper/types.h"
#include "libwrapper/libwrapper.h"
#include "record.h"
#include <QObject>
#include <QList>
#include <QString>

// Bridges the QML side with with the CGo hydron API
class Bridge: public QObject
{
    Q_OBJECT

public slots:
    QList<QRecord*> search(const QString &tags);
};
