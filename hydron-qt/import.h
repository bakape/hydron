#pragma once

#include <QFile>
#include <QNetworkReply>
#include <QObject>

// Asynchronously manages a single file upload
class FileUploader : public QObject {
    Q_OBJECT

public:
    // delete_after: delete the file from disk, after uploading
    FileUploader(bool delete_after);

    // path: path to file
    // tags: optional tags to add to all images
    // fetch_tags: fetch tags for image from boorus
    void run(QString path, QString tags, bool fetch_tags);

public slots:
    void onFinished();
    void onSSLError(QList<QSslError>);

signals:
    // Empty json string on error
    void done(QString json, QString error);
    void error(QString err);

private:
    const bool delete_after;
    QFile* file = nullptr;
    QNetworkReply* res = nullptr;
};

// Asynchronously manages a multiple file upload
class BatchUploader : public QObject {
    Q_OBJECT

public:
    // Import all added files
    // tags: optional tags to add to all images
    // delete_after: delete the file from disk, after uploading
    // fetch_tags: fetch tags for image from boorus
    Q_INVOKABLE void run(QString tags, bool delete_after, bool fetch_tags);

    // Add file paths to batch uploader as a file URL list
    Q_INVOKABLE void add_paths(QString paths);

    // Reset to initial state
    Q_INVOKABLE void reset();

signals:
    // Empty JSON string on error
    void done(QString json, QString error, float progress);
    void error(QString err);

private slots:
    // Simple forwarding from child tasks
    void onDone(QString json, QString error);
    void onError(QString err);

private:
    int total = 0;
    int done_count = 0;
    QVector<QString> paths;
    QVector<FileUploader*> workers;
};
