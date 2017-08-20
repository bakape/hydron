import QtQuick 2.9
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.3
import QtMultimedia 5.9

Rectangle {
    property variant model

    visible: false
    color: SystemPalette.base || "white"
    focus: true

    SplitView {
        anchors.fill: parent

        ScrollView {
            Layout.preferredWidth: 200
            Layout.minimumWidth: 0
            Layout.maximumWidth: parent.width * 0.7
            width: 200

            ListView {
                id: tags
                anchors.fill: parent
                interactive: false
                boundsBehavior: Flickable.StopAtBounds
                model: ListModel {}
                delegate: Text {
                    text: tag
                }
            }
        }

        Rectangle {
            property string url

            id: display
            Layout.fillWidth: true

            Text {
                id: error
                visible: false
                anchors.fill: parent
                anchors.centerIn: parent
            }

            Image {
                id: image
                visible: false
                asynchronous: true
                fillMode: Image.PreserveAspectFit
                anchors.fill: parent
            }

            AnimatedImage {
                id: animated
                visible: false
                asynchronous: true
                fillMode: Image.PreserveAspectFit
                anchors.fill: parent
            }

            // MediaPlayer can not be hidden. Need to contain it.
            Rectangle {
                id: mediaContainer
                visible: false
                anchors.fill: parent

                // TODO: Media controls

                MediaPlayer {
                    id: media
                    loops: MediaPlayer.Infinite
                    autoPlay: true
                    autoLoad: true
                }

                VideoOutput {
                    anchors.fill: parent
                    source: media
                }

                MouseArea {
                    id: playArea
                    anchors.fill: parent
                    onReleased: {
                        if (media.playbackState === MediaPlayer.PlayingState) {
                            media.pause()
                        } else {
                            media.play()
                        }
                    }
                }
            }

            MouseArea {
                id: mouseArea
                anchors.fill: parent
                acceptedButtons: Qt.RightButton | Qt.LeftButton
                onClicked: parent.showMenu()
                drag.target: display
            }

            Drag.active: mouseArea.drag.active
            Drag.dragType: Drag.Automatic
            Drag.supportedActions: Qt.CopyAction
            Drag.mimeData: {
                "text/uri-list": display.url
            }

            function showMenu() {
                Qt.createComponent("FileMenu.qml")
                    .createObject(fileView, { ids: [ fileView.model.sha1 ] })
                    .popup()
            }
        }
    }

    Keys.onPressed: {
        switch (event.key) {
        case Qt.Key_Escape:
            remove()
            break
        case Qt.Key_Delete:
            window.removeFiles([model.sha1])
            // TODO: Navigate to next file
            remove()
            break
        }

        // TODO: Keyboard navigation in this mode

    }

    function render(id) {
        tags.model.clear()
        fileView.visible = true
        forceActiveFocus()
        window.header.visible = false
        browser.visible = false

        // Fetch more detailed record struct
        var data = JSON.parse(go.get(id))
        fileView.model = data
        if (data === null) {
            error.visible = true
            error.text = "File not found"
            display.url = null
            return
        }

        display.url = data.sourcePath
        switch (data.type) {
        case "jpg":
        case "png":
        case "webp":
        case "tiff":
        case "ico":
        case "bmp":
            image.visible = true
            image.source = data.sourcePath
            break
        case "gif":
            animated.visible = true
            animated.source = data.sourcePath
            break
        case "webm":
        case "ogg":
        case "mkv":
        case "mp4":
        case "avi":
        case "mov":
        case "wmv":
        case "flv":
            mediaContainer.visible = true
            media.source = data.sourcePath
            break
        case "mp3":
        case "aac":
        case "wave":
        case "flac":
        case "midi":
            image.visible = true
            image.source = data.thumPath
            mediaContainer.visible = true
            media.source = data.sourcePath
            break
        default: // PSD, PDF and others
            error.visible = true
            error.text = "Preview not available for this file type"
            return
        }

        for (var i = 0; i < data.tags.length; i++) {
            tags.model.append({tag: data.tags[i]})
        }
    }

    // Reset to empty state
    function empty() {
        browser.visible = true
        visible = false
        window.header.visible = true

        tags.model.clear()

        error.visible = false

        animated.source = ""
        animated.visible = false

        image.source = ""
        image.visible = false

        mediaContainer.visible = false
        media.source = ""
    }

    // Reset to empty state and hide overlay
    function remove() {
        empty()
        browser.forceActiveFocus()
    }
}
