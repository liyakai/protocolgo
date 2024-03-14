package main

import (
	"flag"
	"protocolgo/src/gui"
	"protocolgo/src/logic"
	"protocolgo/src/utils"

	"fyne.io/fyne/v2/app"
)

func main() {
	InitMainWindow()
}

func InitMainWindow() {
	application := app.New()
	window := application.NewWindow("protocolgo")

	var app gui.StApp
	app.Window = &window
	app.App = &application
	app.CoreMgr = logic.CoreManager{}

	// 初始化日志
	utils.InitLogger(*flag.String("loglevel", "info", "sets log level. trace/debug/info/warn/error/fatal/panic"))

	app.MakeUI()

	window.ShowAndRun()
}

func InitConfig(bDebug *bool) {
}
