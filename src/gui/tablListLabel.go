package gui

import (
	"protocolgo/src/logic"

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
	logrus.Info("Right click! Item: " + msg)
	// 创建一个fyne的popUp
	popUpContent := container.NewVBox()
	popUp := widget.NewPopUp(popUpContent, fyne.CurrentApp().Driver().CanvasForObject(m))

	// 增加 Edit 选项
	popUpContent.Add(widget.NewButton("Edit", func() {
		logrus.Info("Edit clicked. Item: "+msg+",tabletype:", m.tabletype)
		m.app.EditUnit(m.tabletype, msg)
		// popUp.Hide() // 隐藏窗口
	}))

	// 增加 Delete选项
	popUpContent.Add(widget.NewButton("Delete", func() {
		dialog.NewConfirm("Confirmation", "Are you sure to delete?", func(response bool) {
			if response { // if 'Yes' clicked
				logrus.Info("User confirm to delete: " + msg)
				msg, _ := m.data.Get()
				m.app.CoreMgr.DeleteCurrUnit(m.tabletype, msg)
			}
			popUp.Hide() // 隐藏窗口
		}, *m.app.Window).Show()
	}))

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
	logrus.Info("Double clicked! Item: "+msg+",tabletype:", m.tabletype)
	m.app.EditUnit(m.tabletype, msg)
}
