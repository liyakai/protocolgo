package gui

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"protocolgo/src/logic"
	"protocolgo/src/utils"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/flopp/go-findfont"
	"github.com/goki/freetype/truetype"
	"github.com/sirupsen/logrus"
)

type StApp struct {
	App     *fyne.App
	Window  *fyne.Window      // 主窗口.
	CoreMgr logic.CoreManager // 管理器
}

// 生成UI
func (stapp *StApp) MakeUI() {
	// 解决中文乱码
	// stapp.SuitForChinese()
	// 设置窗口大小
	(*stapp.Window).Resize(fyne.NewSize(1600, 1200))
	// 设置主题
	(*stapp.App).Settings().SetTheme(theme.DarkTheme())
	// 添加菜单
	stapp.CreateMenuItem()
	// 创建布局
	stapp.CreateMainContainer()
	// 设置退出策略
	stapp.SetOnClose()

	// 初始化管理器
	stapp.CoreMgr.Init()

}

// 解决中文乱码问题
func (stapp *StApp) SuitForChinese() {
	//设置中文字体
	fontPath, err := findfont.Find("SIMFANG.TTF")
	if err != nil {
		panic(err)
	}
	logrus.Info("Found 'SIMFANG.ttf' in ", fontPath)

	// load the font with the freetype library
	// 原作者使用的ioutil.ReadFile已经弃用
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		panic(err)
	}
	_, err = truetype.Parse(fontData)
	if err != nil {
		panic(err)
	}
	os.Setenv("FYNE_FONT", fontPath)
	os.Setenv("FYNE_FONT_MONOSPACE", fontPath)
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
		saveDialog.Resize(fyne.NewSize(1500, 1100))
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
			stapp.CoreMgr.ReadXmlFromReader(reader)
			logrus.Info("Open xml file done.file path:", xml_file_path)
		}, *stapp.Window)
		file_picker.Resize(fyne.NewSize(1500, 1100))
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
		container.NewTabItem("Enum", stapp.CreateEnumTab()),
		container.NewTabItem("Message", stapp.CreateMessageTab()),
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

// 创建枚举页签
func (stapp *StApp) CreateEnumTab() fyne.CanvasObject {
	// 创建一个列表的数据源
	stapp.CoreMgr.EnumTableList = binding.NewStringList()

	// 创建一个列表
	list := widget.NewListWithData(stapp.CoreMgr.EnumTableList,
		func() fyne.CanvasObject {
			label := &TableListLabel{}
			label.ExtendBaseWidget(label)
			return label
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*TableListLabel).Bind(i.(binding.String))
			o.(*TableListLabel).data = i.(binding.String)
			o.(*TableListLabel).app = stapp
			o.(*TableListLabel).tabletype = logic.TableType_Enum
		},
	)

	// 使用垂直布局将上部和下部容器组合在一起
	buttonwithlist := container.NewBorder(
		stapp.CreateEnumTabListInstruction(),
		nil,
		nil,
		nil,
		container.NewStack(list),
	)

	return buttonwithlist
}

// 创建Message页签
func (stapp *StApp) CreateMessageTab() fyne.CanvasObject {
	// 创建一个列表的数据源
	stapp.CoreMgr.MessageTableList = binding.NewStringList()

	// 创建一个列表
	list := widget.NewListWithData(stapp.CoreMgr.MessageTableList,
		func() fyne.CanvasObject {
			label := &TableListLabel{}
			label.ExtendBaseWidget(label)
			return label
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*TableListLabel).Bind(i.(binding.String))
			o.(*TableListLabel).data = i.(binding.String)
			o.(*TableListLabel).app = stapp
			o.(*TableListLabel).tabletype = logic.TableType_Message
		},
	)

	// 使用垂直布局将上部和下部容器组合在一起
	buttonwithlist := container.NewBorder(
		stapp.CreateMessageTabListInstruction(),
		nil,
		nil,
		nil,
		container.NewStack(list),
	)

	return buttonwithlist
}

// 创建枚举的list的说明和button
func (stapp *StApp) CreateEnumTabListInstruction() fyne.CanvasObject {
	label := widget.NewLabel("enum list:")
	button := widget.NewButton("Add new", func() {
		// label.SetText("按钮被按下了!")
		stapp.CreateNewEnumeUnit()
		// stapp.CoreMgr.AddNewEnum("NewEnum")
	})
	// 使用HBox将searchEntry和searchButton安排在同一行，并使用HSplit来设置比例
	topContainer := container.NewHSplit(container.NewStack(label), button)
	topContainer.Offset = 0.75 //设置searchEntry 占 3/4， searchButton 占 1/4
	return topContainer
}

// 创建Message的list的说明和button
func (stapp *StApp) CreateMessageTabListInstruction() fyne.CanvasObject {
	label := widget.NewLabel("message list:")
	button := widget.NewButton("Add new", func() {
		// label.SetText("按钮被按下了!")
		stapp.CreateNewMessageUnit()
		// stapp.CoreMgr.AddNewEnum("NewEnum")
	})
	// 使用HBox将searchEntry和searchButton安排在同一行，并使用HSplit来设置比例
	topContainer := container.NewHSplit(container.NewStack(label), button)
	topContainer.Offset = 0.75 //设置searchEntry 占 3/4， searchButton 占 1/4
	return topContainer
}

// 检查 EditEnum
func (stapp *StApp) CheckEditEnum(editMsg logic.EditEnum) bool {
	// 检查 message name 的合法性
	if editMsg.EnumName == "" || strings.Contains(editMsg.EnumName, " ") || utils.CheckPositiveInteger(editMsg.EnumName) || utils.CheckStartWithNum(editMsg.EnumName) {
		logrus.Error("CheckEditEnum failed. EnumName: ", editMsg.EnumName)
		dialog.ShowInformation("Error!", "The message name is invalid", *stapp.Window)
		return false
	}

	for _, rowComponents := range editMsg.RowList {
		// 检查 EntryIndex 的合法性
		if rowComponents.EntryIndex.Text == "" || !utils.CheckPositiveInteger(rowComponents.EntryIndex.Text) {
			logrus.Error("CheckEditMessage failed. EntryIndex: ", rowComponents.EntryIndex.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryIndex is invalid", *stapp.Window)
			return false
		}

	}

	if !logic.CheckEnumFieldNameList(editMsg.RowList) {
		dialog.ShowInformation("Error!", " The field name ara duplicate", *stapp.Window)
		return false
	}

	if !logic.CheckEnumFieldIndexList(editMsg.RowList) {
		dialog.ShowInformation("Error!", " The field index ara duplicate", *stapp.Window)
		return false
	}
	return true
}

// 检查 EditMessage
func (stapp *StApp) CheckEditMessage(editMsg logic.EditMessage) bool {
	// 检查 message name 的合法性
	if editMsg.MsgName == "" || strings.Contains(editMsg.MsgName, " ") || utils.CheckPositiveInteger(editMsg.MsgName) || utils.CheckStartWithNum(editMsg.MsgName) {
		logrus.Error("CheckEditMessage failed. MsgName: ", editMsg.MsgName)
		dialog.ShowInformation("Error!", "The message name is invalid", *stapp.Window)
		return false
	}

	for _, rowComponents := range editMsg.RowList {
		// 检查 EntryIndex 的合法性
		if rowComponents.EntryIndex.Text == "" || !utils.CheckPositiveInteger(rowComponents.EntryIndex.Text) {
			logrus.Error("CheckEditMessage failed. EntryIndex: ", rowComponents.EntryIndex.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryIndex is invalid", *stapp.Window)
			return false
		}
		// 检查 key 的合法性
		if rowComponents.EntryKey.Text == "" || strings.Contains(rowComponents.EntryKey.Text, " ") {
			logrus.Error("CheckEditMessage failed. EntryKey: ", rowComponents.EntryKey.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryKey is invalid", *stapp.Window)
			return false
		}
		// 检查 value 的合法性
		if rowComponents.EntryValue.Text == "" || strings.Contains(rowComponents.EntryValue.Text, " ") {
			logrus.Error("CheckEditMessage failed. EntryValue: ", rowComponents.EntryValue.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryValue is invalid", *stapp.Window)
			return false
		}
	}

	if !logic.CheckMessageFieldNameList(editMsg.RowList) {
		dialog.ShowInformation("Error!", " The field name ara duplicate", *stapp.Window)
		return false
	}

	if !logic.CheckMessageFieldIndexList(editMsg.RowList) {
		dialog.ShowInformation("Error!", " The field index ara duplicate", *stapp.Window)
		return false
	}
	return true
}

// 创建新Enum的编辑页面
func (stapp *StApp) CreateNewEnumeUnit() {
	dialogContent := container.NewVBox()

	customDialog := dialog.NewCustomWithoutButtons("Edit Enum", container.NewVScroll(dialogContent), *stapp.Window)
	customDialog.Resize(fyne.NewSize(1500, 1100))
	// 创建输入信息的容器
	inputInfoContainer := container.NewVBox()
	// 创建输入框
	inputUnitName := widget.NewEntry()
	inputUnitName.SetPlaceHolder("Enter Enum name...")
	inputInfoContainer.Add(inputUnitName)

	nEntryIndex := 0
	// 在外部定义一个列表来保存每一行的组件
	var rowList []logic.EditRowEnum

	// 创建一个"Add" 按钮，点击后在VBox中添加新的Entry
	attrBox := container.NewVBox()

	// 增加字段
	addButton := widget.NewButton("Add field", func() {
		nEntryIndex = nEntryIndex + 1

		entryIndex := widget.NewEntry()
		entryIndex.SetText(strconv.Itoa(nEntryIndex))
		entryName := widget.NewEntry()
		entryName.SetPlaceHolder("Enter name...")

		entryKeySelect := xwidget.NewCompletionEntry([]string{})
		// 设置默认值
		// When the use typed text, complete the list.
		entryKeySelect.OnChanged = func(s string) {
			// completion start for text length >= 3
			if len(s) < 3 {
				entryKeySelect.HideCompletion()
				return
			}

			// Make a search on wikipedia
			resp, err := http.Get(
				"https://en.wikipedia.org/w/api.php?action=opensearch&search=" + entryKeySelect.Text,
			)
			if err != nil {
				entryKeySelect.HideCompletion()
				return
			}

			// Get the list of possible completion
			var results [][]string
			json.NewDecoder(resp.Body).Decode(&results)

			// no results
			if len(results) == 0 {
				entryKeySelect.HideCompletion()
				return
			}

			// then show them
			entryKeySelect.SetOptions(results[1])
			entryKeySelect.ShowCompletion()
		}

		oneRow := container.NewHSplit(entryName, entryIndex)
		oneRow.Offset = 0.95

		// 创建一个新的RowComponents实例并保存到列表中,加入列表,方便获取数值
		editRow := logic.EditRowEnum{
			EntryIndex: entryIndex,
			EntryName:  entryName,
		}
		var deleteFunc func() // 声明删除操作函数
		// 在每一行添加一个"删除"按钮
		deleteButton := widget.NewButton("Delete", func() {
			if deleteFunc != nil {
				deleteFunc()
			}
		})
		oneRowWithDeleteButton := container.NewHSplit(deleteButton, oneRow)
		oneRowWithDeleteButton.Offset = 0.01

		// 将删除操作定义为一个独立的函数
		deleteFunc = func() {
			// 从rowList和attrBox中移除该行
			rowList = editRow.RemoveElementFromSlice(rowList, editRow)
			attrBox.Remove(oneRowWithDeleteButton)
			attrBox.Refresh()
			customDialog.Refresh()
		}

		attrBox.Add(oneRowWithDeleteButton)
		attrBox.Refresh()

		rowList = append(rowList, editRow)

	})

	// 创建可以新增列的container
	attrBoader := container.NewBorder(nil, nil, nil, addButton, attrBox)
	inputInfoContainer.Add(attrBoader)

	// 增加关闭,保存按钮
	buttons := container.NewHBox(
		// Cancel Button
		widget.NewButton("Cancel", func() {
			// Cancel logic goes here
			customDialog.Hide()
		}),
		// Save Button
		widget.NewButton("Save", func() {
			// Save logic goes here
			logrus.Info("[CreateNewEnum]Save. inputUnitName: " + inputUnitName.Text)

			var editMsg logic.EditEnum
			editMsg.EnumName = inputUnitName.Text
			editMsg.RowList = rowList
			if !stapp.CheckEditEnum(editMsg) {
				logrus.Error("[CreateNewEnum] CheckEditMessage failed.")
				return
			}

			if !stapp.CoreMgr.AddNewEnum(editMsg) {
				logrus.Error("[CreateNewEnum] AddNewEnum failed.")
				return
			}

			customDialog.Hide()
		}),
	)
	inputInfoContainer.Add(container.NewCenter(buttons))
	dialogContent.Add(inputInfoContainer)
	customDialog.Show()

}

// 创建新Message的编辑页面
func (stapp *StApp) CreateNewMessageUnit() {
	dialogContent := container.NewVBox()

	customDialog := dialog.NewCustomWithoutButtons("Edit Message", container.NewVScroll(dialogContent), *stapp.Window)
	customDialog.Resize(fyne.NewSize(1500, 1100))
	// 创建输入信息的容器
	inputInfoContainer := container.NewVBox()
	// 创建输入框
	inputUnitName := widget.NewEntry()
	inputUnitName.SetPlaceHolder("Enter Message name...")
	inputInfoContainer.Add(inputUnitName)

	nEntryIndex := 0
	// 在外部定义一个列表来保存每一行的组件
	var rowList []logic.EditRowMessage

	// 创建一个"Add" 按钮，点击后在VBox中添加新的Entry
	attrBox := container.NewVBox()

	// 增加字段
	addButton := widget.NewButton("Add field", func() {
		nEntryIndex = nEntryIndex + 1
		selectEntryType := widget.NewSelect([]string{"optional", "repeated"}, nil)
		selectEntryType.Selected = "optional"
		entryIndex := widget.NewEntry()
		entryIndex.SetText(strconv.Itoa(nEntryIndex))
		entryKey := widget.NewEntry()
		entryKey.SetPlaceHolder("Enter type...")
		entryValue := widget.NewEntry()
		entryValue.SetPlaceHolder("Enter variable name...")

		entryKeySelect := xwidget.NewCompletionEntry([]string{})
		// 设置默认值
		// When the use typed text, complete the list.
		entryKeySelect.OnChanged = func(s string) {
			// completion start for text length >= 3
			if len(s) < 3 {
				entryKeySelect.HideCompletion()
				return
			}

			// Make a search on wikipedia
			resp, err := http.Get(
				"https://en.wikipedia.org/w/api.php?action=opensearch&search=" + entryKeySelect.Text,
			)
			if err != nil {
				entryKeySelect.HideCompletion()
				return
			}

			// Get the list of possible completion
			var results [][]string
			json.NewDecoder(resp.Body).Decode(&results)

			// no results
			if len(results) == 0 {
				entryKeySelect.HideCompletion()
				return
			}

			// then show them
			entryKeySelect.SetOptions(results[1])
			entryKeySelect.ShowCompletion()
		}

		oneRowInfo := container.NewHSplit(selectEntryType, container.NewHSplit(entryKey, entryValue))
		oneRowInfo.Offset = 0.05
		oneRow := container.NewHSplit(oneRowInfo, entryIndex)
		oneRow.Offset = 0.95

		// 创建一个新的RowComponents实例并保存到列表中,加入列表,方便获取数值
		editRow := logic.EditRowMessage{
			EntryIndex: entryIndex,
			EntryType:  selectEntryType,
			EntryKey:   entryKey,
			EntryValue: entryValue,
		}
		var deleteFunc func() // 声明删除操作函数
		// 在每一行添加一个"删除"按钮
		deleteButton := widget.NewButton("Delete", func() {
			if deleteFunc != nil {
				deleteFunc()
			}
		})
		oneRowWithDeleteButton := container.NewHSplit(deleteButton, oneRow)
		oneRowWithDeleteButton.Offset = 0.01

		// 将删除操作定义为一个独立的函数
		deleteFunc = func() {
			// 从rowList和attrBox中移除该行
			rowList = editRow.RemoveElementFromSlice(rowList, editRow)
			attrBox.Remove(oneRowWithDeleteButton)
			attrBox.Refresh()
			customDialog.Refresh()
		}

		attrBox.Add(oneRowWithDeleteButton)
		attrBox.Refresh()

		rowList = append(rowList, editRow)

	})

	// 创建可以新增列的container
	attrBoader := container.NewBorder(nil, nil, nil, addButton, attrBox)
	inputInfoContainer.Add(attrBoader)

	// 增加关闭,保存按钮
	buttons := container.NewHBox(
		// Cancel Button
		widget.NewButton("Cancel", func() {
			// Cancel logic goes here
			customDialog.Hide()
		}),
		// Save Button
		widget.NewButton("Save", func() {
			// Save logic goes here
			logrus.Info("[CreateNewMessage]Save. inputUnitName: " + inputUnitName.Text)

			var editMsg logic.EditMessage
			editMsg.MsgName = inputUnitName.Text
			editMsg.RowList = rowList
			if !stapp.CheckEditMessage(editMsg) {
				logrus.Error("[CreateNewMessage] CheckEditMessage failed.")
				return
			}

			if !stapp.CoreMgr.AddNewMessage(editMsg) {
				logrus.Error("[CreateNewMessage] AddNewMessage failed.")
				return
			}

			customDialog.Hide()
		}),
	)
	inputInfoContainer.Add(container.NewCenter(buttons))
	dialogContent.Add(inputInfoContainer)
	customDialog.Show()

}
