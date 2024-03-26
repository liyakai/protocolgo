package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type UnitTypeActionItem struct {
	title string
	icon  fyne.Resource
}

func (m *UnitTypeActionItem) ToolbarObject() fyne.CanvasObject {
	return widget.NewButtonWithIcon(m.title, m.icon, func() {
		// 这里是点击动作项时要执行的动作
		fmt.Println("动作项被点击")
	})
}
