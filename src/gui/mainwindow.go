package gui

import (
	"log"
	"os"
	"protocolgo/src/logic"
	"protocolgo/src/utils"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"

	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"github.com/goki/freetype/truetype"
	"github.com/sirupsen/logrus"

	xwidget "fyne.io/x/fyne/widget"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type StApp struct {
	App     *fyne.App
	Window  *fyne.Window      // 主窗口.
	CoreMgr logic.CoreManager // 管理器
	tables  *container.AppTabs
}

// 生成UI
func (stapp *StApp) MakeUI() {
	// 初始化管理器
	stapp.CoreMgr.Init()

	// 解决中文乱码
	// stapp.SuitForChinese()
	// 设置窗口大小
	(*stapp.Window).Resize(fyne.NewSize(1200, 900))
	// 设置主题
	(*stapp.App).Settings().SetTheme(theme.DarkTheme())
	// 添加菜单
	stapp.CreateMenuItem()
	// 创建布局
	stapp.CreateMainContainer()
	// 设置退出策略
	stapp.SetOnClose()

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
		saveDialog.Resize(fyne.NewSize(1100, 800))
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
		file_picker.Resize(fyne.NewSize(1100, 800))
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
	stapp.tables = container.NewAppTabs(
		container.NewTabItem("Enum", stapp.CreateTab(logic.TableType_Enum)),
		container.NewTabItem("Message", stapp.CreateTab(logic.TableType_Message)),
	)

	// 使用垂直布局将上部和下部容器组合在一起
	mainContainer := container.NewBorder(
		topContainer,
		nil,
		nil,
		nil,
		container.NewStack(stapp.tables),
	)
	(*stapp.Window).SetContent(mainContainer)
}

// 创建搜索框的内容
func (stapp *StApp) CreateTopSearchContainer() fyne.CanvasObject {
	// 创建一个输入框作为搜索框
	// searchEntry := widget.NewEntry()
	// searchEntry.SetPlaceHolder("Search here")

	// searchButton := widget.NewButton("Search", func() {
	// 	// 在这里添加搜索按钮点击后的逻辑，例如打印输入的搜索内容
	// 	input := searchEntry.Text
	// 	// 逻辑处理处
	// 	// fyne.LogInfo("You have searched for:", input)
	// 	logrus.Info("You have searched for:", input)
	// })
	searchFields := stapp.CoreMgr.GetAllUseableEntryType()
	searchEntry := xwidget.NewCompletionEntry(searchFields)
	searchEntry.SetPlaceHolder("Search here")
	// 设置默认值
	// When the use typed text, complete the list.
	searchEntry.OnChanged = func(str string) {
		logrus.Debug("[CreateTopSearchContainer] OnChanged str:", str)
		if str != "" {
			matches := fuzzy.RankFind(str, searchFields)
			sort.Sort(matches)
			var strMatches []string
			for _, matchone := range matches {
				strMatches = append(strMatches, matchone.Target)
			}
			searchEntry.SetOptions(strMatches)
			// 设置焦点
			eTableType := stapp.CoreMgr.SyncTableListWithETree(str)
			logrus.Debug("[CreateTopSearchContainer] Forcuse table eTableType:", eTableType)
			if eTableType == logic.TableType_Enum {
				stapp.tables.SelectIndex(0)
			} else if eTableType == logic.TableType_Message {
				stapp.tables.SelectIndex(1)
			}
			searchEntry.ShowCompletion()
		} else {
			stapp.CoreMgr.SyncMessageListWithETree()
		}

	}
	searchEntry.OnSubmitted = func(str string) {
		// 设置焦点
		eTableType := stapp.CoreMgr.SyncTableListWithETree(str)
		logrus.Debug("[CreateTopSearchContainer] Forcuse table eTableType:", eTableType)
		if eTableType == logic.TableType_Enum {
			stapp.tables.SelectIndex(0)
		} else if eTableType == logic.TableType_Message {
			stapp.tables.SelectIndex(1)
		}
	}

	// 使用HBox将searchEntry和searchButton安排在同一行，并使用HSplit来设置比例
	// topContainer := container.NewHSplit(container.NewStack(searchEntry), searchButton)
	// topContainer.Offset = 0.75 //设置searchEntry 占 3/4， searchButton 占 1/4
	topContainer := container.NewStack(searchEntry)
	return topContainer
}

// 创建页签
func (stapp *StApp) CreateTab(tabletype logic.ETableType) fyne.CanvasObject {
	// 创建一个列表
	list := widget.NewListWithData(*stapp.CoreMgr.GetTableListByType(tabletype),
		func() fyne.CanvasObject {
			label := &TableListLabel{}
			label.ExtendBaseWidget(label)
			return label
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*TableListLabel).Bind(i.(binding.String))
			o.(*TableListLabel).data = i.(binding.String)
			o.(*TableListLabel).app = stapp
			o.(*TableListLabel).tabletype = tabletype
		},
	)

	// 使用垂直布局将上部和下部容器组合在一起
	buttonwithlist := container.NewBorder(
		stapp.CreateTabListInstruction(tabletype),
		nil,
		nil,
		nil,
		container.NewStack(list),
	)

	return buttonwithlist
}

// 创建list的说明和button
func (stapp *StApp) CreateTabListInstruction(tabletype logic.ETableType) fyne.CanvasObject {
	label := widget.NewLabel(stapp.CoreMgr.GetLableStingByType(tabletype))
	button := widget.NewButton("Add new", func() {
		// label.SetText("按钮被按下了!")
		stapp.EditUnit(tabletype, "")
		// stapp.CoreMgr.AddNewEnum("NewEnum")
	})
	// 使用HBox将searchEntry和searchButton安排在同一行，并使用HSplit来设置比例
	topContainer := container.NewHSplit(container.NewStack(label), button)
	topContainer.Offset = 0.75 //设置searchEntry 占 3/4， searchButton 占 1/4
	return topContainer
}

func (stapp *StApp) CreateRowForEditUnit(tabletype logic.ETableType, strRowUnit logic.StStrRowUnit, attrBox *fyne.Container, rowList *[]logic.StRowUnit) {
	var entryOption *widget.Select
	// var entryType *widget.Entry
	var entryTypeSelect *xwidget.CompletionEntry

	if tabletype == logic.TableType_Message {
		entryOption = widget.NewSelect([]string{"optional", "repeated"}, nil)
		if strRowUnit.EntryOption == "" {
			entryOption.Selected = "optional"
		} else {
			entryOption.Selected = strRowUnit.EntryOption
		}

		// entryType = widget.NewEntry()
		// entryType.SetPlaceHolder("Enter type...")
		// if strRowUnit.EntryType != "" {
		// 	entryType.SetText(strRowUnit.EntryType)
		// }

		searchFields := stapp.CoreMgr.GetAllUseableEntryType()
		entryTypeSelect = xwidget.NewCompletionEntry(searchFields)
		// 设置默认值
		// When the use typed text, complete the list.
		entryTypeSelect.OnChanged = func(str string) {
			// completion start for text length >= 3
			matches := fuzzy.RankFind(str, searchFields)
			sort.Sort(matches)
			var strMatches []string
			for _, matchone := range matches {
				strMatches = append(strMatches, matchone.Target)
			}
			entryTypeSelect.SetOptions(strMatches)
			entryTypeSelect.ShowCompletion()
		}

	}

	entryName := widget.NewEntry()
	entryName.SetPlaceHolder("Enter variable name...")
	if strRowUnit.EntryName != "" {
		entryName.SetText(strRowUnit.EntryName)
	}

	entryIndex := widget.NewEntry()
	entryIndex.SetText(strRowUnit.EntryIndex)

	var oneRow *container.Split
	if tabletype == logic.TableType_Message {
		oneRowInfo := container.NewHSplit(entryOption, container.NewHSplit(entryTypeSelect, entryName))
		oneRowInfo.Offset = 0.05
		oneRow = container.NewHSplit(oneRowInfo, entryIndex)
		oneRow.Offset = 0.95
	} else if tabletype == logic.TableType_Enum {
		oneRow = container.NewHSplit(entryName, entryIndex)
		oneRow.Offset = 0.95
	}

	// 创建一个新的RowComponents实例并保存到列表中,加入列表,方便获取数值
	stRow := logic.StRowUnit{
		EntryIndex:  entryIndex,
		EntryOption: entryOption,
		EntryType:   entryTypeSelect,
		EntryName:   entryName,
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
		*rowList = stRow.RemoveElementFromSlice(*rowList, stRow)
		attrBox.Remove(oneRowWithDeleteButton)
		attrBox.Refresh()
		// customDialog.Refresh()
	}
	attrBox.Add(oneRowWithDeleteButton)
	attrBox.Refresh()
	*rowList = append(*rowList, stRow)
	logrus.Info("[CreateRowForEditUnit] done. tabletype:", tabletype, ",strRowUnit:", strRowUnit)
}

// 创建新Message的编辑页面
func (stapp *StApp) EditUnit(tabletype logic.ETableType, unitname string) {
	dialogContent := container.NewVBox()
	customDialog := dialog.NewCustomWithoutButtons(stapp.CoreMgr.GetEditTableTitle(tabletype), container.NewVScroll(dialogContent), *stapp.Window)
	customDialog.Resize(fyne.NewSize(1100, 800))
	bCreateNew := false // 是否是新的节点
	etreeRow := stapp.CoreMgr.GetEtreeElem(tabletype, unitname)
	if unitname == "" || etreeRow == nil {
		bCreateNew = true
		logrus.Info("[EditUnit] Create new unit. tabletype:", tabletype, ",customDialog:", customDialog)
	} else {
		logrus.Info("[EditUnit] Edit old unit. tabletype:", tabletype, ", Tag:", etreeRow.Tag, ",customDialog:", customDialog)
	}

	// 创建输入信息的容器
	inputInfoContainer := container.NewVBox()
	// 创建输入框
	inputUnitName := widget.NewEntry()
	inputUnitName.SetPlaceHolder("Enter name...")
	if !bCreateNew {
		inputUnitName.SetText(unitname)
		inputUnitName.TextStyle.Bold = true
		inputUnitName.Disable()
	}
	inputInfoContainer.Add(inputUnitName)

	nEntryIndex := 0
	// 在外部定义一个列表来保存每一行的组件
	var rowList []logic.StRowUnit

	// 创建一个"Add" 按钮，点击后在VBox中添加新的Entry
	attrBox := container.NewVBox()

	// 先展示老的字段
	// 遍历子元素
	if !bCreateNew {
		for _, child := range etreeRow.ChildElements() {
			var rowUnit logic.StStrRowUnit
			entryOption := child.SelectAttr("EntryOption")
			if entryOption != nil {
				rowUnit.EntryOption = entryOption.Value
			}
			entryType := child.SelectAttr("EntryType")
			if entryType != nil {
				rowUnit.EntryType = entryType.Value
			}
			entryName := child.SelectAttr("EntryName")
			if entryName != nil {
				rowUnit.EntryName = entryName.Value
			}
			entryIndex := child.SelectAttr("EntryIndex")
			if entryIndex != nil {
				rowUnit.EntryIndex = entryIndex.Value
				index, err := strconv.Atoi(rowUnit.EntryIndex)
				if err == nil && index > nEntryIndex {
					nEntryIndex = index
				}
			}
			logrus.Info("[EditUnit] Add old row info to list. tabletype:", tabletype, ",rowUnit:", rowUnit)
			stapp.CreateRowForEditUnit(tabletype, rowUnit, attrBox, &rowList)
		}
	}

	// 增加新的字段
	addButton := widget.NewButton("Add field", func() {
		nEntryIndex = nEntryIndex + 1
		stapp.CreateRowForEditUnit(tabletype, logic.StStrRowUnit{
			EntryIndex: strconv.Itoa(nEntryIndex),
		}, attrBox, &rowList)
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

			var stUnit logic.StUnit
			stUnit.UnitName = inputUnitName.Text
			stUnit.TableType = tabletype
			stUnit.RowList = rowList
			if !stapp.CheckStUnit(stUnit, bCreateNew) {
				logrus.Error("[CreateNewMessage] CheckEditMessage failed.")
				return
			}

			if !stapp.CoreMgr.EditUnit(stUnit) {
				logrus.Error("[CreateNewMessage] EditUnit failed.")
				return
			}

			customDialog.Hide()
		}),
	)

	inputInfoContainer.Add(container.NewCenter(buttons))
	dialogContent.Add(inputInfoContainer)
	customDialog.Show()
}

// 检查 StUnit
func (stapp *StApp) CheckStUnit(stUnit logic.StUnit, bCreateNew bool) bool {
	// 检查 name 的合法性
	if stUnit.UnitName == "" || strings.Contains(stUnit.UnitName, " ") || utils.CheckPositiveInteger(stUnit.UnitName) || utils.CheckStartWithNum(stUnit.UnitName) {
		logrus.Error("CheckStUnit failed. invalid MsgName: ", stUnit.UnitName)
		dialog.ShowInformation("Error!", "The name is invalid", *stapp.Window)
		return false
	}

	// 检查 name 是否已经存在
	if bCreateNew && stapp.CoreMgr.CheckExistSameName(stUnit.UnitName) {
		logrus.Error("CheckStUnit failed. Repeated MsgName: ", stUnit.UnitName)
		dialog.ShowInformation("Error!", "The name["+stUnit.UnitName+"] is already exist.", *stapp.Window)
		return false
	}

	for _, rowComponents := range stUnit.RowList {
		// 检查 EntryIndex 的合法性
		if rowComponents.EntryIndex != nil && (rowComponents.EntryIndex.Text == "" || !utils.CheckPositiveInteger(rowComponents.EntryIndex.Text)) {
			logrus.Error("CheckStUnit failed. EntryIndex: ", rowComponents.EntryIndex.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryIndex is invalid", *stapp.Window)
			return false
		}
		// 检查 类型 的合法性
		if rowComponents.EntryType != nil && (rowComponents.EntryType.Text == "" || strings.Contains(rowComponents.EntryType.Text, " ") || rowComponents.EntryType.Text == stUnit.UnitName || utils.CheckPositiveInteger(rowComponents.EntryType.Text) || utils.CheckStartWithNum(rowComponents.EntryType.Text)) {
			logrus.Error("CheckStUnit failed. EntryType: ", rowComponents.EntryType.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryType is invalid", *stapp.Window)
			return false
		}
		// 检查 变量名 的合法性
		if rowComponents.EntryName != nil && (rowComponents.EntryName.Text == "" || strings.Contains(rowComponents.EntryName.Text, " ") || utils.CheckPositiveInteger(rowComponents.EntryName.Text) || utils.CheckStartWithNum(rowComponents.EntryName.Text)) {
			logrus.Error("CheckStUnit failed. EntryName: ", rowComponents.EntryName.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryName is invalid", *stapp.Window)
			return false
		}
	}

	if !logic.CheckFieldNameList(stUnit.RowList) {
		dialog.ShowInformation("Error!", " The field name ara duplicate", *stapp.Window)
		return false
	}

	if !logic.CheckFieldIndexList(stUnit.RowList) {
		dialog.ShowInformation("Error!", " The field index ara duplicate", *stapp.Window)
		return false
	}
	return true
}

// 搜索候选代码
// {
// 	entryKeySelect := xwidget.NewCompletionEntry([]string{})
// 		// 设置默认值
// 		// When the use typed text, complete the list.
// 		entryKeySelect.OnChanged = func(s string) {
// 			// completion start for text length >= 3
// 			if len(s) < 3 {
// 				entryKeySelect.HideCompletion()
// 				return
// 			}

// 			// Make a search on wikipedia
// 			resp, err := http.Get(
// 				"https://en.wikipedia.org/w/api.php?action=opensearch&search=" + entryKeySelect.Text,
// 			)
// 			if err != nil {
// 				entryKeySelect.HideCompletion()
// 				return
// 			}

// 			// Get the list of possible completion
// 			var results [][]string
// 			json.NewDecoder(resp.Body).Decode(&results)

// 			// no results
// 			if len(results) == 0 {
// 				entryKeySelect.HideCompletion()
// 				return
// 			}

// 			// then show them
// 			entryKeySelect.SetOptions(results[1])
// 			entryKeySelect.ShowCompletion()
// 		}
// }
