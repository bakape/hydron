import QtQuick 2.7
import QtQuick.Layouts 1.3
import QtQuick.Controls 1.4

TextField {
    Layout.fillWidth: true
    placeholderText: "Search"
    focus: true
    activeFocusOnTab: true

    onAccepted: {
        fileView.empty()
        browser.loadThumbnails(text)
    }
    Component.onCompleted: {
        forceActiveFocus()
    }

    Keys.onReleased: {
        suggestions.model.clear()
        var mod = event.modifiers
        if (mod & Qt.ControlModifier
                || mod & Qt.AltModifier
                || mod & Qt.MetaModifier) {
            return
        }
        switch (event.key) {
        case Qt.Key_Return:
        case Qt.Key_Escape:
        case Qt.Key_Tab:
            return
        }

        var data = JSON.parse(go.completeTag(text))
        for (var i = 0; i < data.length; i++) {
            suggestions.model.append({tag: data[i]})
        }
    }
}
