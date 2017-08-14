import QtQuick 2.7

ListView {
    anchors.leftMargin: 7
    interactive: false
    model: ListModel {}

    delegate: Rectangle {
        color: SystemPalette.base || "white"
        width: childrenRect.width
        height: childrenRect.height

        Text {
            text: tag
        }
    }
}
