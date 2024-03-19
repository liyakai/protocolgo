package logic

import "fyne.io/fyne/v2/widget"

type EditRowMessage struct {
	EntryIndex      *widget.Entry
	SelectComponent *widget.Select
	EntryKey        *widget.Entry
	EntryValue      *widget.Entry
}

func (editrow *EditRowMessage) RemoveElementFromSlice(s []EditRowMessage, elementToBeDeleted EditRowMessage) []EditRowMessage {
	for i, element := range s {
		// 使用适当的比较来确定哪一个元素应被删除
		if element == elementToBeDeleted {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
