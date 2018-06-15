import QtQuick 2.9
import QtQuick.Controls 2.2

Dialog {
    property string error

    modal: true
    focus: true
    closePolicy: Popup.CloseOnEscape
    title: "Error"
    standardButtons: Dialog.Ok
    x: (parent.width - width) / 2
    y: (parent.height - height) / 2

    contentItem: Rectangle {
        implicitWidth: 300
        implicitHeight: 50
        Text {
            text: error
            anchors.centerIn: parent
        }
    }
}
