package main

import (
	"os"

	hcore "github.com/bakape/hydron/core"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/qml"
)

func main() {
	// Load hydron runtime
	if err := hcore.Init(); err != nil {
		panic(err)
	}
	defer hcore.ShutDown()

	core.QCoreApplication_SetAttribute(core.Qt__AA_EnableHighDpiScaling, true)
	gui.NewQGuiApplication(len(os.Args), os.Args)
	view := qml.NewQQmlApplicationEngine(nil)
	buildBridge(view)
	view.Load(core.NewQUrl3("qrc:///qml/main.qml", 0))
	gui.QGuiApplication_Exec()
}
