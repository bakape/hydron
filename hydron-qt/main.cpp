#include "bridge.h"
#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include <QQmlContext>

Bridge *instance = 0;

int main(int argc, char *argv[])
{
    char *err = go::startHydron();
    if (err) {
        throw err;
    }

    QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QGuiApplication app(argc, argv);

    QQmlApplicationEngine engine;
    Bridge bridge;
    instance = &bridge;
    engine.rootContext()->setContextProperty("go", &bridge);
    engine.load(QUrl(QLatin1String("qrc:/main.qml")));

    return app.exec();
}
