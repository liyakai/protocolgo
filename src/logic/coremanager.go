package logic

import (
	"io"
	"strings"

	"protocolgo/src/utils"

	"fyne.io/fyne/v2/data/binding"
	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

type CoreManager struct {
	DocEtree    *etree.Document
	XmlFilePath string             // 打开的Xml文件路径
	Table1List  binding.StringList // table1 数据源
}

func (Stapp *CoreManager) Init() {
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

// 新增枚举
func (Stapp *CoreManager) AddNewEnum(enumName string) bool {
	if nil == Stapp.DocEtree {
		logrus.Error("AddNewEnum failed. Stapp.DocEtree is nil, open the xml")
		return false
	}
	if strings.Contains(enumName, " ") {
		logrus.Error("AddNewEnum failed. enumName has empty space.")
		return false
	}
	// 先查找是否有枚举的分类
	enum_catagory := Stapp.DocEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.DocEtree.CreateElement("enum")
	}
	// 在查找枚举中是否有对应的key
	enum_unit := enum_catagory.FindElement(enumName)
	if enum_unit != nil {
		logrus.Error("AddNewEnum failed. repeatted enum name. enumName:", enumName)
		return false
	}
	enum_unit = enum_catagory.CreateElement(enumName)
	// elem_enum.CreateComment("Comment")
	// enum_unit.CreateAttr("AttrKey", "AttrValue")
	enum_atom := enum_unit.CreateElement("name")
	enum_atom.CreateAttr("Attr", "attr")
	Stapp.SaveToXmlFile()
	logrus.Info("AddNewEnum done. enumName:", enumName)
	return true
}

// 新增Message
func (Stapp *CoreManager) AddNewMessage(editMsg EditMessage) bool {
	if nil == Stapp.DocEtree {
		logrus.Error("AddNewMessage failed. Stapp.DocEtree is nil, open the xml")
		return false
	}

	// 先查找是否有枚举的分类
	msg_catagory := Stapp.DocEtree.FindElement("message")
	if msg_catagory == nil {
		msg_catagory = Stapp.DocEtree.CreateElement("message")
	}
	// 在查找枚举中是否有对应的key
	enum_unit := msg_catagory.FindElement(editMsg.MsgName)
	if enum_unit != nil {
		logrus.Error("AddNewMessage failed. repeatted enum name. enumName:", editMsg.MsgName)
		return false
	}
	enum_unit = msg_catagory.CreateElement(editMsg.MsgName)
	// elem_enum.CreateComment("Comment")
	// enum_unit.CreateAttr("AttrKey", "AttrValue")
	for _, row := range editMsg.RowList {
		enum_atom := enum_unit.CreateElement(editMsg.MsgName)
		enum_atom.CreateAttr("EntryType", row.EntryType.Selected)
		enum_atom.CreateAttr("EntryKey", row.EntryKey.Text)
		enum_atom.CreateAttr("EntryValue", row.EntryValue.Text)
		enum_atom.CreateAttr("EntryIndex", row.EntryIndex.Text)
	}
	// Stapp.Table1List.Append(editMsg.MsgName)
	Stapp.SyncMessageListWithETree()

	Stapp.SaveToXmlFile()
	logrus.Info("AddNewMessage done. enumName:", editMsg.MsgName)
	return true
}

func (Stapp *CoreManager) SyncMessageListWithETree() bool {
	if nil == Stapp.DocEtree {
		logrus.Error("SyncListWithETree failed. Stapp.DocEtree is nil, open the xml")
		return false
	}

	// 先查找是否有枚举的分类
	msg_catagory := Stapp.DocEtree.FindElement("message")
	if msg_catagory == nil {
		msg_catagory = Stapp.DocEtree.CreateElement("message")
	}
	newListString := []string{}

	// 遍历子元素
	for _, child := range msg_catagory.ChildElements() {
		newListString = append(newListString, child.Tag)
	}
	Stapp.Table1List.Set(newListString)
	logrus.Info("SyncListWithETree done.")
	return true
}
