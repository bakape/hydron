import QtQuick 2.9
import QtQuick.Controls 1.4

Menu {
    property variant ids

    MenuItem {
        text: "Delete"
        onTriggered: window.removeFiles(ids)
    }
}
