package logic

import (
	"sort"

	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

// Enum 的一行数据
type EditRowEnum struct {
	EntryName  *widget.Entry
	EntryIndex *widget.Entry
}

// Enum 数据
type EditEnum struct {
	EnumName string
	RowList  []EditRowEnum
}

func (editrow *EditRowEnum) RemoveElementFromSlice(s []EditRowEnum, elementToBeDeleted EditRowEnum) []EditRowEnum {
	for i, element := range s {
		// 使用适当的比较来确定哪一个元素应被删除
		if element == elementToBeDeleted {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// 检查字段名是否有相同的,或者有空的.
func CheckEnumFieldNameList(rowList []EditRowEnum) bool {
	fieldValueNames := make([]string, len(rowList))
	for i, row := range rowList {
		if row.EntryName.Text == "" {
			logrus.Error("Found an empty field name. Index:", row.EntryIndex.Text)
			return false
		}
		fieldValueNames[i] = row.EntryName.Text
	}

	sort.Strings(fieldValueNames)

	for i := 0; i < len(fieldValueNames)-1; i++ {
		if fieldValueNames[i] == fieldValueNames[i+1] {
			logrus.Error("Found duplicate field names. fieldValueNames:", fieldValueNames[i])
			return false
		}
	}

	return true
}

// 检查字段序列号是否有相同的,或者有空的.
func CheckEnumFieldIndexList(rowList []EditRowEnum) bool {
	fieldIndexes := make([]string, len(rowList))
	for i, row := range rowList {
		if row.EntryIndex.Text == "" {
			logrus.Error("Found an empty field index. EntryName:", row.EntryName.Text)
			return false
		}
		fieldIndexes[i] = row.EntryIndex.Text
	}

	sort.Strings(fieldIndexes)

	for i := 0; i < len(fieldIndexes)-1; i++ {
		if fieldIndexes[i] == fieldIndexes[i+1] {
			logrus.Error("Found duplicate field index. fieldIndexes:", fieldIndexes[i])
			return false
		}
	}

	return true
}
