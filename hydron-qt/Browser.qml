import QtQuick 2.9
import QtQuick.Layouts 1.3
import "http.js" as HTTP
import "paths.js" as Paths

GridView {
    cellHeight: 155
    cellWidth: 155
    model: ListModel {
    }
    highlightFollowsCurrentItem: true

    focus: true
    activeFocusOnTab: true

    function loadThumbnails(tags) {
        clear()

        HTTP.get("/images/search/" + tags, function (data, err) {
            if (err) {
                displayError(err)
                return
            }

            for (var i = 0; i < data.length; i++) {
                var m = data[i]
                m.selected = false
                model.append(m)
            }
        })
    }

    function clear() {
        model.clear()
    }

    // Append a file to the grid
    function append(m) {
        m.selected = false
        model.append(m)
    }

    delegate: Rectangle {
        height: 155
        width: 155
        color: selected ? (SystemPalette.highlight
                           || "lightsteelblue") : "transparent"
        radius: 5
        Image {
            anchors {
                horizontalCenter: parent.horizontalCenter
                verticalCenter: parent.verticalCenter
            }
            asynchronous: true
            sourceSize: "150x150"
            source: Paths.thumb(sha1, thumb.is_png)
            focus: true
        }
    }

    Keys.onPressed: {
        if (event.modifiers & Qt.MetaModifier) {
            return event.accepted = true
        }

        // TODO: Page Up and Page Down
        if (event.modifiers & Qt.ControlModifier) {
            switch (event.key) {
            case Qt.Key_A:
                for (var i = 0; i < model.count; i++) {
                    model.get(i).selected = true
                }
                break
            }
        } else {
            switch (event.key) {
            case Qt.Key_Up:
                moveCurrentIndexUp()
                selectCurrent()
                break
            case Qt.Key_Down:
                moveCurrentIndexDown()
                selectCurrent()
                break
            case Qt.Key_Left:
                moveCurrentIndexLeft()
                selectCurrent()
                break
            case Qt.Key_Right:
                moveCurrentIndexRight()
                selectCurrent()
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
                browser.remove(selected())
                break
            }
        }
        event.accepted = true
    }

    MouseArea {
        id: mouseArea
        anchors.fill: parent
        acceptedButtons: Qt.LeftButton | Qt.RightButton
        drag.target: parent

        onReleased: {
            forceActiveFocus()
            var i = indexAt(mouse.x, mouse.y + contentY)
            var multiple = !!(mouse.modifiers & Qt.ControlModifier)
            if (mouse.button === Qt.RightButton) {
                var m = model.get(i)
                if (m && !m.selected) {
                    select(i, multiple)
                }
                showMenu()
            } else if (mouse.modifiers & Qt.ShiftModifier && i > 0
                       && i !== currentIndex) {
                if (!(mouse.modifiers & Qt.ControlModifier)) {
                    clearSelected()
                }

                var target = i
                for (; i !== currentIndex; target < currentIndex ? i++ : i--) {
                    m = model.get(i)
                    m.selected = !m.selected
                }

                // Always select the starting position
                model.get(currentIndex).selected = true

                positionViewAtIndex(target, GridView.Contain)
                currentIndex = target
            } else {
                select(i, multiple)
            }
        }

        onDoubleClicked: {
            open(indexAt(mouse.x, mouse.y + contentY))
        }
    }

    //    Drag.active: url ? mouseArea.drag.active : false
    //    Drag.dragType: Drag.Automatic
    //    Drag.supportedActions: Qt.CopyAction
    //    Drag.mimeData: {
    //        "text/uri-list": url
    //    }

    // Select and highlight a file. Optionally allow multiple selection.
    function select(i, multiple) {
        currentIndex = i

        if (!multiple || i === -1) {
            clearSelected()
        }

        if (i !== -1) {
            var m = model.get(i)
            if (multiple && m.selected) {
                m.selected = false
            } else {
                m.selected = true
            }
            positionViewAtIndex(i, GridView.Contain)
        }
    }

    // Select the currently positioned over item
    function selectCurrent() {
        clearSelected()
        model.get(currentIndex).selected = true
    }

    function clearSelected() {
        for (var j = 0; j < model.count; j++) {
            model.get(j).selected = false
        }
    }

    // Open a thumbnail for full screeen viewing
    function open(i) {
        fileView.render(model.get(i).sha1)
    }

    // Return selected files as an array of IDs
    function selected() {
        var ids = []
        for (var i = 0; i < model.count; i++) {
            if (model.get(i).selected) {
                ids.push(i)
            }
        }
        return ids
    }

    // Show context menu for selected files
    function showMenu() {
        Qt.createComponent("FileMenu.qml").createObject(fileView, {
                                                            ids: selected()
                                                        }).popup()
    }

    // Remove a file from the database and update UI accordingly
    function remove(indeces) {
        for (var i = 0; i < indeces.length; i++) {
            // Compensate for index drift
            var index = indeces[i] - i
            var m = model.get(index)
            HTTP.send("/images/" + m.sha1, "DELETE", undefined,
                      function (res, err) {
                          if (err) {
                              displayError(err)
                          }
                      })
            model.remove(index)
        }
    }
}
