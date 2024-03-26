package logic

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

// CompletionEntry is an Entry with options displayed in a PopUpMenu.
type CompletionEntry struct {
	widget.Entry
	popupMenu     *widget.PopUp
	navigableList *navigableList
	Options       []string
	pause         bool
	itemHeight    float32

	CustomCreate func() fyne.CanvasObject
	CustomUpdate func(id widget.ListItemID, object fyne.CanvasObject)

	// 新增鼠标事件
	bShowMouseMenu    bool
	OnMouseIn         func(*desktop.MouseEvent) `json:"-"`
	OnMouseOut        func()                    `json:"-"`
	OnTappedSecondary func(pe *fyne.PointEvent) `json:"-"`
}

// NewCompletionEntry creates a new CompletionEntry which creates a popup menu that responds to keystrokes to navigate through the items without losing the editing ability of the text input.
func NewCompletionEntry(options []string) *CompletionEntry {
	c := &CompletionEntry{Options: options}
	c.ExtendBaseWidget(c)
	c.bShowMouseMenu = false
	return c
}

// 设置是否显示鼠标滑过菜单
func (c *CompletionEntry) ShowMouseMenu(bMouseMenu bool) {
	c.bShowMouseMenu = bMouseMenu
}

// HideCompletion hides the completion menu.
func (c *CompletionEntry) HideCompletion() {
	if c.popupMenu != nil {
		c.popupMenu.Hide()
	}
}

// Move changes the relative position of the select entry.
//
// Implements: fyne.Widget
func (c *CompletionEntry) Move(pos fyne.Position) {
	c.Entry.Move(pos)
	if c.popupMenu != nil {
		c.popupMenu.Resize(c.maxSize())
		c.popupMenu.Move(c.popUpPos())
	}
}

// Refresh the list to update the options to display.
func (c *CompletionEntry) Refresh() {
	c.Entry.Refresh()
	if c.navigableList != nil {
		c.navigableList.SetOptions(c.Options)
	}
}

// SetOptions set the completion list with itemList and update the view.
func (c *CompletionEntry) SetOptions(itemList []string) {
	c.Options = itemList
	c.Refresh()
}

// ShowCompletion displays the completion menu
func (c *CompletionEntry) ShowCompletion() {
	if c.pause {
		return
	}
	if len(c.Options) == 0 {
		c.HideCompletion()
		return
	}

	if c.navigableList == nil {
		c.navigableList = newNavigableList(c.Options, &c.Entry, c.setTextFromMenu, c.HideCompletion,
			c.CustomCreate, c.CustomUpdate)
	} else {
		c.navigableList.UnselectAll()
		c.navigableList.selected = -1
	}
	holder := fyne.CurrentApp().Driver().CanvasForObject(c)

	if c.popupMenu == nil {
		c.popupMenu = widget.NewPopUp(c.navigableList, holder)
	}
	c.popupMenu.Resize(c.maxSize())
	c.popupMenu.ShowAtPosition(c.popUpPos())
	holder.Focus(c.navigableList)
}

// calculate the max size to make the popup to cover everything below the entry
func (c *CompletionEntry) maxSize() fyne.Size {
	cnv := fyne.CurrentApp().Driver().CanvasForObject(c)

	if c.itemHeight == 0 {
		// set item height to cache
		c.itemHeight = c.navigableList.CreateItem().MinSize().Height
	}

	listheight := float32(len(c.Options))*(c.itemHeight+2*theme.Padding()+theme.SeparatorThicknessSize()) + 2*theme.Padding()
	canvasSize := cnv.Size()
	entrySize := c.Size()
	if canvasSize.Height > listheight {
		return fyne.NewSize(entrySize.Width, listheight)
	}

	return fyne.NewSize(
		entrySize.Width,
		canvasSize.Height-c.Position().Y-entrySize.Height-theme.InputBorderSize()-theme.Padding())
}

// calculate where the popup should appear
func (c *CompletionEntry) popUpPos() fyne.Position {
	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(c)
	return entryPos.Add(fyne.NewPos(0, c.Size().Height))
}

// Prevent the menu to open when the user validate value from the menu.
func (c *CompletionEntry) setTextFromMenu(s string) {
	c.pause = true
	c.Entry.SetText(s)
	c.Entry.CursorColumn = len([]rune(s))
	c.Entry.Refresh()
	c.pause = false
	c.popupMenu.Hide()
}

// 确保 CustomEntry 实现了 Focusable 接口的所有方法
func (e *CompletionEntry) FocusGained() {
	e.Entry.FocusGained()
}

func (e *CompletionEntry) FocusLost() {
	e.Entry.FocusLost()
}

func (e *CompletionEntry) TypedRune(r rune) {
	e.Entry.TypedRune(r)
}

func (e *CompletionEntry) TypedKey(key *fyne.KeyEvent) {
	// 确保调用原始的 TypedKey 方法来继承原有的处理逻辑
	e.Entry.TypedKey(key)
}

func (e *CompletionEntry) MouseIn(event *desktop.MouseEvent) {
	// 鼠标移入 Entry 控件的事件处理
	if !e.bShowMouseMenu {
		return
	}
	logrus.Info("[MouseIn] MouseIn.")
	if e.OnMouseIn != nil {
		e.OnMouseIn(event)
	}
}

func (e *CompletionEntry) MouseMoved(*desktop.MouseEvent) {
	// 鼠标在 Entry 控件上移动的事件处理
	if !e.bShowMouseMenu {
		return
	}
	// logrus.Info("[CompletionEntry] MouseMoved. ")
}

func (e *CompletionEntry) MouseOut() {
	// 鼠标移出 Entry 控件的事件处理
	if !e.bShowMouseMenu {
		return
	}
	if e.OnMouseOut != nil {
		e.OnMouseOut()
	}
}

func (e *CompletionEntry) TappedSecondary(pe *fyne.PointEvent) {

	if e.OnTappedSecondary != nil {
		e.OnTappedSecondary(pe)
	}

	// e.Entry.TappedSecondary(pe)
}

type navigableList struct {
	widget.List
	entry           *widget.Entry
	selected        int
	setTextFromMenu func(string)
	hide            func()
	navigating      bool
	items           []string

	customCreate func() fyne.CanvasObject
	customUpdate func(id widget.ListItemID, object fyne.CanvasObject)
}

func newNavigableList(items []string, entry *widget.Entry, setTextFromMenu func(string), hide func(),
	create func() fyne.CanvasObject, update func(id widget.ListItemID, object fyne.CanvasObject)) *navigableList {
	n := &navigableList{
		entry:           entry,
		selected:        -1,
		setTextFromMenu: setTextFromMenu,
		hide:            hide,
		items:           items,
		customCreate:    create,
		customUpdate:    update,
	}

	n.List = widget.List{
		Length: func() int {
			return len(n.items)
		},
		CreateItem: func() fyne.CanvasObject {
			if fn := n.customCreate; fn != nil {
				return fn()
			}
			return widget.NewLabel("")
		},
		UpdateItem: func(i widget.ListItemID, o fyne.CanvasObject) {
			if fn := n.customUpdate; fn != nil {
				fn(i, o)
				return
			}
			o.(*widget.Label).SetText(n.items[i])
		},
		OnSelected: func(id widget.ListItemID) {
			if !n.navigating && id > -1 {
				setTextFromMenu(n.items[id])
			}
			n.navigating = false
		},
	}
	n.ExtendBaseWidget(n)
	return n
}

// Implements: fyne.Focusable
func (n *navigableList) FocusGained() {
}

// Implements: fyne.Focusable
func (n *navigableList) FocusLost() {
}

func (n *navigableList) SetOptions(items []string) {
	n.Unselect(n.selected)
	n.items = items
	n.Refresh()
	n.selected = -1
}

func (n *navigableList) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyDown:
		if n.selected < len(n.items)-1 {
			n.selected++
		} else {
			n.selected = 0
		}
		n.navigating = true
		n.Select(n.selected)

	case fyne.KeyUp:
		if n.selected > 0 {
			n.selected--
		} else {
			n.selected = len(n.items) - 1
		}
		n.navigating = true
		n.Select(n.selected)
	case fyne.KeyReturn, fyne.KeyEnter:
		if n.selected == -1 { // so the user want to submit the entry
			n.hide()
			n.entry.TypedKey(event)
		} else {
			n.navigating = false
			n.OnSelected(n.selected)
		}
	case fyne.KeyEscape:
		n.hide()
	default:
		n.entry.TypedKey(event)

	}
}

func (n *navigableList) TypedRune(r rune) {
	n.entry.TypedRune(r)
}
