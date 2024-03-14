package main

import (
	"flag"
	"protocolgo/src/gui"
	"protocolgo/src/logic"

	"fyne.io/fyne/v2/app"
)

func main() {
	InitMainWindow()
}

func InitMainWindow() {
	application := app.New()
	window := application.NewWindow("protocolgo")

	var coremgr logic.CoreManager
	coremgr.Stapp = &gui.StApp{App: &application, Window: &window}
	coremgr.Utils = &logic.StUtils{CoreMgr: &coremgr}

	// 初始化日志
	logic.InitLogger(*flag.String("loglevel", "info", "sets log level. trace/debug/info/warn/error/fatal/panic"))

	coremgr.Stapp.MakeUI()

	window.ShowAndRun()
}

func InitConfig(bDebug *bool) {
}
