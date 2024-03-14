package logic

import (
	"github.com/beevik/etree"
)

type CoreManager struct {
	DocEtree    *etree.Document
	XmlFilePath string // 打开的Xml文件路径
}

func (Stapp *CoreManager) CreateNewXml() {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.DocEtree {
		Stapp.SaveToFile()
	}
	// 创建新的xml
	DocEtree := etree.NewDocument()
	DocEtree.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	DocEtree.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)

}

func (Stapp *CoreManager) ReadXmlFromFile() {
	// 如果现在打开的xml不为空,则先保存现在打开的xml
	if nil != Stapp.DocEtree {
		Stapp.SaveToFile()
	}
	doc := etree.NewDocument()
	if err := doc.ReadFromFile("bookstore.xml"); err != nil {
		panic(err)
	}
}

func (Stapp *CoreManager) SaveToFile() {
	if nil == Stapp.DocEtree {
		return
	}
	Stapp.SaveToFile()
}
