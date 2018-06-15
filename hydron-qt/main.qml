import QtQuick 2.9
import QtQml 2.2
import QtQuick.Controls 2.2
import QtQuick.Layouts 1.3

ApplicationWindow {
    id: window
    visible: true
    visibility: "Maximized"
    title: "hydron-qt"
    minimumWidth: 640
    minimumHeight: 480

    Item {
        id: overlay
        anchors.fill: parent
        z: 100

        FileView {
            id: fileView
            anchors.fill: parent
        }
    }

    header: ToolBar {
        ColumnLayout {
            anchors.fill: parent
            SearchBar {
                id: searchBar
                Layout.fillWidth: true
            }
            ProgressBar {
                id: progressBar
                visible: false
                Layout.fillWidth: true

                function set(pos) {
                    visible = pos !== 0 && pos !== 1
                    value = pos
                }
            }
        }
    }

    ScrollView {
        id: browserContainer
        anchors.fill: parent

        Browser {
            id: browser
            anchors.fill: parent
        }
    }

    Shortcut {
        sequence: StandardKey.Quit
        context: Qt.ApplicationShortcut
        onActivated: Qt.quit()
    }

    ImportDialog {
        id: importDialog
    }

    DropArea {
        anchors.fill: parent
        keys: ["text/uri-list"]
        onDropped: {
            // Ignore internal drags
            if (drop.source || !drop.hasUrls) {
                return
            }
            drop.accepted = true
            importDialog.addFiles(drop.text)
            importDialog.open()
        }
    }

    ErrorPopup {
        id: errorPopup
    }
}
