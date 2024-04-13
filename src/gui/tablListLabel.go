package gui

import (
	"protocolgo/src/logic"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

// table1 列表项
type TableListLabel struct {
	widget.Label
	data      binding.String
	app       *StApp
	tabletype logic.ETableType
}

// 单击事件
func (m *TableListLabel) Tapped(e *fyne.PointEvent) {
	// 处理单击事件
	msg, _ := m.data.Get()
	logrus.Info("Left click! Item: "+msg+",tabletype:", m.tabletype)
}

// 右击事件
func (m *TableListLabel) TappedSecondary(e *fyne.PointEvent) {
	// 处理右键点击或长按事件
	msg, _ := m.data.Get()
	logrus.Info("Right click! Item: "+msg+",m.tabletype:", m.tabletype)
	// 创建一个fyne的popUp
	popUpContent := container.NewVBox()
	popUp := widget.NewPopUp(popUpContent, fyne.CurrentApp().Driver().CanvasForObject(m))

	// 增加 Edit 选项
	popUpContent.Add(widget.NewButton("Edit", func() {
		// 去除字符串中的[],以及其中的字符
		re := regexp.MustCompile(`\[.*?\]`)
		msg = re.ReplaceAllString(msg, "")
		// 矫正类型
		eTableType := m.tabletype
		if eTableType == logic.TableType_Main {
			eTableType = m.app.CoreMgr.SearchTableListWithName(msg)
		}
		logrus.Info("Edit clicked. Item: "+msg+",tabletype:", m.tabletype)
		m.app.EditUnit(eTableType, msg)
		// popUp.Hide() // 隐藏窗口
	}))

	// 增加 Revert/Delete 选项
	if m.tabletype == logic.TableType_Main {
		popUpContent.Add(widget.NewButton("Revert", func() {
			dialog.NewConfirm("Confirmation", "Are you sure to revert?", func(response bool) {
				if response { // if 'Yes' clicked
					msg, _ := m.data.Get()
					// 去除字符串中的[],以及其中的字符
					re := regexp.MustCompile(`\[.*?\]`)
					msg = re.ReplaceAllString(msg, "")
					// 矫正类型
					tabletype := m.app.CoreMgr.SearchTableListWithName(msg)
					logrus.Info("User confirm to revert: "+msg+",tabletype:", tabletype)

					m.app.CoreMgr.RevertUnitFromChanged(tabletype, msg)
				}
				popUp.Hide() // 隐藏窗口
			}, *m.app.Window).Show()
		}))
	} else {
		popUpContent.Add(widget.NewButton("Delete", func() {
			dialog.NewConfirm("Confirmation", "Are you sure to delete?", func(response bool) {
				if response { // if 'Yes' clicked
					logrus.Info("User confirm to delete: "+msg+",tabletype:", m.tabletype)
					msg, _ := m.data.Get()
					// 去除字符串中的[],以及其中的字符
					re := regexp.MustCompile(`\[.*?\]`)
					msg = re.ReplaceAllString(msg, "")
					m.app.CoreMgr.DeleteCurrUnit(m.tabletype, msg)
				}
				popUp.Hide() // 隐藏窗口
			}, *m.app.Window).Show()
		}))
	}

	// 设置窗口的位置
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(m)
	pos = fyne.NewPos(pos.X+e.Position.X, pos.Y+e.Position.Y)
	popUp.Move(pos)

	// 当点击窗口外部，窗口自动消失
	popUp.Show()
}

// 双击事件
func (m *TableListLabel) DoubleTapped(e *fyne.PointEvent) {
	// 处理双击事件
	msg, _ := m.data.Get()
	// 去除字符串中的[],以及其中的字符
	re := regexp.MustCompile(`\[.*?\]`)
	msg = re.ReplaceAllString(msg, "")
	logrus.Info("Double clicked! Item: "+msg+",tabletype:", m.tabletype)
	// 矫正类型
	eTableType := m.tabletype
	if eTableType == logic.TableType_Main {
		eTableType = m.app.CoreMgr.SearchTableListWithName(msg)
	}
	m.app.EditUnit(eTableType, msg)
}
