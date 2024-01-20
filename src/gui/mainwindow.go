package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

type stApp struct {
	app    *fyne.App
	window *fyne.Window // 主窗口.
}

func Mainwindow() {
	application := app.New()
	window := application.NewWindow("protocolgo")

	var app stApp
	app.app = &application
	app.window = &window

	app.MakeUI()

	window.ShowAndRun()
}

func (stapp *stApp) MakeUI() {
	// 设置窗口大小
	(*stapp.window).Resize(fyne.NewSize(800, 600))
	// 设置主题
	(*stapp.app).Settings().SetTheme(theme.DarkTheme())
	// 添加菜单
	stapp.CreateMenuItem()

}

// 添加菜单
func (stapp *stApp) CreateMenuItem() {

	// 打开菜单项
	openMenuItem := fyne.NewMenuItem("open..", func() {

	})
	// 创建一个一级菜单
	fileMenu := fyne.NewMenu("File", openMenuItem)
	// 创建菜单栏
	menu := fyne.NewMainMenu(fileMenu)

	(*stapp.window).SetMainMenu(menu)
}
