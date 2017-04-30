#include "record.h"
#include "bridge.h"
#include <QQmlContext>
#include <QApplication>
#include <QQmlApplicationEngine>


int main(int argc, char *argv[])
{
    QApplication app(argc, argv);

    QQmlApplicationEngine engine;
    Bridge bridge;
    engine.rootContext()->setContextProperty("go", &bridge);
    engine.load(QUrl(QStringLiteral("qrc:/main.qml")));

    return app.exec();
}
