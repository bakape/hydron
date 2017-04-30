TEMPLATE = app

QT += qml quick widgets

CONFIG += c++11

SOURCES += main.cpp \
    record.cpp \
    bridge.cpp

RESOURCES += qml.qrc

# Additional import path used to resolve QML modules in Qt Creator's code model
QML_IMPORT_PATH =

# Default rules for deployment.
include(deployment.pri)

HEADERS += \
    libwrapper/types.h \
    libwrapper/libwrapper.h \
    record.h \
    bridge.h

LIBS += -L"$$_PRO_FILE_PWD_/libwrapper" -lwrapper
