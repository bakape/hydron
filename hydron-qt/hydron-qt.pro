TEMPLATE = app

QT += qml quick widgets

CONFIG += c++17 qtquickcompiler reduce-relocations ltcg

SOURCES += main.cpp

RESOURCES += qml.qrc

# Additional import path used to resolve QML modules in Qt Creator's code model
QML_IMPORT_PATH =

# Default rules for deployment.
include(deployment.pri)

DISTFILES +=
