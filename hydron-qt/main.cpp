#include "bridge.h"
#include <QApplication>
#include <QCoreApplication>
#include <QObject>
#include <QQmlApplicationEngine>
#include <QQmlContext>

int main(int argc, char *argv[])
{
    char* err = go::startHydron();
    if (err) {
        throw err;
    }

    QApplication app(argc, argv);

    QQmlApplicationEngine engine;
    Bridge bridge;
    engine.rootContext()->setContextProperty("go", &bridge);
    engine.load(QUrl(QStringLiteral("qrc:/main.qml")));

    return app.exec();
}
