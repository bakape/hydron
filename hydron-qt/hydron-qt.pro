TEMPLATE = app

QT += qml quick widgets

CONFIG += c++17

SOURCES += main.cpp \
    import.cpp

RESOURCES += qml.qrc

# Additional import path used to resolve QML modules in Qt Creator's code model
QML_IMPORT_PATH =

# Default rules for deployment.
include(deployment.pri)

DISTFILES +=

HEADERS += \
    import.hpp
