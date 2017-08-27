import QtQuick 2.9
import QtQml 2.2
import QtQuick.Layouts 1.3
import QtQuick.Controls 2.2

TextField {
    Layout.fillWidth: true
    placeholderText: "Search"
    focus: true
    activeFocusOnTab: true

    onAccepted: {
        closeMenu()
        fileView.empty()
        browser.loadThumbnails(text)
    }
    Component.onCompleted: forceActiveFocus()
    onTextChanged: {
        if (!text.length || text[text.length -1 ] === " ") {
            closeMenu()
            return
        }

        var i = text.lastIndexOf(" ")
        var last = i === -1 ? text : text.slice(i + 1)
        var tags = JSON.parse(go.completeTag(last))
        suggestions.model.clear()
        if (tags.length) {
            for (i = 0; i < tags.length; i++) {
                suggestions.model.append({tag: tags[i]})
            }
            menu.focus = false
            menu.isOpen = true
            menu.open()
        }
    }

    Menu {
        id: menu
        property bool isOpen: false
        closePolicy: Popup.CloseOnPressOutsideParent | Popup.CloseOnEscape
        dim: false
        y: searchBar.height
        focus: false

        Instantiator {
            id: suggestions
            model: ListModel {}
            onObjectAdded: menu.addItem(object)
            onObjectRemoved: menu.removeItem(index)
            delegate: MenuItem{
                text: tag
                onTriggered: append(tag)
                Keys.onReturnPressed: triggered()
            }
        }
    }

    Keys.onPressed: {
        if (!menu.isOpen) {
            return
        }

        switch (event.key) {
        case Qt.Key_Down:
        case Qt.Key_Tab:
            event.accepted = true
            menu.focus = true
            menu.forceActiveFocus()
            break
        }
    }

    function closeMenu() {
        menu.close()
        menu.isOpen = false
    }

    function append(tag) {
        var split = text.trim().split(" ")
        if (!split.length) {
            text = tag
        } else {
            split[split.length - 1] = tag
            text = split.join(" ")
        }
        menu.close()
        forceActiveFocus()
    }
}
