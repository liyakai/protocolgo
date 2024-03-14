package gui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
)

type StApp struct {
	App    *fyne.App
	Window *fyne.Window // 主窗口.
}

// 生成UI
func (stapp *StApp) MakeUI() {
	// 设置窗口大小
	(*stapp.Window).Resize(fyne.NewSize(800, 600))
	// 设置主题
	(*stapp.App).Settings().SetTheme(theme.DarkTheme())
	// 添加菜单
	stapp.CreateMenuItem()

}

// 添加菜单
func (stapp *StApp) CreateMenuItem() {

	// 创建新文件
	newMenuItem := fyne.NewMenuItem("new..", func() {
		// stapp.utils.ShowDialog("Debug" "open xml")
		// 打开文件
		file_picker := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			// Check for errors
			if err != nil {
				log.Println("Failed to open file:", err)
				return
			}
		}, *stapp.Window)
		file_picker.Resize(fyne.NewSize(700, 500))
		file_picker.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".xml"}))
		// 显示打开文件的UI
		file_picker.Show()

	})
	// 打开菜单项
	openMenuItem := fyne.NewMenuItem("open..", func() {
		// stapp.utils.ShowDialog("Debug" "open xml")
		// 打开文件
		file_picker := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			// Check for errors
			if err != nil {
				log.Println("Failed to open file:", err)
				return
			}
		}, *stapp.Window)
		file_picker.Resize(fyne.NewSize(700, 500))
		file_picker.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".xml"}))
		// 显示打开文件的UI
		file_picker.Show()

	})
	// 创建一个一级菜单
	fileMenu := fyne.NewMenu("File", newMenuItem, openMenuItem)
	// 创建菜单栏
	menu := fyne.NewMainMenu(fileMenu)

	(*stapp.Window).SetMainMenu(menu)
}
