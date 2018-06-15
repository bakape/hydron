import QtQuick 2.9
import QtQuick.Controls 2.2
import uploader 1.0

Dialog {
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

    Uploader {
        id: uploader

        onDone: {
            progressBar.set(progress)
            if (json) {
                browser.append(JSON.parse(json))
            }
            if (error) {
                errorPopup.render(error)
            }
        }

        onError: {
            errorPopup.render(err)
        }
    }

    onAccepted: {
        browser.clear()
        uploader.run(tags.text, deleteAfter.checked, fetch.checked)
    }

    onRejected: {
        uploader.reset()
    }

    function addFiles(f) {
        uploader.add_paths(f)
    }
}
