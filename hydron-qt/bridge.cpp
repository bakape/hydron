#include "bridge.h"
#include <stdio.h>
#include <QQmlApplicationEngine>
#include <QQuickItem>
#include <QMetaObject>

// Cast JSON C string to QString and handle errors
// TODO: Error reporting to the GUI
#define CAST_JSON \
    if (re.r1) { \
        _pass_error(re.r1); \
    } \
    return QString::fromUtf8(re.r0, -1); \

void _pass_error(const char *err){
    instance->error(QString::fromUtf8(err, -1));
}

void _set_progress_bar(double pos)
{
    instance->set_progress_bar(pos);
}

void _pass_done(const char *rec)
{
    instance->done(QString::fromUtf8(rec, -1));
}

// Bridges the QML side with with the CGo hydron API
QString Bridge::search(const QString &tags)
{
    auto re = go::searchByTags(tags.toUtf8().data());
    CAST_JSON
}

QString Bridge::completeTag(const QString &tags)
{
    auto re = go::completeTag(tags.toUtf8().data());
    CAST_JSON
}

QString Bridge::get(const QString &id)
{
    auto re = go::getRecord(id.toUtf8().data());
    CAST_JSON
}

void Bridge::remove(const QString &id)
{
    auto err = go::remove(id.toUtf8().data());
    if (err) {
        _pass_error(err);
    }
}

void Bridge::importFiles(const QStringList &urls,
            const QString &tags,
            bool del,
            bool fetch)
{
    go::importFiles(urls.join('\n').toUtf8().data(),
                    tags.toUtf8().data(),
                    del,
                    fetch,
                    (void*)(_set_progress_bar),
                    (void*)(_pass_done),
                    (void*)(_pass_error));
}
