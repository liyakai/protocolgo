package logic

import (
	"io"

	"protocolgo/src/utils"

	"fyne.io/fyne/v2/data/binding"
	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

// 定义页签类型
type ETableType int

const (
	TableType_None ETableType = iota + 1
	TableType_Enum
	TableType_Message
)

type CoreManager struct {
	DocEtree         *etree.Document
	XmlFilePath      string             // 打开的Xml文件路径
	EnumTableList    binding.StringList // enum 数据源
	MessageTableList binding.StringList // message 数据源
}

func (Stapp *CoreManager) Init() {
	// 创建一个列表的数据源
	Stapp.EnumTableList = binding.NewStringList()
	Stapp.MessageTableList = binding.NewStringList()

	Stapp.XmlFilePath = utils.GetWorkRootPath() + "/data/protocolgo.xml"
	Stapp.ReadXmlFromFile(Stapp.XmlFilePath)

	logrus.Info("Init CoreManager done. xml file path:", Stapp.XmlFilePath)
}

func (Stapp *CoreManager) CreateNewXml() {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.DocEtree {
		Stapp.SaveToXmlFile()
	}
	// 创建新的xml
	Stapp.DocEtree = etree.NewDocument()
	Stapp.DocEtree.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	Stapp.DocEtree.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)
	Stapp.SaveToXmlFile()
	logrus.Info("CreateNewXml done.")
}

func (Stapp *CoreManager) ReadXmlFromReader(reader io.Reader) {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.DocEtree {
		Stapp.CloseCurrXmlFile()
	}
	Stapp.DocEtree = etree.NewDocument()
	if _, err := Stapp.DocEtree.ReadFrom(reader); err != nil {
		logrus.Error("ReadXmlFromReader failed. err:", err)
		panic(err)
	}
	Stapp.SyncMessageListWithETree()
	logrus.Info("ReadXmlFromReader done.")
}

func (Stapp *CoreManager) ReadXmlFromFile(filename string) {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.DocEtree {
		Stapp.CloseCurrXmlFile()
	}
	Stapp.DocEtree = etree.NewDocument()
	if err := Stapp.DocEtree.ReadFromFile(filename); err != nil {
		logrus.Error("ReadXmlFromFile failed. err:", err)
		panic(err)
	}
	Stapp.SyncMessageListWithETree()
	logrus.Info("ReadXmlFromFile done.")
}

func (Stapp *CoreManager) SaveToXmlFile() bool {
	if nil == Stapp.DocEtree || Stapp.XmlFilePath == "" {
		logrus.Warn("SaveToXmlFile failed. invalid param. Stapp.XmlFilePath:", Stapp.XmlFilePath)
		return false
	}
	Stapp.DocEtree.Indent(4)
	Stapp.DocEtree.WriteToFile(Stapp.XmlFilePath)
	logrus.Info("SaveToXmlFile done. XmlFilePath:", Stapp.XmlFilePath)
	return true
}

func (Stapp *CoreManager) CloseCurrXmlFile() {
	if nil == Stapp.DocEtree {
		logrus.Info("Need not close the xml. Stapp.DocEtree is nil.")
		return
	}
	Stapp.SaveToXmlFile()
	Stapp.XmlFilePath = ""
	logrus.Info("CloseCurrXmlFile done.")
}

func (Stapp *CoreManager) SetCurrXmlFilePath(currFilePath string) {
	Stapp.CloseCurrXmlFile()
	Stapp.XmlFilePath = currFilePath
	logrus.Info("SetCurrXmlFilePath done.currFilePath:", currFilePath)
}

// 处理 StUnit
func (Stapp *CoreManager) EditUnit(stUnit StUnit) bool {
	if nil == Stapp.DocEtree {
		logrus.Error("EditUnit failed. Stapp.DocEtree is nil, open the xml")
		return false
	}

	var strRoot string
	if stUnit.TableType == TableType_Enum {
		strRoot = "enum"
	} else if stUnit.TableType == TableType_Message {
		strRoot = "message"
	} else {
		strRoot = "other"
	}

	// 先查找是否有枚举的分类
	msg_catagory := Stapp.DocEtree.FindElement(strRoot)
	if msg_catagory == nil {
		msg_catagory = Stapp.DocEtree.CreateElement(strRoot)
	}
	// 在查找枚举中是否有对应的key
	unit := msg_catagory.FindElement(stUnit.UnitName)
	if unit != nil {
		msg_catagory.RemoveChild(msg_catagory.SelectElement(stUnit.UnitName))
	}
	unit = msg_catagory.CreateElement(stUnit.UnitName)
	unit.CreateComment(stUnit.UnitComment)
	for _, row := range stUnit.RowList {
		enum_atom := unit.CreateElement(stUnit.UnitName)
		if row.EntryOption != nil {
			enum_atom.CreateAttr("EntryOption", row.EntryOption.Selected)
		}

		if row.EntryType != nil {
			enum_atom.CreateAttr("EntryType", row.EntryType.Text)
		}
		enum_atom.CreateAttr("EntryName", row.EntryName.Text)
		enum_atom.CreateAttr("EntryIndex", row.EntryIndex.Text)
		enum_atom.CreateAttr("EntryComment", row.EntryComment.Text)
	}
	// Stapp.EnumTableList.Append(editMsg.MsgName)
	Stapp.SyncMessageListWithETree()

	Stapp.SaveToXmlFile()
	logrus.Info("EditUnit done. enumName:", stUnit.UnitName)
	return true
}

func (Stapp *CoreManager) GetEtreeRootName(tableType ETableType) string {
	var strUnitType string
	if tableType == TableType_Enum {
		strUnitType = "enum"
	} else if tableType == TableType_Message {
		strUnitType = "message"
	}
	return strUnitType
}

// 删除 enum/message 列表元素
func (Stapp *CoreManager) DeleteCurrUnit(tableType ETableType, rowName string) bool {

	strUnitName := Stapp.GetEtreeRootName(tableType)

	// 先查找是否有枚举的分类
	catagory := Stapp.DocEtree.FindElement(strUnitName)
	if catagory == nil {
		logrus.Error("DeleteCurrUnit failed. TableType:", tableType, ", strUnitName:", strUnitName, ", rowName:", rowName)
		return false
	}
	// 在查找Unit中是否有对应的key
	unit := catagory.FindElement(rowName)
	if unit == nil {
		logrus.Error("DeleteCurrUnit failed. Can not find target.  TableType:", tableType, ", strUnitName:", strUnitName, ", rowName:", rowName)
		return false
	}
	catagory.RemoveChild(catagory.SelectElement(rowName))

	Stapp.SyncMessageListWithETree()

	logrus.Info("DeleteCurrUnit done. TableType:", tableType, ", strUnitName:", strUnitName, ", rowName:", rowName)
	return false
}

func (Stapp *CoreManager) GetTableListByType(tabletype ETableType) *binding.StringList {
	if tabletype == TableType_Enum {
		return &Stapp.EnumTableList
	} else {
		return &Stapp.MessageTableList
	}
}

func (Stapp *CoreManager) GetLableStingByType(tabletype ETableType) string {
	if tabletype == TableType_Enum {
		return "enum list:"
	} else {
		return "message list:"
	}
}

func (Stapp *CoreManager) GetEditTableTitle(tabletype ETableType) string {
	if tabletype == TableType_Enum {
		return "Edit Enum"
	} else {
		return "Edit Message"
	}
}

func (Stapp *CoreManager) GetEtreeElem(tabletype ETableType, rowName string) *etree.Element {
	strUnitName := Stapp.GetEtreeRootName(tabletype)
	// 先查找是否有枚举的分类
	catagory := Stapp.DocEtree.FindElement(strUnitName)
	if catagory == nil {
		logrus.Error("GetEtreeElem failed. tabletype:", tabletype, ", strUnitName:", strUnitName, ", rowName:", rowName)
		return nil
	}
	// 在查找枚举中是否有对应的key
	unit := catagory.FindElement(rowName)
	if unit == nil {
		logrus.Error("GetEtreeElem failed. Can not find target. tabletype:", tabletype, ", strUnitName:", strUnitName, ", rowName:", rowName)
		return nil
	}
	return unit
}

func (Stapp *CoreManager) SyncMessageListWithETree() bool {
	if nil == Stapp.DocEtree {
		logrus.Error("SyncListWithETree failed. Stapp.DocEtree is nil, open the xml")
		return false
	}

	// 先查找是否有 enum 的分类
	enum_catagory := Stapp.DocEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.DocEtree.CreateElement("enum")
	}
	newEnumListString := []string{}

	// 遍历子元素
	for _, child := range enum_catagory.ChildElements() {
		newEnumListString = append(newEnumListString, child.Tag)
	}
	Stapp.EnumTableList.Set(newEnumListString)

	// 先查找是否有message的分类
	msg_catagory := Stapp.DocEtree.FindElement("message")
	if msg_catagory == nil {
		msg_catagory = Stapp.DocEtree.CreateElement("message")
	}
	newMessageListString := []string{}

	// 遍历子元素
	for _, child := range msg_catagory.ChildElements() {
		newMessageListString = append(newMessageListString, child.Tag)
	}

	Stapp.MessageTableList.Set(newMessageListString)
	logrus.Info("SyncListWithETree done.")
	return true
}

// 检查name 是否重复
func (Stapp *CoreManager) CheckExistSameName(name string) bool {
	// 先查找是否有 enum 的分类
	enum_catagory := Stapp.DocEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.DocEtree.CreateElement("enum")
	}

	// 查找 enum 名字
	enum_uint := enum_catagory.FindElement(name)
	if enum_uint != nil {
		return true
	}

	// 先查找是否有message的分类
	msg_catagory := Stapp.DocEtree.FindElement("message")
	if msg_catagory == nil {
		msg_catagory = Stapp.DocEtree.CreateElement("message")
	}
	// 查找 message 名字
	msg_uint := msg_catagory.FindElement(name)
	return msg_uint != nil
}

func (Stapp *CoreManager) GetAllUseableEntryType() []string {
	result := []string{}
	// 先查找是否有 enum 的分类
	enum_catagory := Stapp.DocEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.DocEtree.CreateElement("enum")
	}

	// 遍历子元素
	for _, child := range enum_catagory.ChildElements() {
		result = append(result, child.Tag)
	}

	// 先查找是否有message的分类
	msg_catagory := Stapp.DocEtree.FindElement("message")
	if msg_catagory == nil {
		msg_catagory = Stapp.DocEtree.CreateElement("message")
	}

	// 遍历子元素
	for _, child := range msg_catagory.ChildElements() {
		result = append(result, child.Tag)
	}
	return result
}

func (Stapp *CoreManager) SyncTableListWithETree(name string) ETableType {
	if Stapp.DocEtree == nil {
		return TableType_None
	}
	if name == "" {
		return TableType_None
	}
	Stapp.SyncMessageListWithETree()

	enum_catagory := Stapp.DocEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.DocEtree.CreateElement("enum")
	}
	// 查找 enum 名字
	enum_uint := enum_catagory.FindElement(name)
	if enum_uint != nil {
		Stapp.EnumTableList.Set([]string{name})
		return TableType_Enum
	}

	msg_catagory := Stapp.DocEtree.FindElement("message")
	if msg_catagory == nil {
		msg_catagory = Stapp.DocEtree.CreateElement("message")
	}
	// 查找 message 名字
	msg_uint := msg_catagory.FindElement(name)
	if msg_uint != nil {
		Stapp.MessageTableList.Set([]string{name})
		return TableType_Message
	}
	return TableType_None

}
