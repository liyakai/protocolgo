package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// table1 列表项
type TableListLabel struct {
	widget.Label
}

func (m *TableListLabel) Tapped(e *fyne.PointEvent) {
	fmt.Println("Left click!")
}

func (m *TableListLabel) TappedSecondary(e *fyne.PointEvent) {
	fmt.Println("Right click!")
}
