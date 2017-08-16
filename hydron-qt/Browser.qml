import QtQuick 2.9
import QtQuick.Layouts 1.3

GridView {
    cellHeight: 155
    cellWidth: 155
    model: ListModel {}

    highlightFollowsCurrentItem: false
    property variant highlighted: ({})

    focus: true
    activeFocusOnTab: true

    function loadThumbnails (tags) {
        model.clear()
        highlighted = {}

        var data = JSON.parse(go.search(tags))
        for (var i = 0; i < data.length; i++) {
            var m = data[i]
            m.selected = m.showMenu = false
            model.append(m)
        }
    }

    delegate: Rectangle {
        height: 155
        width: 155
        color: selected
               ? (SystemPalette.highlight || "lightsteelblue")
               : "transparent"
        radius: 5
        Image {
            anchors {
                horizontalCenter: parent.horizontalCenter
                verticalCenter: parent.verticalCenter
            }
            asynchronous: true
            sourceSize: "150x150"
            source: thumbPath
            focus: true
        }
    }

    Keys.onPressed: {
        if (event.modifiers & Qt.MetaModifier) {
            return
        }

        // TODO: Page Up and Page Down

        switch (event.key) {
        case Qt.Key_Up:
            moveCurrentIndexUp()
            select(currentIndex, false)
            break
        case Qt.Key_Down:
            moveCurrentIndexDown()
            select(currentIndex, false)
            break
        case Qt.Key_Left:
            moveCurrentIndexLeft()
            select(currentIndex, false)
            break
        case Qt.Key_Right:
            moveCurrentIndexRight()
            select(currentIndex, false)
            break
        case Qt.Key_Home:
            positionViewAtBeginning()
            break
        case Qt.Key_End:
            positionViewAtEnd()
            break
        case Qt.Key_Space:
        case Qt.Key_Return:
            open(currentIndex)
            break
        case Qt.Key_Delete:
            window.removeFiles(selectedIDs())
            break
        }
    }

    Drag.active: dragArea.drag.active
    Drag.supportedActions: Qt.CopyAction

    MouseArea {
        id: dragArea
        anchors.fill: parent
        acceptedButtons: Qt.LeftButton | Qt.RightButton
        drag.target: parent

        onPressed: {
            forceActiveFocus()
            var i = indexAt(mouse.x, mouse.y + contentY)
            var multiple = !!(mouse.modifiers & Qt.ControlModifier)
            if (mouse.button === Qt.RightButton) {
                if (!model.get(i).selected) {
                    select(i, multiple)
                }
                showMenu()
            } else {
                select(i, multiple)
            }
        }

        onPressAndHold:	parent.grabToImage(function(result) {
            parent.Drag.imageSource = result.url
        })

        onDoubleClicked: {
            open(indexAt(mouse.x, mouse.y + contentY))
        }
    }

    // Select and highlight a file. Optionally allow multiple selection.
    function select(i, multiple) {
        currentIndex = i

        if (!multiple || i === -1) {
            for (var id in highlighted) {
                model.get(parseInt(id)).selected = false
            }
            highlighted = {}
        }

        if (i !== -1) {
            var m = model.get(i)
            if (multiple && m.selected) {
                m.selected = false
                delete highlighted[i]
            } else {
                m.selected = true
                highlighted[i] = true
            }
            positionViewAtIndex(i, GridView.Contain)
        }
    }

    // Open a thumbnail for full screeen viewing
    function open(i) {
        suggestions.model.clear()
        fileView.render(model.get(i).sha1)
    }

    // Return selected files as an array of IDs
    function selectedIDs() {
        var ids = []
        for (var i = 0; i < model.count; i++) {
            var m = model.get(i)
            if (m.selected) {
                ids.push(m.sha1)
            }
        }
        return ids
    }

    // Show context menu for selected files
    function showMenu() {
        Qt.createComponent("FileMenu.qml")
            .createObject(fileView, { ids: selectedIDs() })
            .popup()
    }
}
