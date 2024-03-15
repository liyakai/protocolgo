package gui

import (
	"log"
	"protocolgo/src/logic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

type StApp struct {
	App     *fyne.App
	Window  *fyne.Window      // 主窗口.
	CoreMgr logic.CoreManager // 管理器
}

// 生成UI
func (stapp *StApp) MakeUI() {
	// 设置窗口大小
	(*stapp.Window).Resize(fyne.NewSize(800, 600))
	// 设置主题
	(*stapp.App).Settings().SetTheme(theme.DarkTheme())
	// 添加菜单
	stapp.CreateMenuItem()
	// 创建布局
	stapp.CreateMainContainer()
	// 设置退出策略
	stapp.SetOnClose()

}

// 自定义退出策略
func (stapp *StApp) SetOnClose() {
	// 设置自定义的退出策略
	(*stapp.Window).SetCloseIntercept(func() {
		// 创建并显示确认对话框
		exitConfirm := dialog.NewConfirm("Exit",
			"Are you sure you want to exit the application?",
			func(response bool) {
				if response {
					// 先尝试保存/关闭文件
					stapp.CoreMgr.CloseCurrXmlFile()
					logrus.Info("User closed this app.")
					(*stapp.App).Quit() // 如果用户点击 “Yes”，则退出程序
				}
			},
			*stapp.Window,
		)
		exitConfirm.SetDismissText("No")  // 设置确认对话框的取消按钮文字
		exitConfirm.SetConfirmText("Yes") // 设置确认对话框的确认按钮文字
		exitConfirm.Show()                // 显示确认对话框

		exitConfirm.Show()
	})
}

// 添加菜单
func (stapp *StApp) CreateMenuItem() {

	// 创建新文件
	newMenuItem := fyne.NewMenuItem("new..", func() {
		// 打开文件
		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				logrus.Error("Failed to NewFileSave:", err)
				dialog.ShowError(err, *stapp.Window)
				return
			}
			if writer == nil { // user cancelled
				logrus.Info("Failed to NewFileSave, user cancelled.")
				return
			}
			xml_file_path := writer.URI().Path()
			// 设定 当前打开的xml文件路径
			stapp.CoreMgr.SetCurrXmlFilePath(xml_file_path)
			stapp.CoreMgr.CreateNewXml()
			// 在此处写入你的文件
		}, *stapp.Window)
		saveDialog.Resize(fyne.NewSize(700, 500))
		saveDialog.SetFileName("protocolgo.xml")
		saveDialog.Show()

	})
	// 打开菜单项
	openMenuItem := fyne.NewMenuItem("open..", func() {
		// 打开文件
		file_picker := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			// Check for errors
			if err != nil {
				log.Println("Failed to NewFileOpen:", err)
				return
			}
			xml_file_path := reader.URI().Path()
			// 设定 当前打开的xml文件路径
			stapp.CoreMgr.SetCurrXmlFilePath(xml_file_path)
			// 读取到内存
			stapp.CoreMgr.ReadXmlFromFile(reader)
			logrus.Info("Open xml file done.file path:", xml_file_path)
		}, *stapp.Window)
		file_picker.Resize(fyne.NewSize(700, 500))
		file_picker.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".xml"}))
		// 显示打开文件的UI
		file_picker.Show()

	})
	// 保存菜单项
	saveMenuItem := fyne.NewMenuItem("save..", func() {
		dialog.ShowConfirm("Confirmation", "Are you sure you want to Save?",
			func(response bool) {
				if response {
					// addition logic to save file goes here
					if stapp.CoreMgr.SaveToXmlFile() {
						dialog.ShowInformation("Saved", "File successfully saved.", *stapp.Window)
					} else {
						dialog.ShowInformation("Error", "Unexist opened xml or invalid xml path.", *stapp.Window)
					}

				}
			}, *stapp.Window)

	})
	// 创建一个一级菜单
	fileMenu := fyne.NewMenu("File", newMenuItem, openMenuItem, saveMenuItem)
	// 创建菜单栏
	menu := fyne.NewMainMenu(fileMenu)

	(*stapp.Window).SetMainMenu(menu)
}

// 创建主体布局
func (stapp *StApp) CreateMainContainer() {
	// 创建上部容器
	topContainer := stapp.CreateTopSearchContainer()

	// 创建下部的标签页容器
	tabs := container.NewAppTabs(
		container.NewTabItem("Tab 1", stapp.CreateTab1()),
		container.NewTabItem("Tab 2", container.NewVBox(widget.NewLabel("This is tab 2 content"))),
	)

	// 使用垂直布局将上部和下部容器组合在一起
	mainContainer := container.NewBorder(
		topContainer,
		nil,
		nil,
		nil,
		container.NewStack(tabs),
	)
	(*stapp.Window).SetContent(mainContainer)
}

// 创建搜索框的内容
func (stapp *StApp) CreateTopSearchContainer() fyne.CanvasObject {
	// 创建一个输入框作为搜索框
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search here")

	searchButton := widget.NewButton("Search", func() {
		// 在这里添加搜索按钮点击后的逻辑，例如打印输入的搜索内容
		input := searchEntry.Text
		// 逻辑处理处
		// fyne.LogInfo("You have searched for:", input)
		logrus.Info("You have searched for:", input)
	})
	// 使用HBox将searchEntry和searchButton安排在同一行，并使用HSplit来设置比例
	topContainer := container.NewHSplit(container.NewStack(searchEntry), searchButton)
	topContainer.Offset = 0.75 //设置searchEntry 占 3/4， searchButton 占 1/4
	return topContainer
}

// 创建页签1的内容
func (stapp *StApp) CreateTab1() fyne.CanvasObject {
	// 创建一个列表的数据源
	stapp.CoreMgr.Table1List = binding.NewStringList()

	// 添加列表数据
	stapp.CoreMgr.Table1List.Append("Item 1")
	stapp.CoreMgr.Table1List.Append("Item 2")
	stapp.CoreMgr.Table1List.Append("Item 3")

	// 创建一个列表
	list := widget.NewListWithData(stapp.CoreMgr.Table1List,
		func() fyne.CanvasObject {
			label := &TableListLabel{}
			label.ExtendBaseWidget(label)
			return label
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*TableListLabel).Bind(i.(binding.String))
			o.(*TableListLabel).data = i.(binding.String)
			o.(*TableListLabel).window = stapp.Window
		},
	)
	return list
}
