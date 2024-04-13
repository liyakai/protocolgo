package logic

import (
	"bytes"
	"io"
	"strings"

	"protocolgo/src/utils"

	"fyne.io/fyne/v2/data/binding"
	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

// 定义页签类型
type ETableType int

const (
	TableType_None ETableType = iota + 1
	TableType_Main
	TableType_Enum
	TableType_Data
	TableType_Protocol
	TableType_RPC
)

type CoreManager struct {
	FileEtree        *etree.Document     // 保存到文件的 etree
	ChangedEtree     *etree.Document     // 变化的 etree
	ChangedShowEtree *etree.Document     // 包含源数据和变化数据,用于展示的 etree
	XmlFilePath      string              // 打开的Xml文件路径
	Config           *etree.Document     // 配置数据
	MainTableList    binding.StringList  // Main 数据源
	EnumTableList    binding.StringList  // enum 数据源
	DataTableList    binding.StringList  // data 数据源
	PtcTableList     binding.StringList  // ptc 数据源
	RpcTableList     binding.StringList  // rpc 数据源
	SearchMap        map[string]string   // 所有可搜索元素到列表名字的映射
	SearchBuffer     []string            // 所有可所有元素列表
	References       map[string][]string // 字段的依赖列表
}

func (Stapp *CoreManager) Init() {
	// 创建一个列表的数据源
	Stapp.MainTableList = binding.NewStringList()
	Stapp.EnumTableList = binding.NewStringList()
	Stapp.DataTableList = binding.NewStringList()
	Stapp.PtcTableList = binding.NewStringList()
	Stapp.RpcTableList = binding.NewStringList()

	configXmlPath := utils.GetWorkRootPath() + "/data/config.xml"
	Stapp.ReadConfigFromFile(configXmlPath)

	// 读取协议xml文件
	Stapp.XmlFilePath = utils.GetWorkRootPath() + "/data/protocolgo.xml"
	Stapp.ReadXmlFromFile(Stapp.XmlFilePath)

	logrus.Info("Init CoreManager done. xml file path:", Stapp.XmlFilePath)
}

func (Stapp *CoreManager) CreateNewXml() {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.FileEtree {
		Stapp.SaveToXmlFile()
	}
	// 创建新的xml
	Stapp.FileEtree = etree.NewDocument()
	Stapp.FileEtree.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	Stapp.FileEtree.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)
	Stapp.SaveToXmlFile()

	// 同时创建修改的 etree
	Stapp.ChangedEtree = etree.NewDocument()
	Stapp.ChangedShowEtree = Stapp.FileEtree.Copy()
	logrus.Info("CreateNewXml done.")
}

func (Stapp *CoreManager) ReadXmlFromReader(reader io.Reader) {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.FileEtree {
		Stapp.CloseCurrXmlFile()
	}
	Stapp.FileEtree = etree.NewDocument()
	if _, err := Stapp.FileEtree.ReadFrom(reader); err != nil {
		logrus.Error("ReadXmlFromReader failed. err:", err)
		panic(err)
	}
	// 同时创建修改的 etree
	Stapp.ChangedEtree = etree.NewDocument()
	Stapp.ChangedShowEtree = Stapp.FileEtree.Copy()

	Stapp.SyncListWithETree()
	logrus.Info("ReadXmlFromReader done.")
}

func (Stapp *CoreManager) ReadXmlFromFile(filename string) {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.FileEtree {
		Stapp.CloseCurrXmlFile()
	}
	Stapp.FileEtree = etree.NewDocument()
	if err := Stapp.FileEtree.ReadFromFile(filename); err != nil {
		logrus.Error("ReadXmlFromFile failed. err:", err)
		panic(err)
	}

	// 同时创建修改的 etree
	Stapp.ChangedEtree = etree.NewDocument()
	Stapp.ChangedShowEtree = Stapp.FileEtree.Copy()

	Stapp.SyncListWithETree()
	logrus.Info("ReadXmlFromFile done.")
}

func (Stapp *CoreManager) SaveToXmlFile() bool {
	if nil == Stapp.FileEtree || Stapp.XmlFilePath == "" {
		logrus.Warn("SaveToXmlFile failed. invalid param. Stapp.XmlFilePath:", Stapp.XmlFilePath)
		return false
	}
	Stapp.FileEtree.Indent(4)
	Stapp.FileEtree.WriteToFile(Stapp.XmlFilePath)
	logrus.Info("SaveToXmlFile done. XmlFilePath:", Stapp.XmlFilePath)
	return true
}

func (Stapp *CoreManager) CloseCurrXmlFile() {
	if nil == Stapp.FileEtree {
		logrus.Info("Need not close the xml. Stapp.FileEtree is nil.")
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

// 读取配置
func (Stapp *CoreManager) ReadConfigFromFile(filename string) {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	Stapp.Config = etree.NewDocument()
	if err := Stapp.Config.ReadFromFile(filename); err != nil {
		logrus.Error("ReadConfigFromFile failed. err:", err)
		panic(err)
	}

	logrus.Info("ReadConfigFromFile done.")
}

// 读取服务器命名配置
func (Stapp *CoreManager) GetConfigServerName() *etree.Element {
	// 读取Server及其对应的简写
	configElement := Stapp.Config.FindElement("config")
	if configElement == nil {
		logrus.Error("[ReadConfigFromFile] read config failed.")
		return nil
	}
	configServerShortMap := configElement.FindElement("servershort")
	if configServerShortMap == nil {
		logrus.Error("[ReadConfigFromFile] read config failed.")
		return nil
	}
	return configServerShortMap
}

// 读取服务器全名配置
func (Stapp *CoreManager) GetConfigFullServerName() []string {
	configServerName := Stapp.GetConfigServerName()
	if configServerName == nil {
		logrus.Error("[GetConfigFullServerName] Failed for GetConfigServerName failed.")
		return []string{}
	}
	result := []string{}
	for _, cfgServer := range configServerName.ChildElements() {
		cfgFullName := cfgServer.SelectAttr("FullName")
		if cfgFullName != nil && cfgFullName.Value != "" {
			result = append(result, cfgFullName.Value)
		}
	}
	return result
}

// 通过服务器简称获取全称
func (Stapp *CoreManager) GetFullServerName(shortName string, isClientShort bool) string {
	configServerName := Stapp.GetConfigServerName()
	if configServerName == nil {
		logrus.Error("[GetConfigFullServerName] Failed for GetConfigServerName failed.")
		return ""
	}
	strShortFlag := ""
	if isClientShort {
		strShortFlag = "ClientShortName"
	} else {
		strShortFlag = "ServerShortName"
	}
	result := ""
	for _, cfgServer := range configServerName.ChildElements() {
		cfgFullName := cfgServer.SelectAttr(strShortFlag)
		if cfgFullName != nil && cfgFullName.Value != "" && cfgFullName.Value == shortName {
			result = cfgServer.SelectAttr("FullName").Value
			break
		}
	}
	return result
}

// 获取客户端的简称
func (Stapp *CoreManager) GetClientShortName() (bool, string) {
	configServerName := Stapp.GetConfigServerName()
	if configServerName == nil {
		logrus.Error("[GetConfigFullServerName] Failed for GetConfigServerName failed.")
		return false, ""
	}
	for _, cfgServer := range configServerName.ChildElements() {
		cfgIsClient := cfgServer.SelectAttr("IsClient")
		if cfgIsClient != nil && cfgIsClient.Value != "" && strings.ToLower(cfgIsClient.Value) == "true" {
			eleShortName := cfgServer.SelectAttr("ClientShortName")
			if eleShortName == nil {
				logrus.Error("[GetConfigFullServerName] Failed for ClientShortName is lost.")
				return false, ""
			}
			return true, eleShortName.Value
		}
	}
	logrus.Error("[GetConfigFullServerName] Failed for IsClient is lost.")
	return false, ""
}

// 根据协议名字推测服务器全名
func (Stapp *CoreManager) DetectFullNameByProtoName(protoName string) (result bool, firstName string, secondName string) {
	parts := strings.Split(protoName, "_")
	if len(parts) < 2 {
		logrus.Error("[DetectFullNameByProtoName] Failed for invalid protoName:", protoName)
		return false, "", ""
	}
	strNamePre := parts[0]
	if len(strNamePre) <= 0 {
		logrus.Error("[DetectFullNameByProtoName] Failed for invalid protoName:", protoName)
		return false, "", ""
	}
	lenNamePre := len(strNamePre)
	bIsClientProto := false
	result, strClientSHortName := Stapp.GetClientShortName()
	if result == false {
		logrus.Error("[DetectFullNameByProtoName] Failed for GetClientShortName:", protoName)
		return false, "", ""
	}
	if lenNamePre == 2 && (string(strNamePre[0]) == strClientSHortName || string(strNamePre[1]) == strClientSHortName) {
		bIsClientProto = true
	}

	if bIsClientProto {
		if lenNamePre != 2 {
			logrus.Warn("[DetectFullNameByProtoName] strNamePre is not normal. protoName:", protoName, ", strNamePre:", strNamePre)
			return false, "", ""
		}

	} else {
		if lenNamePre != 4 {
			logrus.Warn("[DetectFullNameByProtoName] strNamePre is not normal. protoName:", protoName, ", strNamePre:", strNamePre)
			return false, "", ""
		}
	}
	firstShort := strNamePre[:lenNamePre/2]
	secondShort := strNamePre[lenNamePre/2:]

	return true, Stapp.GetFullServerName(firstShort, bIsClientProto), Stapp.GetFullServerName(secondShort, bIsClientProto)
}

// 根据全名获取协议名前缀
func (Stapp *CoreManager) GetProtoPreName(strSourceFullName string, strTargetFulleName string) (bool, string) {
	configServerName := Stapp.GetConfigServerName()
	if configServerName == nil {
		logrus.Error("[GetProtoPreName] Failed for GetConfigServerName failed.")
		return false, ""
	}
	bIsClient := false
	strSourceShortName := ""
	strTargetShortName := ""
	for _, cfgServer := range configServerName.ChildElements() {
		eleFullName := cfgServer.SelectAttr("FullName")
		cfgIsClient := cfgServer.SelectAttr("IsClient")
		if eleFullName == nil {
			logrus.Error("[GetProtoPreName] Failed for FullName is lost.")
			return false, ""
		}
		if cfgIsClient != nil && (eleFullName.Value == strSourceFullName || eleFullName.Value == strTargetFulleName) && cfgIsClient.Value != "" && strings.ToLower(cfgIsClient.Value) == "true" {
			bIsClient = true
			logrus.Debug("[GetProtoPreName] Found client end.")
			break
		}
	}
	for _, cfgServer := range configServerName.ChildElements() {
		eleFullName := cfgServer.SelectAttr("FullName")
		eleClientShortName := cfgServer.SelectAttr("ClientShortName")
		eleServerShortName := cfgServer.SelectAttr("ServerShortName")
		if eleFullName == nil || eleClientShortName == nil || eleServerShortName == nil {
			logrus.Error("[GetProtoPreName] Failed for FullName is lost.")
			return false, ""
		}
		if strSourceFullName == eleFullName.Value {
			if bIsClient {
				strSourceShortName = eleClientShortName.Value
			} else {
				strSourceShortName = eleServerShortName.Value
			}
		}
		if strTargetFulleName == eleFullName.Value {
			if bIsClient {
				strTargetShortName = eleClientShortName.Value
			} else {
				strTargetShortName = eleServerShortName.Value
			}
		}
	}
	return true, strSourceShortName + strTargetShortName + "_"
}

// 根据服务器全名和协议名字产生/矫正协议名字
func (Stapp *CoreManager) GetProtoNameFromSourceTargetServer(strSourceFullName string, strTargetFulleName string, strProtoName string) string {
	result, strShortName := Stapp.GetProtoPreName(strSourceFullName, strTargetFulleName)
	if !result {
		logrus.Error("[GetProtoNameFromSourceTargetServer] Failed for GetProtoPreName failed.")
		return ""
	}
	if strProtoName == "" {
		return strShortName
	}
	parts := strings.Split(strProtoName, "_")
	if len(parts) < 2 {
		return strShortName
	}
	for i := 1; i < len(parts); i++ {
		strShortName = strShortName + parts[i]
	}
	logrus.Info("[GetProtoNameFromSourceTargetServer] strSourceFullName:", strSourceFullName, ", strTargetFulleName:", strTargetFulleName, ",strProtoName:", strProtoName, ",strShortName:", strShortName, ",parts:", parts)
	return strShortName

}

func (Stapp *CoreManager) SaveProtoXmlToFile() bool {
	if nil == Stapp.FileEtree {
		logrus.Warn("SaveProtoXmlToFile failed. invalid param.")
		return false
	}
	Stapp.FileEtree.Indent(4)
	Stapp.FileEtree.WriteToFile(Stapp.XmlFilePath)
	logrus.Info("SaveProtoXmlToFile done. XmlFilePath:", Stapp.XmlFilePath)
	return true
}

// Add/Update StUnit
func (Stapp *CoreManager) AddUpdateUnit(stUnit StUnit) bool {
	if nil == Stapp.ChangedShowEtree {
		logrus.Error("AddUpdateUnit failed. Stapp.ChangedShowEtree is nil, open the xml")
		return false
	}

	strRoot := Stapp.GetEtreeRootName(stUnit.TableType)

	// 先查找是否有枚举的分类
	msg_catagory := Stapp.ChangedShowEtree.FindElement(strRoot)
	if msg_catagory == nil {
		msg_catagory = Stapp.ChangedShowEtree.CreateElement(strRoot)
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
		if row.EntryDefault != nil {
			enum_atom.CreateAttr("EntryDefault", row.EntryDefault.Selected)
		}
		enum_atom.CreateAttr("EntryComment", row.EntryComment.Text)
	}
	// Stapp.EnumTableList.Append(editMsg.MsgName)
	Stapp.SyncListWithETree()

	Stapp.SaveToXmlFile()
	logrus.Info("AddUpdateUnit from stUnit done. enumName:", stUnit.UnitName)
	return true
}

// Revert StUnit
func (Stapp *CoreManager) RevertUnitFromChanged(eTableType ETableType, strUnitName string) bool {
	if nil == Stapp.ChangedShowEtree || nil == Stapp.ChangedEtree {
		logrus.Error("RevertUnitFromChanged failed. Stapp.ChangedShowEtree is nil, open the xml")
		return false
	}

	strRoot := Stapp.GetEtreeRootName(eTableType)

	// 先在 ChangedShow 中找是否有对应的分类
	show_catagory := Stapp.ChangedShowEtree.FindElement(strRoot)
	if show_catagory == nil {
		show_catagory = Stapp.ChangedShowEtree.CreateElement(strRoot)
	}
	// 在 ChangedShow 查找是否有对应的key
	show_unit := show_catagory.FindElement(strUnitName)
	if show_unit != nil {
		show_catagory.RemoveChild(show_catagory.SelectElement(strUnitName))
	}

	// 在变化项中定位元素
	// 先查找是否有枚举的分类
	changed_catagory := Stapp.ChangedEtree.FindElement(strRoot)
	if changed_catagory == nil {
		logrus.Error("RevertUnitFromChanged failed. strRoot:", strRoot, " is not exist.eTableType:", eTableType, ",strUnitName:", strUnitName)
		return false
	}
	// 在查找枚举中是否有对应的key
	changed_unit := changed_catagory.FindElement(strUnitName)
	if changed_unit == nil {
		logrus.Error("RevertUnitFromChanged failed. strUnitName:", strUnitName, " is not exist. eTableType:", eTableType, ",strUnitName:", strUnitName)
		return false
	}

	operationAttr := changed_unit.SelectAttr("opertype")
	if operationAttr == nil {
		logrus.Error("RevertUnitFromChanged failed. opertype is not exist. eTableType:", eTableType, ",strUnitName:", strUnitName)
		return false
	}
	strOperType := operationAttr.Value

	if strOperType == "delete" || strOperType == "update" {
		show_catagory.AddChild(changed_unit)
	} else if strOperType == "add" {

	} else {
		logrus.Error("RevertUnitFromChanged failed. opertype is invalid. eTableType:", eTableType, ",strUnitName:", strUnitName, ",strOperType:", strOperType)
		return false
	}

	Stapp.SyncListWithETree()

	logrus.Info("RevertUnitFromChanged done. eTableType:", eTableType, ",strUnitName:", strUnitName)
	return true
}

func (Stapp *CoreManager) GetEtreeRootName(tableType ETableType) string {
	var strUnitType string
	if tableType == TableType_Enum {
		strUnitType = "enum"
	} else if tableType == TableType_Data {
		strUnitType = "data"
	} else if tableType == TableType_Protocol {
		strUnitType = "protocol"
	} else if tableType == TableType_RPC {
		strUnitType = "rpc"
	}
	return strUnitType
}

// 删除 enum/message 列表元素
func (Stapp *CoreManager) DeleteCurrUnit(tableType ETableType, rowName string) bool {

	strUnitName := Stapp.GetEtreeRootName(tableType)

	// 先查找是否有枚举的分类
	catagory := Stapp.ChangedShowEtree.FindElement(strUnitName)
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

	Stapp.SyncListWithETree()

	Stapp.SaveProtoXmlToFile()

	logrus.Info("DeleteCurrUnit done. TableType:", tableType, ", strUnitName:", strUnitName, ", rowName:", rowName)
	return false
}

func (Stapp *CoreManager) GetTableListByType(tabletype ETableType) *binding.StringList {
	if tabletype == TableType_Main {
		return &Stapp.MainTableList
	} else if tabletype == TableType_Enum {
		return &Stapp.EnumTableList
	} else if tabletype == TableType_Data {
		return &Stapp.DataTableList
	} else if tabletype == TableType_Protocol {
		return &Stapp.PtcTableList
	} else if tabletype == TableType_RPC {
		return &Stapp.RpcTableList
	} else {
		return &Stapp.RpcTableList
	}
}

func (Stapp *CoreManager) GetLableStingByType(tabletype ETableType) string {
	if tabletype == TableType_Main {
		return "changed list:"
	} else if tabletype == TableType_Enum {
		return "enum list:"
	} else if tabletype == TableType_Data {
		return "data list:"
	} else if tabletype == TableType_Protocol {
		return "protocol list:"
	} else if tabletype == TableType_RPC {
		return "rpc list:"
	} else {
		return "changed list:"
	}
}

func (Stapp *CoreManager) GetEditTableTitle(tabletype ETableType, name string) string {
	if tabletype == TableType_Main {
		return "Edit Changed:" + name
	} else if tabletype == TableType_Enum {
		return "Edit Enum:" + name
	} else if tabletype == TableType_Data {
		return "Edit Data:" + name
	} else if tabletype == TableType_Protocol {
		return "Edit Protocol:" + name
	} else if tabletype == TableType_RPC {
		return "Edit Rpc:" + name
	} else {
		return "Edit Message:" + name
	}
}

func (Stapp *CoreManager) GetEtreeElem(tabletype ETableType, rowName string) *etree.Element {
	strUnitName := Stapp.GetEtreeRootName(tabletype)
	// 先查找是否有枚举的分类
	catagory := Stapp.ChangedShowEtree.FindElement(strUnitName)
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

func (Stapp *CoreManager) SyncListWithETree() bool {
	if nil == Stapp.ChangedShowEtree {
		logrus.Error("SyncListWithETree failed. Stapp.ChangedShowEtree is nil, open the xml")
		return false
	}
	Stapp.SearchMap = map[string]string{}
	Stapp.References = map[string][]string{}
	Stapp.SyncListWithETreeCatagoryEnum()
	Stapp.SyncListWithETreeCatagoryData()
	Stapp.SyncListWithETreeCatagoryProtocol()
	Stapp.SyncListWithETreeCatagoryRpc()

	Stapp.GetChangedEtree()
	// 将变化数据同步到 main 页签
	Stapp.SyncMainListWithChangedEtree()

	Stapp.SearchBuffer = []string{}
	for key, _ := range Stapp.SearchMap {
		Stapp.SearchBuffer = append(Stapp.SearchBuffer, key)
		// logrus.Debug("SyncListWithETree key:", key)
	}

	logrus.Info("SyncListWithETree done.")
	return true
}

func (Stapp *CoreManager) SyncListWithETreeCatagoryEnum() {
	// 先查找是否有 enum 的分类
	enum_catagory := Stapp.ChangedShowEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.ChangedShowEtree.CreateElement("enum")
	}
	newEnumListString := []string{}

	// 遍历子元素
	for _, EnumClass := range enum_catagory.ChildElements() {
		newEnumListString = append(newEnumListString, EnumClass.Tag)
		// 枚举类名字映射
		Stapp.SearchMap[EnumClass.Tag] = EnumClass.Tag
		Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToLower(EnumClass.Tag)] = EnumClass.Tag
		Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToUpper(EnumClass.Tag)] = EnumClass.Tag
		// 将注释映射
		for _, child := range EnumClass.Child {
			// 检查该子元素是否为注释
			if comment, ok := child.(*etree.Comment); ok {
				if comment.Data != "" {
					Stapp.SearchMap["["+EnumClass.Tag+"]"+comment.Data] = EnumClass.Tag
					Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToLower(comment.Data)] = EnumClass.Tag
					Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToUpper(comment.Data)] = EnumClass.Tag
				}

				break
			}
		}
		for _, EnumClassConent := range EnumClass.ChildElements() {
			entryName := EnumClassConent.SelectAttr("EntryName")
			if entryName != nil && entryName.Value != "" {
				Stapp.SearchMap["["+EnumClass.Tag+"]"+entryName.Value] = EnumClass.Tag
				Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToLower(entryName.Value)] = EnumClass.Tag
				Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToUpper(entryName.Value)] = EnumClass.Tag
				// logrus.Debug("SyncListWithETree entryName.Value:", entryName.Value)
			}
			entryComment := EnumClassConent.SelectAttr("EntryComment")
			if entryComment != nil && entryComment.Value != "" {
				Stapp.SearchMap["["+EnumClass.Tag+"]"+entryComment.Value] = EnumClass.Tag
				Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToLower(entryComment.Value)] = EnumClass.Tag
				Stapp.SearchMap["["+EnumClass.Tag+"]"+strings.ToUpper(entryComment.Value)] = EnumClass.Tag
				// logrus.Debug("SyncListWithETree entryComment.Value:", entryComment.Value)
			}
		}

	}
	Stapp.EnumTableList.Set(newEnumListString)
}

func (Stapp *CoreManager) SyncListWithETreeCatagoryData() {
	// 先查找是否有 protocol 的分类
	data_catagory := Stapp.ChangedShowEtree.FindElement("data")
	if data_catagory == nil {
		data_catagory = Stapp.ChangedShowEtree.CreateElement("data")
	}
	newDataListString := []string{}

	// 遍历子元素
	for _, DataClass := range data_catagory.ChildElements() {
		newDataListString = append(newDataListString, DataClass.Tag)
		// 枚举类名字映射
		Stapp.SearchMap[DataClass.Tag] = DataClass.Tag
		Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToLower(DataClass.Tag)] = DataClass.Tag
		Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToUpper(DataClass.Tag)] = DataClass.Tag
		// 将注释映射
		for _, child := range DataClass.Child {
			// 检查该子元素是否为注释
			if comment, ok := child.(*etree.Comment); ok {
				if comment.Data != "" {
					Stapp.SearchMap["["+DataClass.Tag+"]"+comment.Data] = DataClass.Tag
					Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToLower(comment.Data)] = DataClass.Tag
					Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToUpper(comment.Data)] = DataClass.Tag
				}
				break
			}
		}
		for _, PtcClassConent := range DataClass.ChildElements() {
			entryName := PtcClassConent.SelectAttr("EntryName")
			if entryName != nil && entryName.Value != "" {
				Stapp.SearchMap["["+DataClass.Tag+"]"+entryName.Value] = DataClass.Tag
				Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToLower(entryName.Value)] = DataClass.Tag
				Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToUpper(entryName.Value)] = DataClass.Tag
				// logrus.Debug("SyncListWithETree entryName.Value:", entryName.Value)
			}
			entryComment := PtcClassConent.SelectAttr("EntryComment")
			if entryComment != nil && entryComment.Value != "" {
				Stapp.SearchMap["["+DataClass.Tag+"]"+entryComment.Value] = DataClass.Tag
				Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToLower(entryComment.Value)] = DataClass.Tag
				Stapp.SearchMap["["+DataClass.Tag+"]"+strings.ToUpper(entryComment.Value)] = DataClass.Tag
				// logrus.Debug("SyncListWithETree entryComment.Value:", entryComment.Value)
			}
			// 记录依赖
			entryType := PtcClassConent.SelectAttr("EntryType")
			if entryType != nil && entryType.Value != "" && !Stapp.CheckProtoType(entryType.Value) {
				Stapp.References[entryType.Value] = append(Stapp.References[entryType.Value], DataClass.Tag)
				logrus.Debug("SyncListWithETree init References. Value:", entryType.Value, ", MsgClass.Tag:", DataClass.Tag, ",--->Stapp.References:", Stapp.References)
			}
		}
	}
	Stapp.DataTableList.Set(newDataListString)
}
func (Stapp *CoreManager) SyncListWithETreeCatagoryProtocol() {
	// 先查找是否有 data 的分类
	ptc_catagory := Stapp.ChangedShowEtree.FindElement("protocol")
	if ptc_catagory == nil {
		ptc_catagory = Stapp.ChangedShowEtree.CreateElement("protocol")
	}
	newPtcListString := []string{}

	// 遍历子元素
	for _, PtcClass := range ptc_catagory.ChildElements() {
		newPtcListString = append(newPtcListString, PtcClass.Tag)
		// 枚举类名字映射
		Stapp.SearchMap[PtcClass.Tag] = PtcClass.Tag
		Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToLower(PtcClass.Tag)] = PtcClass.Tag
		Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToUpper(PtcClass.Tag)] = PtcClass.Tag
		// 将注释映射
		for _, child := range PtcClass.Child {
			// 检查该子元素是否为注释
			if comment, ok := child.(*etree.Comment); ok {
				if comment.Data != "" {
					Stapp.SearchMap["["+PtcClass.Tag+"]"+comment.Data] = PtcClass.Tag
					Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToLower(comment.Data)] = PtcClass.Tag
					Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToUpper(comment.Data)] = PtcClass.Tag
				}
				break
			}
		}
		for _, PtcClassConent := range PtcClass.ChildElements() {
			entryName := PtcClassConent.SelectAttr("EntryName")
			if entryName != nil && entryName.Value != "" {
				Stapp.SearchMap["["+PtcClass.Tag+"]"+entryName.Value] = PtcClass.Tag
				Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToLower(entryName.Value)] = PtcClass.Tag
				Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToUpper(entryName.Value)] = PtcClass.Tag
				// logrus.Debug("SyncListWithETree entryName.Value:", entryName.Value)
			}
			entryComment := PtcClassConent.SelectAttr("EntryComment")
			if entryComment != nil && entryComment.Value != "" {
				Stapp.SearchMap["["+PtcClass.Tag+"]"+entryComment.Value] = PtcClass.Tag
				Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToLower(entryComment.Value)] = PtcClass.Tag
				Stapp.SearchMap["["+PtcClass.Tag+"]"+strings.ToUpper(entryComment.Value)] = PtcClass.Tag
				// logrus.Debug("SyncListWithETree entryComment.Value:", entryComment.Value)
			}
			// 记录依赖
			entryType := PtcClassConent.SelectAttr("EntryType")
			if entryType != nil && entryType.Value != "" && !Stapp.CheckProtoType(entryType.Value) {
				Stapp.References[entryType.Value] = append(Stapp.References[entryType.Value], PtcClass.Tag)
				// logrus.Debug("SyncListWithETree init References. Value:", entryType.Value, ", MsgClass.Tag:", MsgClass.Tag, ",--->Stapp.References:", Stapp.References)
			}
		}
	}
	Stapp.PtcTableList.Set(newPtcListString)
}
func (Stapp *CoreManager) SyncListWithETreeCatagoryRpc() {
}
func (Stapp *CoreManager) SyncMainListWithChangedEtree() {

	if Stapp.ChangedEtree == nil {
		logrus.Error("SyncMainListWithChangedEtree failed. ChangedEtree is nil.")
		return
	}

	newChangedListString := []string{}

	// 遍历子元素
	for _, cataClass := range Stapp.ChangedEtree.ChildElements() {
		for _, diffClass := range cataClass.ChildElements() {
			newChangedListString = append(newChangedListString, "["+diffClass.SelectAttr("opertype").Value+"]"+diffClass.Tag)
			logrus.Debug("SyncMainListWithChangedEtree. newChangedListString:", newChangedListString)
		}
	}
	logrus.Debug("SyncMainListWithChangedEtree done. newChangedListString:", newChangedListString)
	Stapp.MainTableList.Set(newChangedListString)
}

// 检查name 是否重复
func (Stapp *CoreManager) CheckExistSameName(name string) bool {
	// 先查找是否有 enum 的分类
	enum_catagory := Stapp.ChangedShowEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.ChangedShowEtree.CreateElement("enum")
	}

	// 查找 enum 名字
	enum_uint := enum_catagory.FindElement(name)
	if enum_uint != nil {
		return true
	}

	// 先查找是否有message的分类
	msg_catagory := Stapp.ChangedShowEtree.FindElement("message")
	if msg_catagory == nil {
		msg_catagory = Stapp.ChangedShowEtree.CreateElement("message")
	}
	// 查找 message 名字
	msg_uint := msg_catagory.FindElement(name)
	return msg_uint != nil
}

func (Stapp *CoreManager) GetAllSearchName() []string {
	// for _, name := range Stapp.SearchBuffer {
	// 	logrus.Debug("GetAllSearchName SearchBuffer name:", name)
	// }
	return Stapp.SearchBuffer
}

func (Stapp *CoreManager) GetListNameBySearchName(searchname string) string {
	return Stapp.SearchMap[searchname]
}

func (Stapp *CoreManager) GetAllUseableEntryType() []string {
	result := []string{}
	// 先查找是否有 enum 的分类
	enum_catagory := Stapp.ChangedShowEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.ChangedShowEtree.CreateElement("enum")
	}

	// 遍历子元素
	for _, child := range enum_catagory.ChildElements() {
		result = append(result, child.Tag)
	}

	// 先查找是否有 data 的分类
	msg_catagory := Stapp.ChangedShowEtree.FindElement("data")
	if msg_catagory == nil {
		msg_catagory = Stapp.ChangedShowEtree.CreateElement("data")
	}

	// 遍历子元素
	for _, child := range msg_catagory.ChildElements() {
		result = append(result, child.Tag)
	}

	// 先查找是否有 data 的分类
	ptc_catagory := Stapp.ChangedShowEtree.FindElement("protocol")
	if ptc_catagory == nil {
		ptc_catagory = Stapp.ChangedShowEtree.CreateElement("protocol")
	}

	// 遍历子元素
	for _, child := range ptc_catagory.ChildElements() {
		result = append(result, child.Tag)
	}

	return result
}

func (coremgr *CoreManager) GetProtoType() []string {
	return []string{"int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "float", "double", "bool", "string", "bytes"}
}

func (coremgr *CoreManager) CheckProtoType(str string) bool {
	for _, v := range coremgr.GetProtoType() {
		if v == str {
			return true
		}
	}
	return false
}

func (coremgr *CoreManager) GetAllUseableEntryTypeWithProtoType() []string {
	result := coremgr.GetAllUseableEntryType()
	result = append(result, coremgr.GetProtoType()...)
	return result
}

// 获取References
func (coremgr *CoreManager) GetReferences(str string) []string {
	return coremgr.References[str]
}

func (Stapp *CoreManager) SearchTableListWithName(name string) ETableType {
	if Stapp.ChangedShowEtree == nil {
		return TableType_None
	}
	if name == "" {
		return TableType_None
	}
	// Stapp.SyncListWithETree()

	enum_catagory := Stapp.ChangedShowEtree.FindElement("enum")
	if enum_catagory == nil {
		enum_catagory = Stapp.ChangedShowEtree.CreateElement("enum")
	}
	// 查找 enum 名字
	enum_uint := enum_catagory.FindElement(name)
	if enum_uint != nil {
		return TableType_Enum
	}

	changed_enum_catagory := Stapp.ChangedEtree.FindElement("enum")
	if changed_enum_catagory != nil {
		// 查找 enum 名字
		changed_enum_uint := changed_enum_catagory.FindElement(name)
		if changed_enum_uint != nil {
			return TableType_Enum
		}
	}

	data_catagory := Stapp.ChangedShowEtree.FindElement("data")
	if data_catagory == nil {
		data_catagory = Stapp.ChangedShowEtree.CreateElement("data")
	}
	// 查找 data 名字
	data_uint := data_catagory.FindElement(name)
	if data_uint != nil {
		return TableType_Data
	}
	changed_data_catagory := Stapp.ChangedEtree.FindElement("data")
	if changed_data_catagory != nil {
		// 查找 enum 名字
		changed_data_uint := changed_data_catagory.FindElement(name)
		if changed_data_uint != nil {
			return TableType_Data
		}
	}

	msg_catagory := Stapp.ChangedShowEtree.FindElement("protocol")
	if msg_catagory == nil {
		msg_catagory = Stapp.ChangedShowEtree.CreateElement("protocol")
	}

	// 查找 message 名字
	msg_uint := msg_catagory.FindElement(name)
	if msg_uint != nil {
		return TableType_Protocol
	}

	changed_msg_catagory := Stapp.ChangedEtree.FindElement("protocol")
	if changed_msg_catagory != nil {
		// 查找 enum 名字
		changed_enum_uint := changed_msg_catagory.FindElement(name)
		if changed_enum_uint != nil {
			return TableType_Protocol
		}
	}

	rpc_catagory := Stapp.ChangedShowEtree.FindElement("rpc")
	if rpc_catagory == nil {
		rpc_catagory = Stapp.ChangedShowEtree.CreateElement("rpc")
	}
	// 查找 message 名字
	rpc_uint := rpc_catagory.FindElement(name)
	if rpc_uint != nil {
		return TableType_RPC
	}

	changed_rpc_catagory := Stapp.ChangedEtree.FindElement("rpc")
	if changed_rpc_catagory != nil {
		// 查找 enum 名字
		changed_rpc_uint := changed_rpc_catagory.FindElement(name)
		if changed_rpc_uint != nil {
			return TableType_RPC
		}
	}

	return TableType_None
}

// 根据枚举类型获取枚举名字
func (coremgr *CoreManager) GetVarListOfEnum(strEnumName string) []string {
	enumElement := coremgr.GetEtreeElem(TableType_Enum, strEnumName)
	if enumElement == nil {
		logrus.Error("[CoreManager] failed for GetEtreeElem, strEnumName:", strEnumName)
		return []string{}
	}
	result := []string{}
	for _, cfgEnum := range enumElement.ChildElements() {
		enumVar := cfgEnum.SelectAttr("EntryName")
		if enumVar != nil && enumVar.Value != "" {
			result = append(result, enumVar.Value)
		}
	}
	return result
}

// 根据 FileEtree 和 ChangedShowEtree 算出差异和差异类型
func (coremgr *CoreManager) GetChangedEtree() {
	if coremgr.FileEtree == nil || coremgr.ChangedEtree == nil || coremgr.ChangedShowEtree == nil {
		logrus.Error("[CoreManager] GetChangedEtree failed. invalid etree.")
	}
	coremgr.ChangedEtree = etree.NewDocument()
	coremgr.GetEtreeDiff("enum", "delete", coremgr.FileEtree, coremgr.ChangedShowEtree, coremgr.ChangedEtree)
	coremgr.GetEtreeDiff("enum", "add", coremgr.ChangedShowEtree, coremgr.FileEtree, coremgr.ChangedEtree)

	coremgr.GetEtreeDiff("data", "delete", coremgr.FileEtree, coremgr.ChangedShowEtree, coremgr.ChangedEtree)
	coremgr.GetEtreeDiff("data", "add", coremgr.ChangedShowEtree, coremgr.FileEtree, coremgr.ChangedEtree)

	coremgr.GetEtreeDiff("protocol", "delete", coremgr.FileEtree, coremgr.ChangedShowEtree, coremgr.ChangedEtree)
	coremgr.GetEtreeDiff("protocol", "add", coremgr.ChangedShowEtree, coremgr.FileEtree, coremgr.ChangedEtree)

	coremgr.GetEtreeDiff("rpc", "delete", coremgr.FileEtree, coremgr.ChangedShowEtree, coremgr.ChangedEtree)
	coremgr.GetEtreeDiff("rpc", "add", coremgr.ChangedShowEtree, coremgr.FileEtree, coremgr.ChangedEtree)

	changedBuffer := new(bytes.Buffer)
	coremgr.ChangedEtree.WriteTo(changedBuffer)
	logrus.Info("[CoreManager] GetChangedEtree done. coremgr.ChangedEtree:", changedBuffer.String())
}

// 寻找两个 etree 之间的 差集 eTreeA - eTreeB
func (coremgr *CoreManager) GetEtreeDiff(strTagName string, strOperType string, eTreeA *etree.Document, eTreeB *etree.Document, eTreeDiff *etree.Document) bool {
	if eTreeA == nil || eTreeB == nil || eTreeDiff == nil {
		logrus.Error("[CoreManager] GetEtreeDiff failed. invalid etree.")
		return false
	}

	// 对比 enum
	cataA := eTreeA.FindElement(strTagName)
	if cataA == nil {
		cataA = eTreeA.CreateElement(strTagName)
	}
	cataB := eTreeB.FindElement(strTagName)
	if cataB == nil {
		cataB = eTreeA.CreateElement(strTagName)
	}
	cataDiff := eTreeDiff.FindElement(strTagName)

	for _, elemA := range cataA.ChildElements() {
		elemB := cataB.FindElement(elemA.Tag)
		if elemB == nil {
			if cataDiff == nil {
				cataDiff = eTreeDiff.CreateElement(strTagName)
			}
			diffElem := elemA.Copy()
			diffElem.CreateAttr("opertype", strOperType)
			cataDiff.AddChild(diffElem)
			logrus.Debug("[CoreManager] GetEtreeDiff found. strTagName:", strTagName, ",strOperType:", strOperType)
		} else {
			// 双方都存在,那就检查双方是否一致.
			// 先检查是否已经在diff中.
			if cataDiff != nil {
				enumElemDiff := cataDiff.FindElement(elemA.Tag)
				if enumElemDiff != nil {
					continue
				}
			}
			if !coremgr.CheckSameUnit(elemA, elemB) {
				if cataDiff == nil {
					cataDiff = eTreeDiff.CreateElement(strTagName)
				}
				diffElem := elemA.Copy()
				diffElem.CreateAttr("opertype", "update")
				cataDiff.AddChild(diffElem)
			}

		}
	}
	return true
}

// 检查两个单元是否完全一致
func (coremgr *CoreManager) CheckSameUnit(unitA *etree.Element, unitB *etree.Element) bool {
	if unitA == nil && unitB == nil {
		return true
	}
	if unitA == nil || unitB == nil {
		return false
	}

	// 检查名字是否相同
	if unitA.Tag != unitB.Tag {
		return false
	}
	// 检查元素的个数是否相同
	if len(unitA.Child) != len(unitB.Child) {
		return false
	}
	// 检查注释是否一致
	unitAComment := ""
	for _, child := range unitA.Child {
		// 检查该子元素是否为注释
		if comment, ok := child.(*etree.Comment); ok {
			unitAComment = comment.Data
			break
		}
	}
	unitBComment := ""
	for _, child := range unitB.Child {
		// 检查该子元素是否为注释
		if comment, ok := child.(*etree.Comment); ok {
			unitBComment = comment.Data
			break
		}
	}
	if unitAComment != unitBComment {
		return false
	}
	indexChild := 0
	for _, childA := range unitA.ChildElements() {
		childB := unitB.ChildElements()[indexChild]
		indexChild = indexChild + 1
		if len(childA.Attr) != len(childB.Attr) {
			return false
		}
		indexAttr := 0
		// 细致检查每个元素是否相同
		for _, attrA := range childA.Attr {
			attrB := childB.Attr[indexAttr]
			indexAttr = indexAttr + 1
			if attrA.Value != attrB.Value {
				return false
			}
		}
	}

	return true
}
