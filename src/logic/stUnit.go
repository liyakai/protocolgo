package logic

import (
	"os"
	"regexp"
	"sort"

	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
)

// Unit 的一行数据
type StRowUnit struct {
	EntryIndex  *widget.Entry
	EntryOption *widget.Select
	// EntryType   *widget.Entry
	EntryType    *CompletionEntry
	EntryName    *widget.Entry
	EntryDefault *widget.Select
	EntryComment *widget.Entry
}

type StStrRowUnit struct {
	EntryIndex   string
	EntryOption  string
	EntryType    string
	EntryName    string
	EntryDefault string
	EntryComment string
}

// Unit 集合
type StUnits struct {
	UnitListName string
	UnitList     []StUnit
}

// Unit 数据
type StUnit struct {
	UnitName     string
	UnitComment  string
	TableType    ETableType
	SubTableType ESubTableType
	RowList      []StRowUnit
	IsCreatNew   bool
}

// Unit container
type StUnitContainer struct {
	UnitNameEntry    *widget.Entry
	UnitCommentEntry *widget.Entry
	TableType        ETableType
	RowList          *[]StRowUnit
	IsCreatNew       bool
}

func (editrow *StRowUnit) RemoveElementFromSlice(s []StRowUnit, elementToBeDeleted StRowUnit) []StRowUnit {
	for i, element := range s {
		// 使用适当的比较来确定哪一个元素应被删除
		if element == elementToBeDeleted {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// 检查字段名是否有相同的,或者有空的.
func CheckFieldNameList(rowList []StRowUnit) bool {
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
func CheckFieldIndexList(rowList []StRowUnit) bool {
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

func (stUnitContainer *StUnitContainer) GetStUnit() StUnit {
	var stUnit StUnit
	stUnit.UnitName = stUnitContainer.UnitNameEntry.Text
	stUnit.UnitComment = stUnitContainer.UnitCommentEntry.Text
	stUnit.TableType = stUnitContainer.TableType
	stUnit.RowList = *stUnitContainer.RowList
	stUnit.IsCreatNew = stUnitContainer.IsCreatNew
	return stUnit
}

// 检查路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

// 验证IP地址是否合法的函数
func IsValidIP(ip string) bool {
	// IPv4地址的正则表达式
	ipv4Pattern := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`

	// 匹配IP地址
	match, err := regexp.MatchString(ipv4Pattern, ip)
	if err != nil {
		return false // 正则表达式错误，视为非法
	}

	return match
}

// 验证端口是否合法的函数
func IsValidPort(port string) bool {
	// 端口号的正则表达式，包括1-5位数字，范围在1-65535之间
	portPattern := `^([1-9]|[1-9]\d{1,4}|[1-5]\d{4}|6([0-4]\d{3}|5([0-4]\d{2}|5([0-2]\d|3[0-5]))))$`

	// 匹配端口号
	match, err := regexp.MatchString(portPattern, port)
	if err != nil {
		return false // 正则表达式错误，视为非法
	}

	return match
}
