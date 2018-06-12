import QtQuick 2.9
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.3
import QtMultimedia 5.9
import "http.js" as HTTP
import "paths.js" as Paths

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
                model: ListModel {
                }
                delegate: Text {
                    text: tag
                }
            }
        }

        Rectangle {
            property string url

            id: display
            Layout.fillWidth: true

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
            Drag.mimeData: {"text/uri-list": display.url}

            function showMenu() {
                Qt.createComponent("FileMenu.qml").createObject(fileView, {
                                                                    ids: [fileView.model.sha1]
                                                                }).popup()
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

        event.accepted = true;
    }

    function render(id) {
        tags.model.clear()
        fileView.visible = true
        forceActiveFocus()
        window.header.visible = false
        browser.visible = false

        // Fetch more detailed record struct
        HTTP.get("/images/" + id, function (data, err) {
            fileView.model = null
            if (data === null || err) {
                errorPopup.render(err || "File not found")
                return
            }

            var srcPath = Paths.source(data.sha1, data.type)
            display.url = srcPath
            switch (data.type) {
            case 0:
            case 1:
            case 3:
            case 7:
            case 5:
                image.visible = true
                image.source = srcPath
                break
            case 2:
                animated.visible = true
                animated.source = srcPath
                break
            case 9:
            case 8:
            case 10:
            case 11:
            case 12:
            case 13:
            case 14:
            case 15:
                mediaContainer.visible = true
                media.source = srcPath
                break
            default:
                // PSD, PDF and others
                errorPopup.render("Preview not available for this file type")
                return
            }

            fileView.model = data

            // TODO: Sort and group tags
            if (data.tags) {
                for (var i = 0; i < data.tags.length; i++) {
                    var t = data.tags[i]
                    tags.model.append({tag: t.tag})
                }
            }
        })
    }

    // Reset to empty state
    function empty() {
        browser.visible = true
        visible = false
        window.header.visible = true

        tags.model.clear()

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
