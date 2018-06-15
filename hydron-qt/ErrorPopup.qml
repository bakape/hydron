import QtQuick 2.9
import QtQuick.Controls 2.2


// TODO: Overlapping errors are currently silenced. Need to create new popups for each
Dialog {
    modal: true
    focus: true
    closePolicy: Popup.CloseOnEscape
    title: "Error"
    standardButtons: Dialog.Ok
    x: (parent.width - width) / 2
    y: (parent.height - height) / 2 - searchBar.height
    width: height * 1.618

    Text {
        id: errorText
        wrapMode: Text.Wrap
    }

    function render(err) {
        errorText.text = err
        open()
    }
}
