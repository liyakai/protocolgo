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
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/beevik/etree"
	"github.com/flopp/go-findfont"
	"github.com/goki/freetype/truetype"
	"github.com/sirupsen/logrus"

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
	stapp.SuitForChinese()
	// 设置图标
	stapp.SetIcon()
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
	// 先从workspace找字体
	fontPath := utils.GetWorkRootPath() + "/data/localziti.ttf"
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		//设置中文字体
		fontPath, err = findfont.Find("SIMFANG.TTF")
		if err != nil {
			panic(err)
		}
		logrus.Info("Found 'SIMFANG.ttf' in ", fontPath)
		// load the font with the freetype library
		// 原作者使用的ioutil.ReadFile已经弃用
		fontData, err = os.ReadFile(fontPath)
		if err != nil {
			panic(err)
		}
	} else {
		logrus.Info("Found local fontPath:", fontPath)
	}

	_, err = truetype.Parse(fontData)
	if err != nil {
		panic(err)
	}
	logrus.Info("SuitForChinese before set env. fontPath:", fontPath)
	os.Setenv("FYNE_FONT", fontPath)
	os.Setenv("FYNE_FONT_MONOSPACE", fontPath)
	logrus.Info("SuitForChinese done. fontPath:", fontPath)
}

// 设置图标
func (stapp *StApp) SetIcon() {
	iconPath := utils.GetWorkRootPath() + "/icon/protobuf.png"
	// 读取图标文件
	bytes, err := os.ReadFile(iconPath)
	if err != nil {
		log.Fatal(err)
	}
	icon := fyne.NewStaticResource("Icon", bytes)
	(*stapp.Window).SetIcon(icon)
	(*stapp.App).SetIcon(icon)
	logrus.Info("SetIcon done. iconPath:", iconPath)
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
					stapp.CoreMgr.CloseCurrProtoXmlFile()
					stapp.CoreMgr.CloseSSH()
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
	newMenuItem := fyne.NewMenuItem("new proto xml..", func() {
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

		// 创建退出快捷键
		(*stapp.Window).Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
			// 在这里检查按下的是不是 Esc 键
			if ke.Name == fyne.KeyEscape {
				// 如果是 Esc 键，隐藏自定义对话框
				saveDialog.Hide()
			}
		})

		saveDialog.Show()

	})
	// 打开菜单项
	openMenuItem := fyne.NewMenuItem("open..", func() {
		// 打开文件
		file_picker := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			// Check for errors
			if err != nil {
				logrus.Info("Failed to NewFileOpen:", err)
				return
			}
			if reader == nil {
				// 用户点击了取消
				logrus.Info("User canceled to open config xml.")
				// dialog.ShowInformation("Cancelled", "File selection cancelled", *stapp.Window)
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

		// 创建退出快捷键
		(*stapp.Window).Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
			// 在这里检查按下的是不是 Esc 键
			if ke.Name == fyne.KeyEscape {
				// 如果是 Esc 键，隐藏自定义对话框
				file_picker.Hide()
			}
		})

		// 显示打开文件的UI
		file_picker.Show()

	})
	// 打开远程配置
	// openRemoteConfig := fyne.NewMenuItem("open ssh..", func() {
	// 	dialogContent := container.NewGridWithRows(4)
	// 	customDialog := dialog.NewCustomWithoutButtons("open remote xml", container.NewVScroll(dialogContent), *stapp.Window)
	// 	customDialog.Resize(fyne.NewSize(600, 400))

	// 	dialogContent.Add(stapp.GetRemoteSshUI(customDialog))
	// 	// 创建退出快捷键
	// 	(*stapp.Window).Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
	// 		// 在这里检查按下的是不是 Esc 键
	// 		if ke.Name == fyne.KeyEscape {
	// 			// 如果是 Esc 键，隐藏自定义对话框
	// 			customDialog.Hide()
	// 		}
	// 	})

	// 	customDialog.Show()

	// })
	// 保存菜单项
	saveMenuItem := fyne.NewMenuItem("save..", func() {
		dialog.ShowConfirm("Confirmation", "Are you sure you want to Save?",
			func(response bool) {
				if response {
					// addition logic to save file goes here
					if stapp.CoreMgr.SaveToProtoXmlFile() {
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
		container.NewTabItem("Main", stapp.CreateTab(logic.TableType_Main)),
		container.NewTabItem("Enum", stapp.CreateTab(logic.TableType_Enum)),
		container.NewTabItem("Data", stapp.CreateTab(logic.TableType_Data)),
		container.NewTabItem("Ptc", stapp.CreateTab(logic.TableType_Protocol)),
		container.NewTabItem("Rpc", stapp.CreateTab(logic.TableType_RPC)),
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
	searchFields := stapp.CoreMgr.GetAllSearchName()
	searchEntry := logic.NewCompletionEntry(searchFields)
	searchEntry.SetPlaceHolder("Search here")
	searchEntry.SetMinRowsVisible(2)
	// 设置默认值
	// When the use typed text, complete the list.
	searchEntry.OnChanged = func(str string) {
		if str != "" {
			matches := fuzzy.RankFind(str, stapp.CoreMgr.GetAllSearchName())
			sort.Sort(matches)
			var strMatches []string
			for _, matchone := range matches {
				strMatches = append(strMatches, matchone.Target)
			}
			searchEntry.SetOptions(strMatches)
			// 设置焦点
			strEntryName := stapp.CoreMgr.GetListNameBySearchName(str)
			if strEntryName != "" {
				eTableType := stapp.CoreMgr.SearchTableListWithName(strEntryName)
				if eTableType == logic.TableType_Enum {
					stapp.tables.SelectIndex(1)
				} else if eTableType == logic.TableType_Data {
					stapp.tables.SelectIndex(2)
				} else if eTableType == logic.TableType_Protocol {
					stapp.tables.SelectIndex(3)
				} else if eTableType == logic.TableType_RPC {
					stapp.tables.SelectIndex(4)
				}
			}
			searchEntry.ShowCompletion()

		} else {
			stapp.CoreMgr.SyncListWithETree()
		}

	}
	searchEntry.OnSubmitted = func(str string) {
		// 找出搜索内容对应的key
		strListName := stapp.CoreMgr.GetListNameBySearchName(str)
		if strListName == "" {
			logrus.Info("[searchEntry] Cannot find the list name by str:", str)
		}
		// 设置焦点
		eTableType := stapp.CoreMgr.SearchTableListWithName(strListName)
		if eTableType == logic.TableType_Enum {
			stapp.tables.SelectIndex(0)
			stapp.CoreMgr.EnumTableList.Set([]string{strListName})
		} else if eTableType == logic.TableType_Data {
			stapp.tables.SelectIndex(1)
			stapp.CoreMgr.DataTableList.Set([]string{strListName})
		} else if eTableType == logic.TableType_Protocol {
			stapp.tables.SelectIndex(2)
			stapp.CoreMgr.PtcTableList.Set([]string{strListName})
		} else if eTableType == logic.TableType_RPC {
			stapp.tables.SelectIndex(3)
			stapp.CoreMgr.RpcTableList.Set([]string{strListName})
		}
	}

	// 添加快捷键,Ctrl+f 聚焦搜索栏
	(*stapp.Window).Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyF,
		Modifier: fyne.KeyModifierControl,
	}, func(shortcut fyne.Shortcut) {
		(*stapp.Window).Canvas().Focus(searchEntry)
	})

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
		stapp.CreateButtomCanvas(tabletype),
		nil,
		nil,
		list,
	)

	return buttonwithlist
}

// 创建list的说明和button
func (stapp *StApp) CreateTabListInstruction(tabletype logic.ETableType) fyne.CanvasObject {
	label := widget.NewLabel(stapp.CoreMgr.GetTopLableStingByType(tabletype))
	var button *widget.Button
	if tabletype == logic.TableType_Main {
		button = widget.NewButton("Save to File", func() {
			stapp.CoreMgr.SaveProtoXmlToFile()
		})
	} else {
		button = widget.NewButton("Add new", func() {
			stapp.EditUnit(tabletype, "")
		})
	}
	// 使用HBox将searchEntry和searchButton安排在同一行，并使用HSplit来设置比例
	topContainer := container.NewHSplit(container.NewStack(label), button)
	topContainer.Offset = 0.75 //设置searchEntry 占 3/4， searchButton 占 1/4
	return topContainer

}

// 创建新Message的编辑页面
func (stapp *StApp) EditUnit(tabletype logic.ETableType, unitname string) {
	dialogContent := container.NewGridWithRows(4)
	customDialog := dialog.NewCustomWithoutButtons(stapp.CoreMgr.GetEditTableTitle(tabletype, unitname), container.NewVScroll(dialogContent), *stapp.Window)
	customDialog.Resize(fyne.NewSize(1100, 800))

	subtabletype := logic.SubTableType_None
	if tabletype == logic.TableType_RPC {
		subtabletype = logic.SubTableType_RpcReq
	}
	inputInfoContainer, unitReq := stapp.GetUnitDetailContainer(customDialog, tabletype, subtabletype, unitname)
	dialogContent.Add(inputInfoContainer)

	var inputInfoContainer2 *fyne.Container
	var unitAck *logic.StUnitContainer
	if tabletype == logic.TableType_RPC {
		subtabletype = logic.SubTableType_RpcAck
		inputInfoContainer2, unitAck = stapp.GetUnitDetailContainer(customDialog, tabletype, subtabletype, unitname)
		logrus.Debug("[EditUnit] debug unit name. tabletype:", tabletype, ",unitname:", unitname, ",unitname:", unitname)
	}

	// 获取取消/保存按钮组,以及依赖列表
	refCancelSaveButtons, referencesList := stapp.GetUnitDetailButtons(unitReq, unitAck, customDialog)
	// dialogContent.Add(referencesList)
	if tabletype == logic.TableType_RPC {
		inputInfoContainer2.Add(container.NewCenter(refCancelSaveButtons))
		dialogContent.Add(inputInfoContainer2)
		// dialogContent.Add(referencesList2)
	} else {
		inputInfoContainer.Add(container.NewCenter(refCancelSaveButtons))
	}

	dialogContent.Add(referencesList)

	// 创建退出快捷键
	(*stapp.Window).Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		// 在这里检查按下的是不是 Esc 键
		if ke.Name == fyne.KeyEscape {
			// 如果是 Esc 键，隐藏自定义对话框
			customDialog.Hide()
		}
	})

	customDialog.Show()
}

func (stapp *StApp) GetUnitDetailContainer(customDialog *dialog.CustomDialog, tabletype logic.ETableType, subtabletype logic.ESubTableType, unitname string) (*fyne.Container, *logic.StUnitContainer) {
	bCreateNew := false // 是否是新的节点
	etreeRow := stapp.CoreMgr.GetEtreeElem(tabletype, subtabletype, unitname)
	if unitname == "" || etreeRow == nil {
		bCreateNew = true
		logrus.Info("[EditUnit] Create new unit. tabletype:", tabletype, ",customDialog:", customDialog)
	} else {
		logrus.Info("[EditUnit] Edit old unit. tabletype:", tabletype, ", Tag:", etreeRow.Tag, ",customDialog:", customDialog)
	}

	// 创建输入信息的容器
	inputInfoContainer := container.NewVBox()
	isSucess, firstFullName, secondFullName := stapp.CoreMgr.DetectFullNameByProtoName(unitname)

	// 创建源数据下拉框
	selectSourceServer := widget.NewSelect(stapp.CoreMgr.GetConfigFullServerName(), nil)
	selectTargetServer := widget.NewSelect(stapp.CoreMgr.GetConfigFullServerName(), nil)
	if isSucess {
		selectSourceServer.Selected = firstFullName
		selectTargetServer.Selected = secondFullName
	}

	selectServer := container.NewHSplit(selectSourceServer, selectTargetServer)
	// 创建输入框
	inputUnitName := widget.NewEntry()
	inputUnitName.SetPlaceHolder("Enter name...")
	if !bCreateNew {
		inputUnitName.SetText(unitname)
		inputUnitName.TextStyle.Bold = true
		// inputUnitName.Disable()
	}

	// 如果下拉框修改了,则更新inputUnitName名字.[互相联动]
	selectSourceServer.OnChanged = func(strSourceName string) {
		inputUnitName.SetText(stapp.CoreMgr.GetProtoNameFromSourceTargetServer(selectSourceServer.Selected, selectTargetServer.Selected, inputUnitName.Text))
		inputUnitName.Refresh()
	}
	selectTargetServer.OnChanged = func(strTargetName string) {
		inputUnitName.SetText(stapp.CoreMgr.GetProtoNameFromSourceTargetServer(selectSourceServer.Selected, selectTargetServer.Selected, inputUnitName.Text))
		inputUnitName.Refresh()
	}
	// 如果输入框改了,则自动修改下拉框.[互相联动]
	inputUnitName.OnChanged = func(strProtoName string) {
		if tabletype != logic.TableType_Protocol && tabletype != logic.TableType_RPC {
			return
		}
		isSucess, firstFullName, secondFullName := stapp.CoreMgr.DetectFullNameByProtoName(strProtoName)
		if isSucess {
			logrus.Info("[EditUnit] inputUnitName OnChanged. strProtoName:", strProtoName, " firstFullName:", firstFullName, ",secondFullName:", secondFullName)
			selectSourceServer.Selected = firstFullName
			selectTargetServer.Selected = secondFullName
			selectSourceServer.Refresh()
			selectTargetServer.Refresh()
		}
	}

	selectSeerverInputName := container.NewHSplit(selectServer, inputUnitName)
	selectSeerverInputName.Offset = 0.2
	if tabletype == logic.TableType_Enum {
		inputInfoContainer.Add(inputUnitName)
	} else if tabletype == logic.TableType_Data {
		inputInfoContainer.Add(inputUnitName)
	} else if tabletype == logic.TableType_Protocol {
		inputInfoContainer.Add(selectSeerverInputName)
	} else if tabletype == logic.TableType_RPC {
		if subtabletype == logic.SubTableType_RpcReq {
			inputInfoContainer.Add(selectSeerverInputName)
		} else if subtabletype == logic.SubTableType_RpcAck {
			inputInfoContainer.Add(widget.NewLabel(unitname + " Ack:"))
		} else {
			logrus.Error("[EduitUnit] Invalid subtabletype:", subtabletype)
		}

	}

	// 注释
	inputUnitComment := widget.NewMultiLineEntry()
	inputUnitComment.SetPlaceHolder("Enter Comment...")
	inputUnitComment.SetMinRowsVisible(3)
	inputInfoContainer.Add(inputUnitComment)
	if !bCreateNew {
		for _, child := range etreeRow.Child {
			// 检查该子元素是否为注释
			if comment, ok := child.(*etree.Comment); ok {
				inputUnitComment.SetText(comment.Data)
				break
			}
		}
	}

	nEntryIndex := 0
	// 在外部定义一个列表来保存每一行的组件
	rowList := new([]logic.StRowUnit)

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
			entryDefault := child.SelectAttr("EntryDefault")
			if entryDefault != nil {
				rowUnit.EntryDefault = entryDefault.Value
			}
			entryContent := child.SelectAttr("EntryComment")
			if entryContent != nil {
				rowUnit.EntryComment = entryContent.Value
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
			stapp.CreateRowForEditUnit(tabletype, rowUnit, attrBox, rowList)
		}
	}

	// 增加新的字段
	addButton := widget.NewButton("Add field", func() {
		if tabletype != logic.TableType_Enum {
			nEntryIndex = nEntryIndex + 1
		}
		stapp.CreateRowForEditUnit(tabletype, logic.StStrRowUnit{
			EntryIndex: strconv.Itoa(nEntryIndex),
		}, attrBox, rowList)

		if tabletype == logic.TableType_Enum {
			nEntryIndex = nEntryIndex + 1
		}
	})

	// 创建可以新增列的container
	attrBoader := container.NewBorder(nil, nil, nil, addButton, attrBox)
	inputInfoContainer.Add(attrBoader)

	var stUnitContainer logic.StUnitContainer
	stUnitContainer.UnitNameEntry = inputUnitName
	stUnitContainer.UnitCommentEntry = inputUnitComment
	stUnitContainer.TableType = tabletype
	stUnitContainer.RowList = rowList
	stUnitContainer.IsCreatNew = bCreateNew

	// inputInfoContainer.Add(container.NewCenter(buttons))
	return inputInfoContainer, &stUnitContainer
}

func (stapp *StApp) GetRemoteSshUI(customDialog *dialog.CustomDialog) *fyne.Container {
	sshInfoContainer := container.NewVBox()

	// 从 config 获取默认配置
	bRet, strIp, strPort, strUsername, strPassword := stapp.CoreMgr.GetSSHConfig()
	// ip 行
	labelIpEntry := widget.NewLabel("ip:")
	labelIpEntry.Alignment = fyne.TextAlignTrailing
	inputIpEntry := widget.NewEntry()
	if bRet {
		inputIpEntry.Text = strIp
	} else {
		inputIpEntry.SetPlaceHolder("Enter ip address...")
	}
	splitInputIP := container.NewHSplit(labelIpEntry, inputIpEntry)
	splitInputIP.Offset = 0.15
	sshInfoContainer.Add(splitInputIP)

	// port 行
	labelPortEntry := widget.NewLabel("port:")
	labelPortEntry.Alignment = fyne.TextAlignTrailing
	inputPortEntry := widget.NewEntry()
	if bRet {
		inputPortEntry.Text = strPort
	} else {
		inputPortEntry.SetPlaceHolder("Enter port...")
	}
	splitInputPort := container.NewHSplit(labelPortEntry, inputPortEntry)
	splitInputPort.Offset = 0.15
	sshInfoContainer.Add(splitInputPort)

	// use 行
	labelUserEntry := widget.NewLabel("user:")
	labelUserEntry.Alignment = fyne.TextAlignTrailing
	inputUserEntry := widget.NewEntry()
	if bRet {
		inputUserEntry.Text = strUsername
	} else {
		inputUserEntry.SetPlaceHolder("Enter user name...")
	}
	splitInputUser := container.NewHSplit(labelUserEntry, inputUserEntry)
	splitInputUser.Offset = 0.15
	sshInfoContainer.Add(splitInputUser)

	// password 行
	labelPasswordEntry := widget.NewLabel("passwd:")
	labelPasswordEntry.Alignment = fyne.TextAlignTrailing
	inputPasswordEntry := widget.NewEntry()
	if bRet {
		inputPasswordEntry.Text = strPassword
	} else {
		inputPasswordEntry.SetPlaceHolder("Enter password...")
	}
	inputPasswordEntry.Password = true
	splitInputPassword := container.NewHSplit(labelPasswordEntry, inputPasswordEntry)
	splitInputPassword.Offset = 0.15
	sshInfoContainer.Add(splitInputPassword)

	// 增加新的字段
	connectButton := widget.NewButton("Connect", func() {
		isSucc, strError := stapp.CoreMgr.OpenSSH(inputIpEntry.Text, inputPortEntry.Text, inputUserEntry.Text, inputPasswordEntry.Text)
		if !isSucc {
			dialog.ShowInformation("Error!", "OpenSSH failed for "+strError, *stapp.Window)
			logrus.Error("[GetRemoteSshUI] OpenSSH failed for " + strError)
		}
	})
	// 设置按钮样式
	// connectButton.Importance = widget.HighImportance
	cancelButton := widget.NewButton("Cancel", func() {
		customDialog.Hide()
	})

	buttons := container.NewHBox(cancelButton, connectButton)
	sshInfoContainer.Add(container.NewCenter(buttons))

	return sshInfoContainer
}

func (stapp *StApp) GetUnitDetailButtons(stUnitReq *logic.StUnitContainer, stUnitAck *logic.StUnitContainer, customDialog *dialog.CustomDialog) (*fyne.Container, *fyne.Container) {
	var referencesList *fyne.Container
	referencesList = stapp.CreateEntryReferenceListCanvas(stUnitReq.UnitNameEntry.Text)
	referencesList.Hide()
	bShowReferenceList := false

	// 增加关闭,保存按钮
	buttons := container.NewHBox(
		widget.NewButton("References", func() {
			referencesList = stapp.CreateEntryReferenceListCanvas(stUnitReq.UnitNameEntry.Text)
			referencesList.Hide()
			// Cancel logic goes here
			if bShowReferenceList {
				bShowReferenceList = false
				referencesList.Hide()
			} else {
				bShowReferenceList = true
				referencesList.Show()
			}

		}),
		// Cancel Button
		widget.NewButton("Cancel", func() {
			// Cancel logic goes here
			customDialog.Hide()
		}),
		// Save Button
		widget.NewButton("Save", func() {
			// Save logic goes here
			logrus.Info("[CreateNewMessage]Save. UnitName: " + stUnitReq.GetStUnit().UnitName)

			var stUnits logic.StUnits
			stUnits.UnitListName = stUnitReq.GetStUnit().UnitName
			stReqUnit := stUnitReq.GetStUnit()
			stReqUnit.SubTableType = logic.SubTableType_RpcReq
			stUnits.UnitList = append(stUnits.UnitList, stReqUnit)

			if stUnitAck != nil {
				stAckUnit := stUnitAck.GetStUnit()
				// rpc的回包tag名字与请求tag名字相同
				stAckUnit.UnitName = stUnitReq.GetStUnit().UnitName
				stAckUnit.SubTableType = logic.SubTableType_RpcAck
				stUnits.UnitList = append(stUnits.UnitList, stAckUnit)
			}

			if !stapp.CheckStUnits(stUnits) {
				logrus.Error("[CreateNewMessage] CheckEditMessage failed.")
				return
			}

			if !stapp.CoreMgr.AddUpdateUnits(stUnits) {
				logrus.Error("[CreateNewMessage] AddUpdateUnits failed.")
				return
			}

			customDialog.Hide()
		}),
	)
	return buttons, referencesList
}

func (stapp *StApp) CreateRowForEditUnit(tabletype logic.ETableType, strRowUnit logic.StStrRowUnit, attrBox *fyne.Container, rowList *[]logic.StRowUnit) {
	var entryOption *widget.Select
	// var entryType *widget.Entry
	var entryTypeSelect *logic.CompletionEntry
	var entryDefault *widget.Select
	containerDefaultComment := container.NewGridWithRows(1)
	entryName := widget.NewEntry()
	entryComment := widget.NewEntry()
	entryIndex := widget.NewEntry()

	if tabletype == logic.TableType_Protocol || tabletype == logic.TableType_Data || tabletype == logic.TableType_RPC {
		entryOption = widget.NewSelect([]string{"optional", "repeated"}, nil)
		if strRowUnit.EntryOption == "" {
			entryOption.Selected = "optional"
		} else {
			entryOption.Selected = strRowUnit.EntryOption
		}

		// 枚举默认值
		entryDefault = widget.NewSelect([]string{}, nil)
		// entryDefault.SetPlaceHolder("Default Value...")

		// 输入框
		// entryType = widget.NewEntry()
		// entryType.SetPlaceHolder("Enter type...")
		// if strRowUnit.EntryType != "" {
		// 	entryType.SetText(strRowUnit.EntryType)
		// }

		// 搜索框
		searchFields := stapp.CoreMgr.GetAllUseableEntryTypeWithProtoType()
		entryTypeSelect = logic.NewCompletionEntry(searchFields)
		entryTypeSelect.ShowMouseMenu(true) // 不依靠鼠标事件了,这个没用了.
		entryTypeSelect.SetPlaceHolder("Enter filed type...")
		if strRowUnit.EntryType != "" {
			entryTypeSelect.SetText(strRowUnit.EntryType)
		}
		// When the use typed text, complete the list.
		entryTypeSelect.OnChanged = func(str string) {
			if str == "" {
				return
			}
			matches := fuzzy.RankFind(str, searchFields)
			sort.Sort(matches)
			var strMatches []string
			for _, matchone := range matches {
				strMatches = append(strMatches, matchone.Target)
			}
			entryTypeSelect.SetOptions(strMatches)
			entryTypeSelect.ShowCompletion()

			// 检查新值是否是枚举类型
			containerDefaultComment.Objects = nil
			entryDefault.SetOptions([]string{})
			if stapp.CoreMgr.SearchTableListWithName(str) == logic.TableType_Enum {
				entryDefault.SetOptions(stapp.CoreMgr.GetVarListOfEnum(entryTypeSelect.Text))
				containerDefaultComment.Add(entryDefault)
			}
			containerDefaultComment.Add(entryComment)
			containerDefaultComment.Refresh()
		}

		entryDefault.OnChanged = func(str string) {
			entryDefault.SetOptions(stapp.CoreMgr.GetVarListOfEnum(entryTypeSelect.Text))
		}
		// 鼠标右键事件
		entryTypeSelect.OnTappedSecondary = func(pe *fyne.PointEvent) {
			if entryTypeSelect.Text == "" {
				return
			}
			stapp.CreateEntryTypeInfo(entryTypeSelect, pe)
		}

	}

	entryName.SetPlaceHolder("Enter variable name...")
	if strRowUnit.EntryName != "" {
		entryName.SetText(strRowUnit.EntryName)
	}

	entryComment.SetPlaceHolder("Comment...")
	if strRowUnit.EntryComment != "" {
		entryComment.SetText(strRowUnit.EntryComment)
	}

	entryIndex.SetText(strRowUnit.EntryIndex)

	var oneRow *container.Split
	if tabletype == logic.TableType_Enum {
		oneRowNameIndex := container.NewHSplit(entryName, entryIndex)
		oneRowNameIndex.Offset = 0.9
		oneRow = container.NewHSplit(oneRowNameIndex, entryComment)
		oneRow.Offset = 0.6
	} else if tabletype == logic.TableType_Protocol || tabletype == logic.TableType_Data || tabletype == logic.TableType_RPC {
		oneRowKeyValue := container.NewHSplit(entryTypeSelect, entryName)
		oneRowKeyValue.Offset = 0.35

		oneRowOptionKeyValue := container.NewHSplit(entryOption, oneRowKeyValue)
		oneRowOptionKeyValue.Offset = 0.15

		oneRowOptionKeyValueIndex := container.NewHSplit(oneRowOptionKeyValue, entryIndex)
		oneRowOptionKeyValueIndex.Offset = 0.95

		if stapp.CoreMgr.SearchTableListWithName(entryTypeSelect.Text) == logic.TableType_Enum {
			entryDefault.SetOptions(stapp.CoreMgr.GetVarListOfEnum(entryTypeSelect.Text))
			if strRowUnit.EntryDefault != "" {
				entryDefault.SetSelected(strRowUnit.EntryDefault)
			}
			containerDefaultComment.Add(entryDefault)
		}
		containerDefaultComment.Add(entryComment)
		oneRow = container.NewHSplit(oneRowOptionKeyValueIndex, containerDefaultComment)
		oneRow.Offset = 0.7
	}

	// 创建一个新的RowComponents实例并保存到列表中,加入列表,方便获取数值
	stRow := logic.StRowUnit{
		EntryIndex:   entryIndex,
		EntryOption:  entryOption,
		EntryType:    entryTypeSelect,
		EntryName:    entryName,
		EntryDefault: entryDefault,
		EntryComment: entryComment,
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

// 创建详情页
func (stapp *StApp) CreateEntryTypeInfo(entry *logic.CompletionEntry, pe *fyne.PointEvent) {

	// 如果是proto type 则不创建
	if stapp.CoreMgr.CheckProtoType(entry.Text) {
		return
	}
	// 右键按钮
	// clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	canvas := fyne.CurrentApp().Driver().CanvasForObject((*stapp.Window).Content())

	// cutItem := fyne.NewMenuItem("Cut", func() {
	// 	canvas.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutCut{Clipboard: clipboard})
	// })
	// copyItem := fyne.NewMenuItem("Copy", func() {
	// 	canvas.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutCopy{Clipboard: clipboard})
	// })
	// pasteItem := fyne.NewMenuItem("Paste", func() {
	// 	canvas.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
	// })
	detailInfoItem := fyne.NewMenuItem("Edit", func() {
		// canvas.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
		stapp.EditUnit(stapp.CoreMgr.SearchTableListWithName(entry.Text), entry.Text)

	})
	referencesItem := fyne.NewMenuItem("References", func() {
		// canvas.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
		stapp.CreateEntryReferenceListSingle(entry, pe)
	})

	popUpPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(entry)
	popUpPos = fyne.NewPos(popUpPos.X+pe.Position.X, popUpPos.Y+pe.Position.Y)

	menu := fyne.NewMenu("", detailInfoItem, referencesItem)

	pPopUp := widget.NewPopUpMenu(menu, canvas)
	pPopUp.ShowAtPosition(popUpPos)
	pPopUp.Show()
	//
	logrus.Info("[CreateRowForEditUnit] CreateEntryTypeInfo. ")

}

func (stapp *StApp) CreateEntryReferenceListSingle(entry *logic.CompletionEntry, pe *fyne.PointEvent) {
	canvas := fyne.CurrentApp().Driver().CanvasForObject((*stapp.Window).Content())
	pPopUp := widget.NewPopUp(stapp.CreateEntryReferenceListCanvas(entry.Text), canvas)
	// 设置窗口大小
	pPopUp.Resize(fyne.NewSize(600, 600))
	// 设置窗口位置
	popUpPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(entry)
	popUpPos = fyne.NewPos(popUpPos.X+pe.Position.X, popUpPos.Y+pe.Position.Y)
	pPopUp.ShowAtPosition(popUpPos)
	// 显示
	pPopUp.Show()
}

func (stapp *StApp) CreateEntryReferenceListCanvas(name string) *fyne.Container {
	bindingStringList := binding.NewStringList()
	logrus.Info("[CreateEntryReferenceList] name:", name, ", GetReferences:", stapp.CoreMgr.GetReferences(name))
	bindingStringList.Set(stapp.CoreMgr.GetReferences(name))
	referenceList := widget.NewListWithData(bindingStringList,
		func() fyne.CanvasObject {
			label := &TableListLabel{}
			label.ExtendBaseWidget(label)
			return label
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*TableListLabel).Bind(i.(binding.String))
			o.(*TableListLabel).data = i.(binding.String)
			o.(*TableListLabel).app = stapp
			o.(*TableListLabel).tabletype = logic.TableType_Protocol
		},
	)
	// 使用垂直布局将上部和下部容器组合在一起
	buttonwithlist := container.NewBorder(
		widget.NewLabel(name+"'s references list:"),
		nil,
		nil,
		nil,
		referenceList,
	)
	return buttonwithlist
}

func (stapp *StApp) DestoryEntryTypeInfo(pPopUp **widget.PopUp) {
	if *pPopUp != nil {
		(*pPopUp).Hide()
		*pPopUp = nil
	}
	logrus.Info("[CreateRowForEditUnit] OnMouseOut. ")
}

// 检查 StUnits
func (stapp *StApp) CheckStUnits(stUnits logic.StUnits) bool {
	// 检查 name 的合法性
	if stUnits.UnitListName == "" || strings.Contains(stUnits.UnitListName, " ") || utils.CheckPositiveInteger(stUnits.UnitListName) || utils.CheckStartWithNum(stUnits.UnitListName) {
		logrus.Error("CheckStUnits failed. invalid MsgName: ", stUnits.UnitListName)
		dialog.ShowInformation("Error!", "The name is invalid", *stapp.Window)
		return false
	}

	for _, stUnit := range stUnits.UnitList {
		if !stapp.CheckStUnit(stUnit) {
			logrus.Error("CheckStUnits failed. invalid stUnit, name : ", stUnit.UnitName, ", subtabletype:", stUnit.SubTableType)
			dialog.ShowInformation("Error!", "The name is invalid", *stapp.Window)
			return false
		}
	}
	return true
}

// 检查 StUnit
func (stapp *StApp) CheckStUnit(stUnit logic.StUnit) bool {
	// 检查 name 的合法性
	if stUnit.UnitName == "" || strings.Contains(stUnit.UnitName, " ") || utils.CheckPositiveInteger(stUnit.UnitName) || utils.CheckStartWithNum(stUnit.UnitName) {
		logrus.Error("CheckStUnit failed. invalid MsgName: ", stUnit.UnitName)
		dialog.ShowInformation("Error!", "The name is invalid", *stapp.Window)
		return false
	}

	// 检查 name 是否已经存在
	if stUnit.IsCreatNew && stapp.CoreMgr.CheckExistSameName(stUnit.UnitName) {
		logrus.Error("CheckStUnit failed. Repeated MsgName: ", stUnit.UnitName)
		dialog.ShowInformation("Error!", "The name["+stUnit.UnitName+"] is already exist.", *stapp.Window)
		return false
	}

	for _, rowComponents := range stUnit.RowList {
		// 检查 EntryIndex 的合法性
		if rowComponents.EntryIndex != nil && (rowComponents.EntryIndex.Text == "") {
			logrus.Error("CheckStUnit failed. EntryIndex: ", rowComponents.EntryIndex.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryIndex is invalid", *stapp.Window)
			return false
		}

		// 检查索引值合法性
		if stUnit.TableType == logic.TableType_Enum {
			if !utils.CheckNaturalInteger(rowComponents.EntryIndex.Text) {
				logrus.Error("CheckStUnit failed. EntryName: ", rowComponents.EntryName.Text)
				dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryIndex is invalid", *stapp.Window)
				return false
			}
		} else if stUnit.TableType == logic.TableType_Protocol {
			if !utils.CheckPositiveInteger(rowComponents.EntryIndex.Text) {
				logrus.Error("CheckStUnit failed. EntryName: ", rowComponents.EntryName.Text)
				dialog.ShowInformation("Error!", "EntryName["+rowComponents.EntryName.Text+"], the EntryIndex is invalid", *stapp.Window)
				return false
			}
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
		// 检查默认值合法性
		if rowComponents.EntryDefault != nil && stapp.CoreMgr.SearchTableListWithName(rowComponents.EntryType.Text) == logic.TableType_Enum && (rowComponents.EntryDefault.Selected == "" || strings.Contains(rowComponents.EntryDefault.Selected, " ") || utils.CheckPositiveInteger(rowComponents.EntryDefault.Selected) || utils.CheckStartWithNum(rowComponents.EntryDefault.Selected)) {
			logrus.Error("CheckStUnit failed. EntryIndex: ", rowComponents.EntryIndex.Text, ",EntryName: ", rowComponents.EntryName.Text)
			dialog.ShowInformation("Error!", "Index["+rowComponents.EntryIndex.Text+"], the EntryDefault is invalid", *stapp.Window)
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

// 创建list的说明和button
func (stapp *StApp) CreateButtomCanvas(tabletype logic.ETableType) fyne.CanvasObject {
	label := widget.NewLabel(stapp.CoreMgr.GetButtomLableStingByType(tabletype))
	var buttonGenProto *widget.Button
	if tabletype == logic.TableType_Main {
		buttonGenProto = widget.NewButton("Generate proto file", func() {
			// stapp.CoreMgr.SaveProtoXmlToFile()
			isSuccess, strProtoPath := stapp.CoreMgr.GetGenProtoPath()
			if !isSuccess {
				logrus.Error("Generate proto file failed for GetGenProtoPath.")
				dialog.ShowInformation("Error!", "Generate proto file failed for GetGenProtoPath.", *stapp.Window)
				return
			}
			logic.GenProto(stapp.CoreMgr.FileEtree, strProtoPath)
		})
		buttonGenProtoToPb := widget.NewButton("Generate pb", func() {
			// stapp.CoreMgr.SaveProtoXmlToFile()
			isSuccess, strProtoPath := stapp.CoreMgr.GetGenProtoPath()
			if !isSuccess {
				logrus.Error("Generate pb file failed for GetGenProtoPath.")
				dialog.ShowInformation("Error!", "Generate pb file failed for GetGenProtoPath.", *stapp.Window)
				return
			}
			isSuccess, strPbPath := stapp.CoreMgr.GetGenPbPath()
			if !isSuccess {
				logrus.Error("Generate pb file failed for GetGenPbPath.")
				dialog.ShowInformation("Error!", "Generate pb file failed for GetGenPbPath.", *stapp.Window)
				return
			}
			isSuccess, strError := logic.GenPbFromProto(strProtoPath, strPbPath)
			if !isSuccess {
				logrus.Error("Generate pb file failed for ", strError)
				dialog.ShowInformation("Error!", "Generate pb file failed for "+strError, *stapp.Window)
				return
			}
		})
		// 使用HBox将searchEntry和searchButton安排在同一行，并使用HSplit来设置比例
		buttomContainer := container.NewHBox(container.NewStack(label), buttonGenProto, buttonGenProtoToPb)
		// buttomContainer.Offset = 0.75 //设置searchEntry 占 3/4， searchButton 占 1/4
		return buttomContainer
	}
	return nil

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
