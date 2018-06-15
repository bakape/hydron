#include "import.hpp"
#include <QFile>
#include <QHttpMultiPart>
#include <QJsonDocument>
#include <QMimeDatabase>
#include <QNetworkAccessManager>
#include <sstream>

void BatchUploader::run(QString tags, bool delete_after, bool fetch_tags)
{
    total = paths.size();
    for (auto p : paths) {
        auto fu = new FileUploader(delete_after);
        workers.append(fu);
        fu->setParent(this);
        connect(fu, SIGNAL(done(QString, QString)), this,
            SLOT(onDone(QString, QString)));
        connect(fu, SIGNAL(error(QString)), this, SLOT(onError(QString)));
        fu->run(p, tags, fetch_tags);
    }
}

void BatchUploader::add_paths(QString paths)
{
    for (auto p : paths.split('\n')) {
        if (p != "") {
            BatchUploader::paths.append(p);
        }
    }
}

void BatchUploader::reset()
{
    total = 0;
    done_count = 0;
    paths.clear();
    for (auto w : workers) {
        delete w;
    }
    workers.clear();
}

void BatchUploader::onDone(QString json, QString error)
{
    emit done(json, error, (float)(++done_count) / (float)total);
    if (done_count == total) {
        reset();
    }
}

void BatchUploader::onError(QString err) { emit error(err); }

FileUploader::FileUploader(bool delete_after)
    : delete_after(delete_after)
{
}

// Create a text part of a multipart form
static QHttpPart text_part(const char* name, std::string val)
{
    std::ostringstream s;
    s << "form data; name=\"" << name << '"';

    QHttpPart p;
    p.setHeader(
        QNetworkRequest::ContentDispositionHeader, QVariant(s.str().data()));
    p.setBody(val.data());
    return p;
}

void FileUploader::run(QString path, QString tags, bool fetch_tags)
{
    if (path.startsWith("file://")) {
        path = path.mid(7, -1);
    }
    file = new QFile(path);
    file->setParent(this);
    if (!file->open(QIODevice::ReadWrite)) {
        emit error(file->errorString());
        return;
    }

    QHttpMultiPart* mp = new QHttpMultiPart(QHttpMultiPart::FormDataType);
    mp->setParent(this);
    mp->append(text_part("tags", tags.toStdString()));
    mp->append(text_part("fetch_tags", std::to_string(fetch_tags)));

    QHttpPart img;
    QMimeDatabase mdb;
    img.setHeader(QNetworkRequest::ContentTypeHeader,
        QVariant(mdb.mimeTypeForFile(path).name()));
    img.setHeader(QNetworkRequest::ContentDispositionHeader,
        QVariant("form-data; name=\"file\""));
    img.setBodyDevice(file);
    mp->append(img);

    static thread_local QNetworkAccessManager mgr;
    res = mgr.post(
        QNetworkRequest(QUrl("http://localhost:8010/api/images/")), mp);
    res->setParent(this);

    connect(res, SIGNAL(finished()), this, SLOT(onFinished()));
    connect(res, SIGNAL(sslErrors(QList<QSslError>)), this,
        SLOT(onSSLError(QList<QSslError>)));
}

void FileUploader::onFinished()
{
    QString json, err;
    const auto s = QString::fromUtf8(res->readAll());
    switch (res->error()) {
    case QNetworkReply::NoError:
        if (delete_after) {
            file->remove();
        }
        json = s;
        break;
    case QNetworkReply::UnknownContentError:
        // Ignore unsupported file types
        if (res->attribute(QNetworkRequest::HttpStatusCodeAttribute) == 415) {
            break;
        }
        [[fallthrough]];
    default:
        err = s;
    }
    emit done(json, err);
}

void FileUploader::onSSLError(QList<QSslError> _)
{
    Q_UNUSED(_);
    res->ignoreSslErrors();
}
