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

        Suggestions {
            id: suggestions
            anchors.fill: parent
        }

        FileView {
            id: fileView
            anchors.fill: parent
        }
    }

    header: SearchBar {
        id: searchBar
        Layout.fillWidth: true
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
}
