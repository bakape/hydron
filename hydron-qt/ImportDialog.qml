import QtQuick 2.9
import QtQuick.Controls 2.2

Dialog {
    property variant files: []
    modal: true
    focus: true
    closePolicy: Popup.CloseOnEscape
    title: "Import files"
    standardButtons: Dialog.Ok | Dialog.Cancel
    x: (parent.width - width) / 2
    y: (parent.height - height) / 2 - searchBar.height
    width: height * 1.618

    Column {
        CheckBox {
            id: deleteAfter
            text: "Delete after import"
        }
        CheckBox {
            id: fetch
            text: "Fetch tags from boorus"
            checked: true
        }
        TextField {
            id: tags
            placeholderText: "Add tags"
        }
    }

    onAccepted: {
        browser.clear()
        go.importFiles(files, tags.text, deleteAfter.checked, fetch.checked)
        files = []
    }
    onRejected: {
        files = []
    }

    function addFiles(f) {
        files = files.concat(f)
    }
}
